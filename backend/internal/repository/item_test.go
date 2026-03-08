package repository

import (
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
)

func testItemRepo(t *testing.T) (*ItemRepository, func()) {
	t.Helper()
	db := testutil.NewTestDB(t, &model.Item{}, &model.User{})
	return NewItemRepository(db), func() {}
}

func TestItemRepository_CreateAndFindByID(t *testing.T) {
	repo, _ := testItemRepo(t)

	item := &model.Item{Title: "Test Item", Description: "desc", Status: "active", UserID: 1}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID == 0 {
		t.Fatal("Create() did not set ID")
	}

	found, err := repo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Title != "Test Item" {
		t.Fatalf("FindByID() title = %q, want %q", found.Title, "Test Item")
	}
}

func TestItemRepository_FindByID_NotFound(t *testing.T) {
	repo, _ := testItemRepo(t)

	_, err := repo.FindByID(9999)
	if err == nil {
		t.Fatal("FindByID() expected error for non-existent ID")
	}
}

func TestItemRepository_List(t *testing.T) {
	repo, _ := testItemRepo(t)

	for i, title := range []string{"Alpha", "Beta", "Gamma"} {
		item := &model.Item{Title: title, Status: "active", UserID: uint(i + 1)}
		if err := repo.Create(item); err != nil {
			t.Fatalf("Create(%q) error = %v", title, err)
		}
	}

	items, total, err := repo.List(0, 10, "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("List() total = %d, want 3", total)
	}
	if len(items) != 3 {
		t.Fatalf("List() len = %d, want 3", len(items))
	}
}

func TestItemRepository_ListWithKeyword(t *testing.T) {
	repo, _ := testItemRepo(t)

	for _, item := range []*model.Item{
		{Title: "Go Programming", Status: "active", UserID: 1},
		{Title: "Rust Programming", Status: "active", UserID: 1},
		{Title: "Other", Status: "active", UserID: 1},
	} {
		if err := repo.Create(item); err != nil {
			t.Fatalf("Create(%q) error = %v", item.Title, err)
		}
	}

	items, total, err := repo.List(0, 10, "Programming", "")
	if err != nil {
		t.Fatalf("List(keyword) error = %v", err)
	}
	if total != 2 {
		t.Fatalf("List(keyword) total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Fatalf("List(keyword) len = %d, want 2", len(items))
	}
}

func TestItemRepository_ListWithStatus(t *testing.T) {
	repo, _ := testItemRepo(t)

	for _, item := range []*model.Item{
		{Title: "Active", Status: "active", UserID: 1},
		{Title: "Inactive", Status: "inactive", UserID: 1},
	} {
		if err := repo.Create(item); err != nil {
			t.Fatalf("Create(%q) error = %v", item.Title, err)
		}
	}

	items, total, err := repo.List(0, 10, "", "active")
	if err != nil {
		t.Fatalf("List(status) error = %v", err)
	}
	if total != 1 {
		t.Fatalf("List(status) total = %d, want 1", total)
	}
	if items[0].Title != "Active" {
		t.Fatalf("List(status) title = %q, want %q", items[0].Title, "Active")
	}
}

func TestItemRepository_Update(t *testing.T) {
	repo, _ := testItemRepo(t)

	item := &model.Item{Title: "Original", Status: "active", UserID: 1}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	item.Title = "Updated"
	if err := repo.Update(item); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	found, _ := repo.FindByID(item.ID)
	if found.Title != "Updated" {
		t.Fatalf("after Update() title = %q, want %q", found.Title, "Updated")
	}
}

func TestItemRepository_SoftDelete(t *testing.T) {
	repo, _ := testItemRepo(t)

	item := &model.Item{Title: "ToDelete", Status: "active", UserID: 1}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.SoftDelete(item.ID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}

	_, err := repo.FindByID(item.ID)
	if err == nil {
		t.Fatal("FindByID() after SoftDelete should return error")
	}
}
