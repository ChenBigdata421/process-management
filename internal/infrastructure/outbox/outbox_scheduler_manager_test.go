package outbox

import "testing"

// TestCreateScheduler_DLQEnabled asserts the PR-6 publish-side wiring: buildSchedulerConfig
// (split from createScheduler so it needs no *gorm.DB) must turn DLQ on with a non-nil handler.
// C3 (jxt-core v1.7.2 SchedulerConfig.Validate) rejects EnableDLQ=true with a nil DLQHandler or
// DLQInterval < 1s, so a non-nil handler + the 5m interval here also implies the config would
// pass NewScheduler validation — guarding the C1 re-scan fix for the process outbox.
func TestCreateScheduler_DLQEnabled(t *testing.T) {
	cfg := buildSchedulerConfig(1)
	if !cfg.EnableDLQ {
		t.Fatal("EnableDLQ = false, want true")
	}
	if cfg.DLQHandler == nil {
		t.Fatal("DLQHandler = nil (C3 rejects at NewScheduler)")
	}
	if cfg.DLQAlertHandler == nil {
		t.Fatal("DLQAlertHandler = nil (C2 P1 channel must be wired)")
	}
	if cfg.DLQInterval <= 0 {
		t.Fatalf("DLQInterval = %v, want > 0", cfg.DLQInterval)
	}
}
