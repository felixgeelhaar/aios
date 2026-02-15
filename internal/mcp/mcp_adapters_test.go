package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/governance"
	"github.com/felixgeelhaar/aios/internal/observability"
	"github.com/felixgeelhaar/aios/internal/runtime"
)

func TestMcpAuditBundleStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "bundle.json")
	store := mcpAuditBundleStore{}

	bundle := governance.AuditBundle{
		GeneratedAt: "2025-01-01T00:00:00Z",
		Signature:   "sig",
		Records:     []governance.AuditRecord{{Category: "test", Decision: "allow"}},
	}
	if err := store.WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := store.LoadBundle(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.GeneratedAt != bundle.GeneratedAt {
		t.Errorf("generated_at mismatch: got %q, want %q", loaded.GeneratedAt, bundle.GeneratedAt)
	}
	if len(loaded.Records) != 1 || loaded.Records[0].Category != "test" {
		t.Error("records mismatch")
	}
}

func TestMcpAuditBundleStore_WriteBundleCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "bundle.json")
	store := mcpAuditBundleStore{}

	if err := store.WriteBundle(path, governance.AuditBundle{GeneratedAt: "now"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestMcpAuditBundleStore_LoadBundleInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := mcpAuditBundleStore{}
	_, err := store.LoadBundle(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestMcpAuditBundleStore_LoadBundleNotFound(t *testing.T) {
	store := mcpAuditBundleStore{}
	_, err := store.LoadBundle("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestMcpSnapshotStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	store := mcpSnapshotStore{}

	snap := observability.NewSnapshot(map[string]float64{"cpu": 42.0})
	if err := store.Append(path, snap); err != nil {
		t.Fatalf("append: %v", err)
	}
	loaded, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(loaded))
	}
	if loaded[0].Metrics["cpu"] != 42.0 {
		t.Error("metrics mismatch")
	}
}

func TestMcpSnapshotStore_AppendMultiple(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	store := mcpSnapshotStore{}

	if err := store.Append(path, observability.NewSnapshot(map[string]float64{"a": 1})); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{"b": 2})); err != nil {
		t.Fatalf("second append: %v", err)
	}
	loaded, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(loaded))
	}
}

func TestMcpSnapshotStore_LoadAllNonexistent(t *testing.T) {
	store := mcpSnapshotStore{}
	loaded, err := store.LoadAll("/nonexistent/history.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty list, got %d", len(loaded))
	}
}

func TestMcpSnapshotStore_LoadAllEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	store := mcpSnapshotStore{}
	loaded, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty list, got %d", len(loaded))
	}
}

func TestMcpSnapshotStore_LoadAllInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := mcpSnapshotStore{}
	_, err := store.LoadAll(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestMcpSnapshotStore_AppendCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "history.json")
	store := mcpSnapshotStore{}
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{"x": 1})); err != nil {
		t.Fatalf("append: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestMcpExecutionReportStore_WriteReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "report.json")
	store := mcpExecutionReportStore{}

	plan := runtime.ExecutionPlan{SkillID: "test-skill", Version: "1.0.0"}
	report := runtime.BuildExecutionReport(plan, "success")
	if err := store.WriteReport(path, report); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var loaded runtime.ExecutionReport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if loaded.SkillID != "test-skill" {
		t.Errorf("expected skill_id %q, got %q", "test-skill", loaded.SkillID)
	}
}

func TestMcpExecutionReportStore_WriteReportCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "report.json")
	store := mcpExecutionReportStore{}

	plan := runtime.ExecutionPlan{SkillID: "s", Version: "1.0.0"}
	report := runtime.BuildExecutionReport(plan, "ok")
	if err := store.WriteReport(path, report); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}
