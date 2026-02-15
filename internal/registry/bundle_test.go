package registry

import (
	"testing"
)

func TestBundleCreate(t *testing.T) {
	br := NewBundleRegistry()
	bundle := Bundle{
		ID:          "team-sdk",
		Name:        "Team SDK",
		Description: "Core skills for team development",
		Version:     "1.0.0",
		Skills: []BundleSkill{
			{ID: "ddd-expert", Version: "1.0.0"},
			{ID: "code-review", Version: "1.0.0"},
		},
	}
	if err := br.Create(bundle); err != nil {
		t.Fatal(err)
	}

	b, ok := br.Get("team-sdk")
	if !ok {
		t.Fatal("bundle not found")
	}
	if b.Name != "Team SDK" {
		t.Errorf("expected name 'Team SDK', got %q", b.Name)
	}
	if len(b.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(b.Skills))
	}
}

func TestBundleList(t *testing.T) {
	br := NewBundleRegistry()
	br.Create(Bundle{ID: "bundle-1", Name: "Bundle 1", Skills: []BundleSkill{{ID: "s1"}}})
	br.Create(Bundle{ID: "bundle-2", Name: "Bundle 2", Skills: []BundleSkill{{ID: "s2"}}})

	list := br.List()
	if len(list) != 2 {
		t.Errorf("expected 2 bundles, got %d", len(list))
	}
}

func TestBundleDelete(t *testing.T) {
	br := NewBundleRegistry()
	br.Create(Bundle{ID: "to-delete", Name: "To Delete", Skills: []BundleSkill{{ID: "s1"}}})

	if err := br.Delete("to-delete"); err != nil {
		t.Fatal(err)
	}

	if _, ok := br.Get("to-delete"); ok {
		t.Error("bundle should have been deleted")
	}
}

func TestBundleAddSkill(t *testing.T) {
	br := NewBundleRegistry()
	br.Create(Bundle{ID: "bundle", Name: "Bundle", Skills: []BundleSkill{{ID: "skill-1", Version: "1.0.0"}}})

	if err := br.AddSkill("bundle", BundleSkill{ID: "skill-2", Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}

	b, _ := br.Get("bundle")
	if len(b.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(b.Skills))
	}
}

func TestBundleValidation(t *testing.T) {
	br := NewBundleRegistry()

	err := br.Create(Bundle{ID: "", Name: "Test", Skills: []BundleSkill{{ID: "s1"}}})
	if err == nil {
		t.Error("expected error for empty id")
	}

	err = br.Create(Bundle{ID: "test", Name: "", Skills: []BundleSkill{{ID: "s1"}}})
	if err == nil {
		t.Error("expected error for empty name")
	}

	err = br.Create(Bundle{ID: "test", Name: "Test", Skills: []BundleSkill{}})
	if err == nil {
		t.Error("expected error for empty skills")
	}
}
