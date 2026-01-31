package config

import (
	"fmt"
	"testing"
	"time"
)

type watchResult struct {
	old     string
	new     string
	current string
}

func TestConfigWatchersReceiveUpdates(t *testing.T) {
	cm := NewConfigManager(nil)
	cm.data["foo"] = "old"

	resultCh := make(chan watchResult, 1)
	globalCh := make(chan map[string]any, 1)

	if err := cm.Watch("foo", func(old, new any) {
		resultCh <- watchResult{
			old:     fmt.Sprint(old),
			new:     fmt.Sprint(new),
			current: cm.Get("foo"),
		}
	}); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	if err := cm.WatchGlobal(func(_, new any) {
		if data, ok := new.(map[string]any); ok {
			globalCh <- data
			return
		}
		globalCh <- map[string]any{}
	}); err != nil {
		t.Fatalf("WatchGlobal failed: %v", err)
	}

	cm.handleSourceUpdate("test", map[string]any{"foo": "new"})

	deadline := time.After(200 * time.Millisecond)
	var gotResult bool
	var gotGlobal bool

	for !gotResult || !gotGlobal {
		select {
		case res := <-resultCh:
			gotResult = true
			if res.old != "old" || res.new != "new" || res.current != "new" {
				t.Fatalf("unexpected watch result: %+v", res)
			}
		case data := <-globalCh:
			gotGlobal = true
			if data["foo"] != "new" {
				t.Fatalf("unexpected global data: %v", data["foo"])
			}
		case <-deadline:
			t.Fatal("watchers did not fire in time")
		}
	}
}

func TestConfigWatchersNormalizeKeys(t *testing.T) {
	cm := NewConfigManager(nil)
	cm.data["foo_bar"] = "old"

	resultCh := make(chan watchResult, 1)
	if err := cm.Watch("FooBar", func(old, new any) {
		resultCh <- watchResult{
			old:     fmt.Sprint(old),
			new:     fmt.Sprint(new),
			current: cm.Get("foo_bar"),
		}
	}); err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	cm.handleSourceUpdate("test", map[string]any{"FOO_BAR": "new"})

	select {
	case res := <-resultCh:
		if res.old != "old" || res.new != "new" || res.current != "new" {
			t.Fatalf("unexpected watch result: %+v", res)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watchers did not fire in time")
	}
}
