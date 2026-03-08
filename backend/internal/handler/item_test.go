package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
)

func testItemHandler(t *testing.T) (*ItemHandler, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestDB(t, &model.Item{}, &model.User{})
	repo := repository.NewItemRepository(db)
	svc := service.NewItemService(repo)
	h := NewItemHandler(svc)

	router := gin.New()
	v1 := router.Group("/api/v1")
	// Simulate auth by injecting user_id into context.
	v1.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	h.RegisterRoutes(v1, func(c *gin.Context) { c.Next() })

	return h, router
}

func decodePayload(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	return payload
}

func TestItemHandler_Create(t *testing.T) {
	_, router := testItemHandler(t)

	body, _ := json.Marshal(dto.CreateItemRequest{
		Title:       "New Item",
		Description: "Some desc",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	payload := decodePayload(t, resp.Body.Bytes())
	data := payload["data"].(map[string]any)
	if data["title"] != "New Item" {
		t.Fatalf("title = %v, want %q", data["title"], "New Item")
	}
}

func TestItemHandler_CreateValidationError(t *testing.T) {
	_, router := testItemHandler(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}

func TestItemHandler_GetByID(t *testing.T) {
	_, router := testItemHandler(t)

	// Create first
	body, _ := json.Marshal(dto.CreateItemRequest{Title: "Get Me"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	createPayload := decodePayload(t, createResp.Body.Bytes())
	data := createPayload["data"].(map[string]any)
	id := data["id"].(float64)

	// Get
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/items/%d", int(id)), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", getResp.Code, http.StatusOK)
	}
}

func TestItemHandler_GetByID_NotFound(t *testing.T) {
	_, router := testItemHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/9999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNotFound)
	}
}

func TestItemHandler_List(t *testing.T) {
	_, router := testItemHandler(t)

	for _, title := range []string{"Item1", "Item2", "Item3"} {
		body, _ := json.Marshal(dto.CreateItemRequest{Title: title})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?page=1&page_size=10", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodePayload(t, resp.Body.Bytes())
	data := payload["data"].(map[string]any)
	total := data["total"].(float64)
	if total != 3 {
		t.Fatalf("total = %v, want 3", total)
	}
}

func TestItemHandler_Update(t *testing.T) {
	_, router := testItemHandler(t)

	body, _ := json.Marshal(dto.CreateItemRequest{Title: "Original"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	createPayload := decodePayload(t, createResp.Body.Bytes())
	id := createPayload["data"].(map[string]any)["id"].(float64)

	updateBody, _ := json.Marshal(dto.UpdateItemRequest{Title: "Updated"})
	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/items/%d", int(id)), bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)

	if updateResp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", updateResp.Code, http.StatusOK)
	}

	payload := decodePayload(t, updateResp.Body.Bytes())
	if payload["data"].(map[string]any)["title"] != "Updated" {
		t.Fatalf("title after update = %v, want %q", payload["data"].(map[string]any)["title"], "Updated")
	}
}

func TestItemHandler_Delete(t *testing.T) {
	_, router := testItemHandler(t)

	body, _ := json.Marshal(dto.CreateItemRequest{Title: "Delete Me"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	createPayload := decodePayload(t, createResp.Body.Bytes())
	id := createPayload["data"].(map[string]any)["id"].(float64)

	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/items/%d", int(id)), nil)
	delResp := httptest.NewRecorder()
	router.ServeHTTP(delResp, delReq)

	if delResp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", delResp.Code, http.StatusOK)
	}

	// Verify item is gone
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/items/%d", int(id)), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusNotFound {
		t.Fatalf("after delete status = %d, want %d", getResp.Code, http.StatusNotFound)
	}
}
