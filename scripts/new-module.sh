#!/bin/bash

# Generate a new backend module with handler, service, repository, model, and DTO files.
# Usage: ./scripts/new-module.sh <module_name>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <module_name>"
    echo "Example: $0 product"
    exit 1
fi

MODULE_NAME=$1
MODULE_NAME_LOWER=$(echo "$MODULE_NAME" | tr '[:upper:]' '[:lower:]')
MODULE_NAME_UPPER=$(echo "${MODULE_NAME_LOWER^}")
BACKEND_DIR="backend/internal"

echo "Generating module: $MODULE_NAME_LOWER"

# Model
cat > "$BACKEND_DIR/model/${MODULE_NAME_LOWER}.go" << EOF
package model

import (
	"time"

	"gorm.io/gorm"
)

// ${MODULE_NAME_UPPER} represents a ${MODULE_NAME_LOWER} resource.
type ${MODULE_NAME_UPPER} struct {
	ID        uint           \`json:"id" gorm:"primaryKey"\`
	Name      string         \`json:"name" gorm:"size:200;not null"\`
	CreatedAt time.Time      \`json:"created_at"\`
	UpdatedAt time.Time      \`json:"updated_at"\`
	DeletedAt gorm.DeletedAt \`json:"-" gorm:"index"\`
}

func (${MODULE_NAME_UPPER}) TableName() string {
	return "${MODULE_NAME_LOWER}s"
}
EOF
echo "  Created model/${MODULE_NAME_LOWER}.go"

# DTO
cat > "$BACKEND_DIR/dto/${MODULE_NAME_LOWER}.go" << EOF
package dto

type Create${MODULE_NAME_UPPER}Request struct {
	Name string \`json:"name" binding:"required,max=200"\`
}

type Update${MODULE_NAME_UPPER}Request struct {
	Name string \`json:"name" binding:"omitempty,max=200"\`
}

type ${MODULE_NAME_UPPER}Response struct {
	ID        uint   \`json:"id"\`
	Name      string \`json:"name"\`
	CreatedAt string \`json:"created_at"\`
	UpdatedAt string \`json:"updated_at"\`
}
EOF
echo "  Created dto/${MODULE_NAME_LOWER}.go"

# Repository
cat > "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}.go" << EOF
package repository

import (
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"gorm.io/gorm"
)

type ${MODULE_NAME_UPPER}Repository struct {
	db *gorm.DB
}

func New${MODULE_NAME_UPPER}Repository(db *gorm.DB) *${MODULE_NAME_UPPER}Repository {
	return &${MODULE_NAME_UPPER}Repository{db: db}
}

func (r *${MODULE_NAME_UPPER}Repository) Create(item *model.${MODULE_NAME_UPPER}) error {
	return r.db.Create(item).Error
}

func (r *${MODULE_NAME_UPPER}Repository) FindByID(id uint) (*model.${MODULE_NAME_UPPER}, error) {
	var item model.${MODULE_NAME_UPPER}
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *${MODULE_NAME_UPPER}Repository) List(offset, limit int) ([]model.${MODULE_NAME_UPPER}, int64, error) {
	var items []model.${MODULE_NAME_UPPER}
	var total int64
	r.db.Model(&model.${MODULE_NAME_UPPER}{}).Count(&total)
	r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&items)
	return items, total, nil
}

func (r *${MODULE_NAME_UPPER}Repository) Update(item *model.${MODULE_NAME_UPPER}) error {
	return r.db.Save(item).Error
}

func (r *${MODULE_NAME_UPPER}Repository) Delete(id uint) error {
	return r.db.Delete(&model.${MODULE_NAME_UPPER}{}, id).Error
}
EOF
echo "  Created repository/${MODULE_NAME_LOWER}.go"

# Service
cat > "$BACKEND_DIR/service/${MODULE_NAME_LOWER}.go" << EOF
package service

import (
	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
)

type ${MODULE_NAME_UPPER}Service struct {
	repo *repository.${MODULE_NAME_UPPER}Repository
}

func New${MODULE_NAME_UPPER}Service(repo *repository.${MODULE_NAME_UPPER}Repository) *${MODULE_NAME_UPPER}Service {
	return &${MODULE_NAME_UPPER}Service{repo: repo}
}

func (s *${MODULE_NAME_UPPER}Service) Create(req *dto.Create${MODULE_NAME_UPPER}Request) (*model.${MODULE_NAME_UPPER}, error) {
	item := &model.${MODULE_NAME_UPPER}{Name: req.Name}
	return item, s.repo.Create(item)
}
EOF
echo "  Created service/${MODULE_NAME_LOWER}.go"

# Handler
cat > "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}.go" << EOF
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

type ${MODULE_NAME_UPPER}Handler struct {
	service *service.${MODULE_NAME_UPPER}Service
}

func New${MODULE_NAME_UPPER}Handler(svc *service.${MODULE_NAME_UPPER}Service) *${MODULE_NAME_UPPER}Handler {
	return &${MODULE_NAME_UPPER}Handler{service: svc}
}

func (h *${MODULE_NAME_UPPER}Handler) Create(c *gin.Context) {
	var req dto.Create${MODULE_NAME_UPPER}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.service.Create(&req)
	if err != nil {
		response.InternalServerError(c, "failed to create")
		return
	}
	response.Created(c, result)
}

func (h *${MODULE_NAME_UPPER}Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	group := r.Group("/${MODULE_NAME_LOWER}s")
	group.Use(authMiddleware)
	{
		group.POST("", h.Create)
	}
}
EOF
echo "  Created handler/${MODULE_NAME_LOWER}.go"

# Repository Test
cat > "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}_test.go" << EOF
package repository

import (
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
)

func Test${MODULE_NAME_UPPER}Repository_CreateAndFindByID(t *testing.T) {
	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := New${MODULE_NAME_UPPER}Repository(db)

	item := &model.${MODULE_NAME_UPPER}{Name: "Test ${MODULE_NAME_UPPER}"}
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
	if found.Name != "Test ${MODULE_NAME_UPPER}" {
		t.Fatalf("FindByID() name = %q, want %q", found.Name, "Test ${MODULE_NAME_UPPER}")
	}
}

func Test${MODULE_NAME_UPPER}Repository_List(t *testing.T) {
	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := New${MODULE_NAME_UPPER}Repository(db)

	repo.Create(&model.${MODULE_NAME_UPPER}{Name: "A"})
	repo.Create(&model.${MODULE_NAME_UPPER}{Name: "B"})

	items, total, err := repo.List(0, 10)
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

func Test${MODULE_NAME_UPPER}Repository_Delete(t *testing.T) {
	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := New${MODULE_NAME_UPPER}Repository(db)

	item := &model.${MODULE_NAME_UPPER}{Name: "To Delete"}
	repo.Create(item)

	if err := repo.Delete(item.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := repo.FindByID(item.ID)
	if err == nil {
		t.Fatal("FindByID() after Delete should return error")
	}
}
EOF
echo "  Created repository/${MODULE_NAME_LOWER}_test.go"

# Handler Test
cat > "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}_test.go" << EOF
package handler

import (
	"bytes"
	"encoding/json"
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

func test${MODULE_NAME_UPPER}Handler(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := repository.New${MODULE_NAME_UPPER}Repository(db)
	svc := service.New${MODULE_NAME_UPPER}Service(repo)
	h := New${MODULE_NAME_UPPER}Handler(svc)

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	h.RegisterRoutes(v1, func(c *gin.Context) { c.Next() })

	return router
}

func Test${MODULE_NAME_UPPER}Handler_Create(t *testing.T) {
	router := test${MODULE_NAME_UPPER}Handler(t)

	body, _ := json.Marshal(dto.Create${MODULE_NAME_UPPER}Request{Name: "New ${MODULE_NAME_UPPER}"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/${MODULE_NAME_LOWER}s", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", resp.Code, http.StatusCreated, resp.Body.String())
	}
}
EOF
echo "  Created handler/${MODULE_NAME_LOWER}_test.go"

echo ""
echo "Module '$MODULE_NAME_LOWER' generated successfully!"
echo "Don't forget to:"
echo "  1. Register routes in cmd/server/main.go"
echo "  2. Add AutoMigrate for the new model"
echo "  3. Wire the dependencies"
