package service

import (
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
)

func testItemService(t *testing.T) *ItemService {
	t.Helper()
	db := testutil.NewTestDB(t, &model.Item{}, &model.User{})
	repo := repository.NewItemRepository(db)
	return NewItemService(repo)
}

func TestItemService_Create(t *testing.T) {
	svc := testItemService(t)

	resp, err := svc.Create(&dto.CreateItemRequest{
		Title:       "Test Item",
		Description: "A description",
	}, 1)

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if resp.Title != "Test Item" {
		t.Fatalf("Create() title = %q, want %q", resp.Title, "Test Item")
	}
	if resp.Status != "active" {
		t.Fatalf("Create() status = %q, want %q", resp.Status, "active")
	}
	if resp.UserID != 1 {
		t.Fatalf("Create() user_id = %d, want 1", resp.UserID)
	}
}

func TestItemService_CreateWithStatus(t *testing.T) {
	svc := testItemService(t)

	resp, err := svc.Create(&dto.CreateItemRequest{
		Title:  "Inactive Item",
		Status: "inactive",
	}, 1)

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if resp.Status != "inactive" {
		t.Fatalf("Create() status = %q, want %q", resp.Status, "inactive")
	}
}

func TestItemService_GetByID(t *testing.T) {
	svc := testItemService(t)

	created, err := svc.Create(&dto.CreateItemRequest{Title: "Find Me"}, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resp, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if resp.Title != "Find Me" {
		t.Fatalf("GetByID() title = %q, want %q", resp.Title, "Find Me")
	}
}

func TestItemService_GetByID_NotFound(t *testing.T) {
	svc := testItemService(t)

	_, err := svc.GetByID(9999)
	if err == nil {
		t.Fatal("GetByID() expected error for non-existent ID")
	}
}

func TestItemService_List(t *testing.T) {
	svc := testItemService(t)

	for _, title := range []string{"A", "B"} {
		if _, err := svc.Create(&dto.CreateItemRequest{Title: title}, 1); err != nil {
			t.Fatalf("Create(%q) error = %v", title, err)
		}
	}

	items, total, err := svc.List(0, 10, "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("List() total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Fatalf("List() len = %d, want 2", len(items))
	}
}

func TestItemService_Update(t *testing.T) {
	svc := testItemService(t)

	created, err := svc.Create(&dto.CreateItemRequest{Title: "Original"}, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resp, err := svc.Update(created.ID, &dto.UpdateItemRequest{Title: "Updated"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if resp.Title != "Updated" {
		t.Fatalf("Update() title = %q, want %q", resp.Title, "Updated")
	}
}

func TestItemService_Update_NotFound(t *testing.T) {
	svc := testItemService(t)

	_, err := svc.Update(9999, &dto.UpdateItemRequest{Title: "nope"})
	if err == nil {
		t.Fatal("Update() expected error for non-existent ID")
	}
}

func TestItemService_Delete(t *testing.T) {
	svc := testItemService(t)

	created, err := svc.Create(&dto.CreateItemRequest{Title: "To Delete"}, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.GetByID(created.ID)
	if err == nil {
		t.Fatal("GetByID() after Delete should return error")
	}
}

func TestItemService_Delete_NotFound(t *testing.T) {
	svc := testItemService(t)

	err := svc.Delete(9999)
	if err == nil {
		t.Fatal("Delete() expected error for non-existent ID")
	}
}
