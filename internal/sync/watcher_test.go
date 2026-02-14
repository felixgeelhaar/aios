package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestPollingWatcherDetectsChangeAndRepairs(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, "client")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(cfgDir, "config.json")
	if err := os.WriteFile(file, []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine()
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := w.Watch(ctx, engine, []string{cfgDir}, func(_ string) error {
		// Simulate an auto-repair writing normalized content.
		return os.WriteFile(file, []byte(`{"v":2}`), 0o644)
	})
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}

	if err := os.WriteFile(file, []byte(`{"v":broken}`), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-events:
	case <-time.After(1 * time.Second):
		t.Fatal("expected drift event")
	}

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if engine.CurrentState() == "clean" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected clean state after repair, got %s", engine.CurrentState())
}

// AC1: File watchers must observe all configured client config directories.
func TestPollingWatcherObservesMultiplePaths(t *testing.T) {
	root := t.TempDir()
	dirs := make([]string, 3)
	files := make([]string, 3)
	for i := range dirs {
		dirs[i] = filepath.Join(root, fmt.Sprintf("client-%d", i))
		if err := os.MkdirAll(dirs[i], 0o755); err != nil {
			t.Fatal(err)
		}
		files[i] = filepath.Join(dirs[i], "config.json")
		if err := os.WriteFile(files[i], []byte(`{"v":1}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	engine := NewEngine()
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repairCount int32
	events, err := w.Watch(ctx, engine, dirs, func(_ string) error {
		atomic.AddInt32(&repairCount, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}

	// Mutate file in the third directory to prove all paths are watched.
	if err := os.WriteFile(files[2], []byte(`{"v":changed}`), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case ev := <-events:
		if ev.Path != dirs[2] {
			t.Fatalf("expected event for %s, got %s", dirs[2], ev.Path)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("expected drift event from third directory")
	}
}

// AC1: Watcher rejects empty path list.
func TestPollingWatcherRejectsEmptyPaths(t *testing.T) {
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx := context.Background()
	_, err := w.Watch(ctx, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

// AC2: Must detect manual edits within polling interval.
func TestPollingWatcherDetectsWithinInterval(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "config.json")
	if err := os.WriteFile(file, []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	interval := 20 * time.Millisecond
	w := NewPollingWatcher(interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := w.Watch(ctx, nil, []string{root}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Mutate and measure how long detection takes.
	if err := os.WriteFile(file, []byte(`{"v":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	select {
	case <-events:
		elapsed := time.Since(start)
		// Should detect within 3x the interval (allowing scheduling jitter).
		if elapsed > 3*interval {
			t.Fatalf("detection took %v, expected within %v", elapsed, 3*interval)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("expected drift event within interval")
	}
}

// AC4: Must auto-repair safe config mismatches without user intervention.
func TestPollingWatcherAutoRepairCalledAutomatically(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "config.json")
	if err := os.WriteFile(file, []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine()
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repaired int32
	_, err := w.Watch(ctx, engine, []string{root}, func(_ string) error {
		atomic.AddInt32(&repaired, 1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Simulate a manual edit.
	if err := os.WriteFile(file, []byte(`{"v":manual-edit}`), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&repaired) > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected repair function to be called automatically")
}

// AC5: Must surface non-repairable drift as alerts (engine stays drifted on repair failure).
func TestPollingWatcherRepairFailureKeepsDrifted(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "config.json")
	if err := os.WriteFile(file, []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine()
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := w.Watch(ctx, engine, []string{root}, func(_ string) error {
		return fmt.Errorf("repair failed: permission denied")
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file, []byte(`{"v":broken}`), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-events:
	case <-time.After(1 * time.Second):
		t.Fatal("expected drift event")
	}

	// Wait for repair attempt to complete.
	time.Sleep(50 * time.Millisecond)

	if engine.CurrentState() != "drifted" {
		t.Fatalf("expected engine to remain drifted after failed repair, got %s", engine.CurrentState())
	}
}

// AC6: Must maintain parity across all synced clients after repair.
func TestPollingWatcherAllClientsCleanAfterRepair(t *testing.T) {
	root := t.TempDir()
	dirs := make([]string, 3)
	files := make([]string, 3)
	for i := range dirs {
		dirs[i] = filepath.Join(root, fmt.Sprintf("client-%d", i))
		if err := os.MkdirAll(dirs[i], 0o755); err != nil {
			t.Fatal(err)
		}
		files[i] = filepath.Join(dirs[i], "config.json")
		if err := os.WriteFile(files[i], []byte(`{"v":1}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	engine := NewEngine()
	w := NewPollingWatcher(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := w.Watch(ctx, engine, dirs, func(_ string) error {
		// Successful repair.
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Mutate one client.
	if err := os.WriteFile(files[0], []byte(`{"v":changed}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for detection and repair cycle.
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if engine.CurrentState() == "clean" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected clean state after repair across all clients, got %s", engine.CurrentState())
}
