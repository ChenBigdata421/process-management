//go:build system

package outbox_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// outbox.NewNoOpEventPublisher (jxt-core event_publisher.go) satisfies outbox.EventPublisher so
// NewScheduler does not panic; processDLQ never publishes (it only does the DLQ steps).
//
// D-PROC-2: process emits NO Prometheus counter (no metrics package exists in this service), so
// this test only spies Handle invocation counts — there is no counter assertion (unlike the
// evidence/security/file-storage variants).

// TestOutboxDLQ_MaxRetryToDeadLetteredToNotified proves the §11 acceptance for process: a max_retry
// row (MySQL) terminates in dead_lettered exactly once, is notified, the handler fires once, and the
// next DLQInterval does NOT re-scan it (C1 fix).
//
// This is a RELEASE-GATE test: it requires a real MySQL with the outbox_events table migrated to
// v1.7.1+ (adds dead_lettered_at / dlq_notified_at). The process MySQL is NOT exposed on a host port
// in this dev env, so PROCESS_MYSQL_DSN is unset and this test SKIPs cleanly per-commit; it runs
// against a real DB pre-merge in the release gate (Task 8.5).
//
// Run with:
//
//	PROCESS_MYSQL_DSN="root:root@tcp(localhost:3306)/process?charset=utf8mb4&parseTime=True&multiStatements=true" \
//	  go test -tags=system ./internal/infrastructure/outbox/ -run TestOutboxDLQ -v
func TestOutboxDLQ_MaxRetryToDeadLetteredToNotified(t *testing.T) {
	dsn := os.Getenv("PROCESS_MYSQL_DSN") // e.g. root:root@tcp(localhost:3306)/process?...&multiStatements=true
	if dsn == "" {
		t.Skip("PROCESS_MYSQL_DSN not set; needs MySQL with outbox_events migrated (multiStatements=true)")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	if err := db.AutoMigrate(&gormadapter.OutboxEventModel{}); err != nil { // ensures columns + index post-bump
		t.Fatalf("automigrate: %v", err)
	}

	// Insert a max_retry row simulating exhausted publish retries. OV#2: sentinel tenant 999999.
	id := "dlq-ev-" + time.Now().Format("150405.000000")
	created := time.Now().UTC()
	if err := db.Exec(`INSERT INTO outbox_events
		(id, tenant_id, aggregate_id, aggregate_type, event_type, payload, status, retry_count, max_retries, last_error, created_at, updated_at, version)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, 999999, "agg-1", "Process", "ProcessCompleted", []byte("{}"), "max_retry", 3, 3,
		"simulated publisher exhaustion", created, created, 1).Error; err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM outbox_events WHERE id = ?", id) })

	spy := &countingHandler{}
	repo := gormadapter.NewGormOutboxRepository(db)
	// NB: jxt-core v1.7.2 C3 Validate enforces DLQInterval >= 1s (scheduler.go:227). The brief's
	// 300ms template predates that floor; bumped to 1s here. Sleeps widened accordingly so ≥1 full
	// DLQInterval elapses before each assertion (first tick CAS+notify, second tick proves no rescan).
	cfg := &outbox.SchedulerConfig{
		PollInterval: 1 * time.Second, BatchSize: 100, TenantID: 999999, // OV#2: sentinel test tenant — 0 means "all tenants" and would CAS-transition unrelated max_retry rows in a shared/dev DB
		EnableRetry: false, EnableCleanup: false, EnableHealthCheck: false, EnableMetrics: false,
		EnableDLQ: true, DLQInterval: 1 * time.Second, // jxt-core C3 floor (≥1s)
		DLQHandler:      outbox.DLQHandlerFunc(spy.Handle),
		DLQAlertHandler: outbox.DLQAlertHandlerFunc(func(context.Context, *outbox.OutboxEvent) error { return nil }),
		ShutdownTimeout: 2 * time.Second,
	}
	s := outbox.NewScheduler(
		outbox.WithRepository(repo),
		outbox.WithEventPublisher(outbox.NewNoOpEventPublisher()),
		outbox.WithTopicMapper(outbox.DefaultTopicMapper),
		outbox.WithSchedulerConfig(cfg),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(2500 * time.Millisecond) // > 1 DLQInterval → step1 (CAS to dead_lettered) + step2 (notify) run; 2nd tick proves no rescan

	// Assert terminal + notified. MySQL: CAST(... AS CHAR) for the datetime(3) columns.
	var status, dlAt, notified sql.NullString
	if err := db.Raw("SELECT status, CAST(dead_lettered_at AS CHAR), CAST(dlq_notified_at AS CHAR) FROM outbox_events WHERE id = ?", id).Row().Scan(&status, &dlAt, &notified); err != nil {
		t.Fatalf("read: %v", err)
	}
	if status.String != "dead_lettered" {
		t.Fatalf("status = %q, want dead_lettered", status.String)
	}
	if !dlAt.Valid || dlAt.String == "" {
		t.Fatal("dead_lettered_at not set")
	}
	if !notified.Valid || notified.String == "" {
		t.Fatal("dlq_notified_at not set (notification step did not complete)")
	}
	if spy.calls != 1 {
		t.Fatalf("Handle calls = %d after first tick, want 1", spy.calls)
	}

	// §11 "no re-scan": a subsequent DLQInterval must NOT re-Handle (dlq_notified_at already set).
	spy.calls = 0
	time.Sleep(1500 * time.Millisecond) // > 1 DLQInterval
	if spy.calls != 0 {
		t.Fatalf("Handle re-called %d time(s) after notification; C1 rescan regression", spy.calls)
	}
	_ = s.Stop(ctx)
}

type countingHandler struct{ calls int }

func (c *countingHandler) Handle(_ context.Context, _ *outbox.OutboxEvent) error { c.calls++; return nil }
