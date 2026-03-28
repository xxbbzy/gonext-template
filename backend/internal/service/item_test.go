package service

import (
	"net/http"
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
)

func testItemService(t *testing.T) *ItemService {
	t.Helper()
	db := testutil.NewTestDB(t, &model.Item{}, &model.User{})
	repo := repository.NewItemRepository(db)
	return NewItemService(repo)
}

func TestItemService_Create(t *testing.T) {
	testCases := []struct {
		name       string
		req        dto.CreateItemRequest
		userID     uint
		wantStatus string
	}{
		{
			name: "create with default status",
			req: dto.CreateItemRequest{
				Title:       "Test Item",
				Description: "A description",
			},
			userID:     1,
			wantStatus: "active",
		},
		{
			name: "create with explicit status",
			req: dto.CreateItemRequest{
				Title:       "Inactive Item",
				Description: "",
				Status:      "inactive",
			},
			userID:     7,
			wantStatus: "inactive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := testItemService(t)

			resp, err := svc.Create(&tc.req, tc.userID)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Create() response is nil")
			}
			if resp.Title != tc.req.Title {
				t.Fatalf("title = %q, want %q", resp.Title, tc.req.Title)
			}
			if resp.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", resp.Status, tc.wantStatus)
			}
			if resp.UserID != tc.userID {
				t.Fatalf("user_id = %d, want %d", resp.UserID, tc.userID)
			}
			if resp.CreatedAt == "" || resp.UpdatedAt == "" {
				t.Fatal("timestamps should not be empty")
			}
		})
	}
}

func TestItemService_GetByID(t *testing.T) {
	testCases := []struct {
		name       string
		seed       bool
		queryID    uint
		wantErr    bool
		wantCode   int
		wantStatus int
		wantTitle  string
	}{
		{
			name:      "get existing item",
			seed:      true,
			wantTitle: "Find Me",
		},
		{
			name:       "get non-existent item",
			seed:       false,
			queryID:    9999,
			wantErr:    true,
			wantCode:   errcode.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := testItemService(t)
			queryID := tc.queryID

			if tc.seed {
				created, err := svc.Create(&dto.CreateItemRequest{Title: "Find Me"}, 1)
				if err != nil {
					t.Fatalf("Create() error = %v", err)
				}
				queryID = created.ID
			}

			resp, err := svc.GetByID(queryID)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("GetByID() error = %v", err)
			}
			if resp.Title != tc.wantTitle {
				t.Fatalf("title = %q, want %q", resp.Title, tc.wantTitle)
			}
		})
	}
}

func TestItemService_Update(t *testing.T) {
	testCases := []struct {
		name          string
		seed          bool
		updateID      uint
		req           dto.UpdateItemRequest
		wantErr       bool
		wantCode      int
		wantStatus    int
		wantTitle     string
		wantDesc      string
		wantItemState string
	}{
		{
			name:          "update existing item",
			seed:          true,
			req:           dto.UpdateItemRequest{Title: "Updated", Description: "Updated desc", Status: "inactive"},
			wantTitle:     "Updated",
			wantDesc:      "Updated desc",
			wantItemState: "inactive",
		},
		{
			name:       "update non-existent item",
			seed:       false,
			updateID:   9999,
			req:        dto.UpdateItemRequest{Title: "nope"},
			wantErr:    true,
			wantCode:   errcode.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := testItemService(t)
			updateID := tc.updateID

			if tc.seed {
				created, err := svc.Create(&dto.CreateItemRequest{Title: "Original", Description: "Original desc"}, 1)
				if err != nil {
					t.Fatalf("Create() error = %v", err)
				}
				updateID = created.ID
			}

			resp, err := svc.Update(updateID, &tc.req)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if resp.Title != tc.wantTitle {
				t.Fatalf("title = %q, want %q", resp.Title, tc.wantTitle)
			}
			if resp.Description != tc.wantDesc {
				t.Fatalf("description = %q, want %q", resp.Description, tc.wantDesc)
			}
			if resp.Status != tc.wantItemState {
				t.Fatalf("status = %q, want %q", resp.Status, tc.wantItemState)
			}
		})
	}
}

func TestItemService_Delete(t *testing.T) {
	testCases := []struct {
		name       string
		seed       bool
		deleteID   uint
		wantErr    bool
		wantCode   int
		wantStatus int
	}{
		{
			name: "delete existing item",
			seed: true,
		},
		{
			name:       "delete non-existent item",
			seed:       false,
			deleteID:   9999,
			wantErr:    true,
			wantCode:   errcode.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := testItemService(t)
			deleteID := tc.deleteID

			if tc.seed {
				created, err := svc.Create(&dto.CreateItemRequest{Title: "To Delete"}, 1)
				if err != nil {
					t.Fatalf("Create() error = %v", err)
				}
				deleteID = created.ID
			}

			err := svc.Delete(deleteID)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			_, err = svc.GetByID(deleteID)
			assertAppError(t, err, errcode.ErrNotFound, http.StatusNotFound)
		})
	}
}

func TestItemService_List(t *testing.T) {
	testCases := []struct {
		name      string
		seedItems []dto.CreateItemRequest
		offset    int
		limit     int
		keyword   string
		status    string
		wantTotal int64
		wantLen   int
	}{
		{
			name: "list all items",
			seedItems: []dto.CreateItemRequest{
				{Title: "Alpha", Description: "first", Status: "active"},
				{Title: "Bravo", Description: "second", Status: "inactive"},
				{Title: "Charlie", Description: "third", Status: "active"},
			},
			offset:    0,
			limit:     10,
			keyword:   "",
			status:    "",
			wantTotal: 3,
			wantLen:   3,
		},
		{
			name: "list by keyword",
			seedItems: []dto.CreateItemRequest{
				{Title: "Alpha", Description: "first", Status: "active"},
				{Title: "Bravo", Description: "find-me", Status: "inactive"},
				{Title: "Charlie", Description: "third", Status: "active"},
			},
			offset:    0,
			limit:     10,
			keyword:   "find-me",
			status:    "",
			wantTotal: 1,
			wantLen:   1,
		},
		{
			name: "list by status",
			seedItems: []dto.CreateItemRequest{
				{Title: "Alpha", Description: "first", Status: "active"},
				{Title: "Bravo", Description: "second", Status: "inactive"},
				{Title: "Charlie", Description: "third", Status: "active"},
			},
			offset:    0,
			limit:     10,
			keyword:   "",
			status:    "inactive",
			wantTotal: 1,
			wantLen:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := testItemService(t)
			for _, item := range tc.seedItems {
				_, err := svc.Create(&item, 1)
				if err != nil {
					t.Fatalf("Create(%q) error = %v", item.Title, err)
				}
			}

			items, total, err := svc.List(tc.offset, tc.limit, tc.keyword, tc.status)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if total != tc.wantTotal {
				t.Fatalf("total = %d, want %d", total, tc.wantTotal)
			}
			if len(items) != tc.wantLen {
				t.Fatalf("len(items) = %d, want %d", len(items), tc.wantLen)
			}
		})
	}
}
