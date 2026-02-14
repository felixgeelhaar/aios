package projectinventory

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
)

type fakeRepo struct {
	inv domain.Inventory
	err error
}

func (f *fakeRepo) Load(context.Context) (domain.Inventory, error) {
	return f.inv, f.err
}

func (f *fakeRepo) Save(_ context.Context, inventory domain.Inventory) error {
	f.inv = inventory
	return f.err
}

type fakeCanonicalizer struct {
	err error
}

func (f fakeCanonicalizer) Canonicalize(path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "/abs/" + path, nil
}

func TestTrackListInspectUntrack(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	svc.now = func() time.Time { return time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC) }

	p, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("track failed: %v", err)
	}
	if p.Path != "/abs/repo1" {
		t.Fatalf("unexpected path: %q", p.Path)
	}

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("unexpected list length: %d", len(list))
	}

	got, err := svc.Inspect(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if got.ID != p.ID {
		t.Fatalf("unexpected inspect id: %q", got.ID)
	}

	if err := svc.Untrack(context.Background(), p.ID); err != nil {
		t.Fatalf("untrack failed: %v", err)
	}
	list, err = svc.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}

func TestTrackRequiresPath(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	_, err := svc.Track(context.Background(), "")
	if !errors.Is(err, domain.ErrProjectPathRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInspectNotFound(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	_, err := svc.Inspect(context.Background(), "missing")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrackDuplicate_ReturnsExisting(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

	p1, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("first track: %v", err)
	}

	p2, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("second track: %v", err)
	}
	if p1.ID != p2.ID {
		t.Errorf("duplicate track should return same project")
	}
}

func TestTrackCanonicalizeError(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{err: errors.New("bad path")})

	_, err := svc.Track(context.Background(), "repo1")
	if err == nil {
		t.Fatal("expected error from canonicalizer")
	}
}

func TestTrackRepoLoadError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("load fail")}
	svc := NewService(repo, fakeCanonicalizer{})

	_, err := svc.Track(context.Background(), "repo1")
	if err == nil {
		t.Fatal("expected error from repo load")
	}
}

func TestListRepoError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("load fail")}
	svc := NewService(repo, fakeCanonicalizer{})

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error from repo load")
	}
}

func TestUntrackEmptySelector(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})

	err := svc.Untrack(context.Background(), "")
	if !errors.Is(err, domain.ErrProjectSelectorRequired) {
		t.Fatalf("expected ErrProjectSelectorRequired, got: %v", err)
	}
}

func TestUntrackNotFound(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})

	err := svc.Untrack(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got: %v", err)
	}
}

func TestUntrackRepoLoadError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("load fail")}
	svc := NewService(repo, fakeCanonicalizer{})

	err := svc.Untrack(context.Background(), "something")
	if err == nil {
		t.Fatal("expected error from repo load")
	}
}

func TestUntrackByPath(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

	p, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("track: %v", err)
	}

	// Untrack by path instead of ID.
	err = svc.Untrack(context.Background(), p.Path)
	if err != nil {
		t.Fatalf("untrack by path: %v", err)
	}
}

func TestUntrackByCanonicalizedPath(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

	_, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("track: %v", err)
	}

	// Untrack using a path that doesn't directly match but canonicalizes to the tracked path.
	err = svc.Untrack(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("untrack by canonicalized path: %v", err)
	}
}

func TestInspectEmptySelector(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})

	_, err := svc.Inspect(context.Background(), "")
	if !errors.Is(err, domain.ErrProjectSelectorRequired) {
		t.Fatalf("expected ErrProjectSelectorRequired, got: %v", err)
	}
}

func TestInspectRepoLoadError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("load fail")}
	svc := NewService(repo, fakeCanonicalizer{})

	_, err := svc.Inspect(context.Background(), "something")
	if err == nil {
		t.Fatal("expected error from repo load")
	}
}

func TestInspectByCanonicalizedPath(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

	p, err := svc.Track(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("track: %v", err)
	}

	// Inspect using a path that canonicalizes to the tracked path.
	got, err := svc.Inspect(context.Background(), "repo1")
	if err != nil {
		t.Fatalf("inspect by canonicalized path: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %q, got %q", p.ID, got.ID)
	}
}

func TestNewService_SetsNowFunc(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeCanonicalizer{})
	// Just verify it doesn't panic and returns a reasonable time.
	now := svc.now()
	if now.IsZero() {
		t.Error("now() returned zero time")
	}
}
