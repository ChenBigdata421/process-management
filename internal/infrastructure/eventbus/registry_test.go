package eventbus

import (
	"testing"
)

func TestRegistry_Topics(t *testing.T) {
	r := NewRegistry()
	r.Register("process.task.events")

	topics := r.Topics()
	if len(topics) != 1 || topics[0] != "process.task.events" {
		t.Fatalf("Topics = %v, want [process.task.events]", topics)
	}
}

func TestRegistry_Dedup(t *testing.T) {
	r := NewRegistry()
	r.Register("process.task.events")
	r.Register("process.task.events")

	topics := r.Topics()
	if len(topics) != 1 {
		t.Fatalf("Topics 去重失败: %v", topics)
	}
}
