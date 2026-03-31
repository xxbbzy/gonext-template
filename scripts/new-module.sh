#!/bin/bash

# Generate a new backend module with convention-aligned boilerplate.
# Usage: ./scripts/new-module.sh <module_name>

set -euo pipefail

usage() {
  echo "Usage: $0 <module_name>"
  echo "Example: $0 product"
}

normalize_module_name() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr '-' '_'
}

is_go_keyword() {
  case "$1" in
    break|default|func|interface|select|case|defer|go|map|struct|chan|else|goto|package|switch|const|fallthrough|if|range|type|continue|for|import|return|var)
      return 0
      ;;
  esac

  return 1
}

to_pascal_case() {
  local input="$1"
  local result=""
  local segment=""

  IFS='_' read -r -a parts <<< "$input"
  for segment in "${parts[@]}"; do
    if [ -z "$segment" ]; then
      continue
    fi
    result+="$(printf '%s%s' "$(printf '%s' "${segment:0:1}" | tr '[:lower:]' '[:upper:]')" "${segment:1}")"
  done

  printf '%s' "$result"
}

lower_first() {
  local input="$1"
  local first="${input:0:1}"
  printf '%s%s' "$(printf '%s' "$first" | tr '[:upper:]' '[:lower:]')" "${input:1}"
}

pluralize() {
  local name="$1"

  case "$name" in
    *[sxz]|*ch|*sh)
      printf '%ses' "$name"
      ;;
    *[!aeiou]y)
      printf '%sies' "${name%y}"
      ;;
    *)
      printf '%ss' "$name"
      ;;
  esac
}

write_file() {
  local path="$1"
  shift

  cat > "$path" <<EOF
$*
EOF
}

if [ $# -ne 1 ] || [ -z "${1:-}" ]; then
  usage
  exit 1
fi

MODULE_NAME_RAW="$1"
MODULE_NAME_LOWER="$(normalize_module_name "$MODULE_NAME_RAW")"

if [[ ! "$MODULE_NAME_LOWER" =~ ^[a-z][a-z0-9_]*$ ]]; then
  echo "module_name must start with a letter and contain only letters, numbers, '_' or '-'." >&2
  exit 1
fi

if is_go_keyword "$MODULE_NAME_LOWER"; then
  echo "module_name must not be a Go keyword after normalization." >&2
  exit 1
fi

MODULE_NAME_UPPER="$(to_pascal_case "$MODULE_NAME_LOWER")"
MODULE_NAME_VAR="$(lower_first "$MODULE_NAME_UPPER")"
MODULE_NAME_PLURAL="$(pluralize "$MODULE_NAME_LOWER")"
MODULE_NAME_PLURAL_UPPER="$(to_pascal_case "$MODULE_NAME_PLURAL")"
MODULE_NAME_PLURAL_VAR="$(lower_first "$MODULE_NAME_PLURAL_UPPER")"
BACKEND_DIR="backend/internal"

TARGET_FILES=(
  "$BACKEND_DIR/model/${MODULE_NAME_LOWER}.go"
  "$BACKEND_DIR/dto/${MODULE_NAME_LOWER}.go"
  "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}.go"
  "$BACKEND_DIR/service/${MODULE_NAME_LOWER}.go"
  "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}.go"
  "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}_test.go"
  "$BACKEND_DIR/service/${MODULE_NAME_LOWER}_test.go"
  "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}_test.go"
)

for file in "${TARGET_FILES[@]}"; do
  if [ -e "$file" ]; then
    echo "Refusing to overwrite existing file: $file" >&2
    exit 1
  fi
done

mkdir -p \
  "$BACKEND_DIR/model" \
  "$BACKEND_DIR/dto" \
  "$BACKEND_DIR/repository" \
  "$BACKEND_DIR/service" \
  "$BACKEND_DIR/handler"

echo "Generating module: $MODULE_NAME_LOWER"

write_file "$BACKEND_DIR/model/${MODULE_NAME_LOWER}.go" "package model

import (
	\"time\"

	\"gorm.io/gorm\"
)

// ${MODULE_NAME_UPPER} represents the ${MODULE_NAME_LOWER} persistence model.
type ${MODULE_NAME_UPPER} struct {
	ID        uint           \`json:\"id\" gorm:\"primaryKey\"\`
	Name      string         \`json:\"name\" gorm:\"size:200;not null\"\`
	CreatedAt time.Time      \`json:\"created_at\"\`
	UpdatedAt time.Time      \`json:\"updated_at\"\`
	DeletedAt gorm.DeletedAt \`json:\"-\" gorm:\"index\"\`
}

// TableName overrides the default table name.
func (${MODULE_NAME_UPPER}) TableName() string {
	return \"${MODULE_NAME_PLURAL}\"
}
"
echo "  Created model/${MODULE_NAME_LOWER}.go"

write_file "$BACKEND_DIR/dto/${MODULE_NAME_LOWER}.go" "package dto

// Create${MODULE_NAME_UPPER}Request represents the create ${MODULE_NAME_LOWER} request body.
type Create${MODULE_NAME_UPPER}Request struct {
	Name string \`json:\"name\" binding:\"required,max=200\"\`
}

// Update${MODULE_NAME_UPPER}Request represents the update ${MODULE_NAME_LOWER} request body.
type Update${MODULE_NAME_UPPER}Request struct {
	Name string \`json:\"name\" binding:\"omitempty,max=200\"\`
}

// ${MODULE_NAME_UPPER}Response represents ${MODULE_NAME_LOWER} data in API responses.
type ${MODULE_NAME_UPPER}Response struct {
	ID        uint   \`json:\"id\"\`
	Name      string \`json:\"name\"\`
	CreatedAt string \`json:\"created_at\"\`
	UpdatedAt string \`json:\"updated_at\"\`
}

// List${MODULE_NAME_PLURAL_UPPER}Query represents list ${MODULE_NAME_PLURAL} query params.
type List${MODULE_NAME_PLURAL_UPPER}Query struct {
	Page     int \`form:\"page\"\`
	PageSize int \`form:\"page_size\"\`
}
"
echo "  Created dto/${MODULE_NAME_LOWER}.go"

write_file "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}.go" "package repository

import (
	\"gorm.io/gorm\"

	\"github.com/xxbbzy/gonext-template/backend/internal/model\"
)

// ${MODULE_NAME_UPPER}Repository handles ${MODULE_NAME_LOWER} data access.
type ${MODULE_NAME_UPPER}Repository struct {
	db *gorm.DB
}

// New${MODULE_NAME_UPPER}Repository creates a new ${MODULE_NAME_UPPER}Repository.
func New${MODULE_NAME_UPPER}Repository(db *gorm.DB) *${MODULE_NAME_UPPER}Repository {
	return &${MODULE_NAME_UPPER}Repository{db: db}
}

// Create inserts a new ${MODULE_NAME_LOWER}.
func (r *${MODULE_NAME_UPPER}Repository) Create(${MODULE_NAME_VAR} *model.${MODULE_NAME_UPPER}) error {
	return r.db.Create(${MODULE_NAME_VAR}).Error
}

// FindByID loads a ${MODULE_NAME_LOWER} by primary key.
func (r *${MODULE_NAME_UPPER}Repository) FindByID(id uint) (*model.${MODULE_NAME_UPPER}, error) {
	var ${MODULE_NAME_VAR} model.${MODULE_NAME_UPPER}
	if err := r.db.First(&${MODULE_NAME_VAR}, id).Error; err != nil {
		return nil, err
	}
	return &${MODULE_NAME_VAR}, nil
}

// List returns a paginated list of ${MODULE_NAME_PLURAL}.
func (r *${MODULE_NAME_UPPER}Repository) List(offset, limit int) ([]model.${MODULE_NAME_UPPER}, int64, error) {
	var ${MODULE_NAME_PLURAL_VAR} []model.${MODULE_NAME_UPPER}
	var total int64

	query := r.db.Model(&model.${MODULE_NAME_UPPER}{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order(\"created_at DESC\").Find(&${MODULE_NAME_PLURAL_VAR}).Error; err != nil {
		return nil, 0, err
	}

	return ${MODULE_NAME_PLURAL_VAR}, total, nil
}

// Update persists changes to an existing ${MODULE_NAME_LOWER}.
func (r *${MODULE_NAME_UPPER}Repository) Update(${MODULE_NAME_VAR} *model.${MODULE_NAME_UPPER}) error {
	return r.db.Save(${MODULE_NAME_VAR}).Error
}

// SoftDelete soft deletes a ${MODULE_NAME_LOWER} by ID.
func (r *${MODULE_NAME_UPPER}Repository) SoftDelete(id uint) error {
	return r.db.Delete(&model.${MODULE_NAME_UPPER}{}, id).Error
}
"
echo "  Created repository/${MODULE_NAME_LOWER}.go"

write_file "$BACKEND_DIR/service/${MODULE_NAME_LOWER}.go" "package service

import (
	\"errors\"
	\"time\"

	\"gorm.io/gorm\"

	\"github.com/xxbbzy/gonext-template/backend/internal/dto\"
	\"github.com/xxbbzy/gonext-template/backend/internal/model\"
	\"github.com/xxbbzy/gonext-template/backend/internal/repository\"
	\"github.com/xxbbzy/gonext-template/backend/pkg/errcode\"
)

// ${MODULE_NAME_UPPER}Service handles ${MODULE_NAME_LOWER} business logic.
type ${MODULE_NAME_UPPER}Service struct {
	${MODULE_NAME_VAR}Repo *repository.${MODULE_NAME_UPPER}Repository
}

// New${MODULE_NAME_UPPER}Service creates a new ${MODULE_NAME_UPPER}Service.
func New${MODULE_NAME_UPPER}Service(${MODULE_NAME_VAR}Repo *repository.${MODULE_NAME_UPPER}Repository) *${MODULE_NAME_UPPER}Service {
	return &${MODULE_NAME_UPPER}Service{${MODULE_NAME_VAR}Repo: ${MODULE_NAME_VAR}Repo}
}

// Create persists a new ${MODULE_NAME_LOWER} and returns its response DTO.
func (s *${MODULE_NAME_UPPER}Service) Create(req *dto.Create${MODULE_NAME_UPPER}Request) (*dto.${MODULE_NAME_UPPER}Response, error) {
	${MODULE_NAME_VAR} := &model.${MODULE_NAME_UPPER}{
		Name: req.Name,
	}

	if err := s.${MODULE_NAME_VAR}Repo.Create(${MODULE_NAME_VAR}); err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.toResponse(${MODULE_NAME_VAR}), nil
}

// GetByID loads a ${MODULE_NAME_LOWER} by ID.
func (s *${MODULE_NAME_UPPER}Service) GetByID(id uint) (*dto.${MODULE_NAME_UPPER}Response, error) {
	${MODULE_NAME_VAR}, err := s.${MODULE_NAME_VAR}Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFoundMsg
		}
		return nil, errcode.ErrInternalServer
	}

	return s.toResponse(${MODULE_NAME_VAR}), nil
}

// List returns paginated ${MODULE_NAME_PLURAL} DTOs.
func (s *${MODULE_NAME_UPPER}Service) List(offset, limit int) ([]dto.${MODULE_NAME_UPPER}Response, int64, error) {
	${MODULE_NAME_PLURAL_VAR}, total, err := s.${MODULE_NAME_VAR}Repo.List(offset, limit)
	if err != nil {
		return nil, 0, errcode.ErrInternalServer
	}

	responses := make([]dto.${MODULE_NAME_UPPER}Response, len(${MODULE_NAME_PLURAL_VAR}))
	for i, ${MODULE_NAME_VAR} := range ${MODULE_NAME_PLURAL_VAR} {
		responses[i] = *s.toResponse(&${MODULE_NAME_VAR})
	}

	return responses, total, nil
}

// Update modifies an existing ${MODULE_NAME_LOWER}.
func (s *${MODULE_NAME_UPPER}Service) Update(id uint, req *dto.Update${MODULE_NAME_UPPER}Request) (*dto.${MODULE_NAME_UPPER}Response, error) {
	${MODULE_NAME_VAR}, err := s.${MODULE_NAME_VAR}Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFoundMsg
		}
		return nil, errcode.ErrInternalServer
	}

	if req.Name != \"\" {
		${MODULE_NAME_VAR}.Name = req.Name
	}

	if err := s.${MODULE_NAME_VAR}Repo.Update(${MODULE_NAME_VAR}); err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.toResponse(${MODULE_NAME_VAR}), nil
}

// Delete soft deletes an existing ${MODULE_NAME_LOWER}.
func (s *${MODULE_NAME_UPPER}Service) Delete(id uint) error {
	if _, err := s.${MODULE_NAME_VAR}Repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFoundMsg
		}
		return errcode.ErrInternalServer
	}

	if err := s.${MODULE_NAME_VAR}Repo.SoftDelete(id); err != nil {
		return errcode.ErrInternalServer
	}

	return nil
}

func (s *${MODULE_NAME_UPPER}Service) toResponse(${MODULE_NAME_VAR} *model.${MODULE_NAME_UPPER}) *dto.${MODULE_NAME_UPPER}Response {
	return &dto.${MODULE_NAME_UPPER}Response{
		ID:        ${MODULE_NAME_VAR}.ID,
		Name:      ${MODULE_NAME_VAR}.Name,
		CreatedAt: ${MODULE_NAME_VAR}.CreatedAt.Format(time.RFC3339),
		UpdatedAt: ${MODULE_NAME_VAR}.UpdatedAt.Format(time.RFC3339),
	}
}
"
echo "  Created service/${MODULE_NAME_LOWER}.go"

write_file "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}.go" "package handler

import (
	\"strconv\"

	\"github.com/gin-gonic/gin\"

	\"github.com/xxbbzy/gonext-template/backend/internal/dto\"
	\"github.com/xxbbzy/gonext-template/backend/internal/service\"
	\"github.com/xxbbzy/gonext-template/backend/pkg/errcode\"
	\"github.com/xxbbzy/gonext-template/backend/pkg/pagination\"
	\"github.com/xxbbzy/gonext-template/backend/pkg/response\"
)

// ${MODULE_NAME_UPPER}Handler handles ${MODULE_NAME_LOWER} HTTP requests.
type ${MODULE_NAME_UPPER}Handler struct {
	${MODULE_NAME_VAR}Service *service.${MODULE_NAME_UPPER}Service
}

// New${MODULE_NAME_UPPER}Handler creates a new ${MODULE_NAME_UPPER}Handler.
func New${MODULE_NAME_UPPER}Handler(${MODULE_NAME_VAR}Service *service.${MODULE_NAME_UPPER}Service) *${MODULE_NAME_UPPER}Handler {
	return &${MODULE_NAME_UPPER}Handler{${MODULE_NAME_VAR}Service: ${MODULE_NAME_VAR}Service}
}

// Create handles ${MODULE_NAME_LOWER} creation.
func (h *${MODULE_NAME_UPPER}Handler) Create(c *gin.Context) {
	var req dto.Create${MODULE_NAME_UPPER}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.${MODULE_NAME_VAR}Service.Create(&req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, \"failed to create ${MODULE_NAME_LOWER}\")
		return
	}

	response.Created(c, result)
}

// GetByID handles loading a ${MODULE_NAME_LOWER} by ID.
func (h *${MODULE_NAME_UPPER}Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param(\"id\"), 10, 32)
	if err != nil {
		response.BadRequest(c, \"invalid ${MODULE_NAME_LOWER} ID\")
		return
	}

	result, svcErr := h.${MODULE_NAME_VAR}Service.GetByID(uint(id))
	if svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, \"failed to get ${MODULE_NAME_LOWER}\")
		return
	}

	response.Success(c, result)
}

// List handles paginated ${MODULE_NAME_PLURAL} responses.
func (h *${MODULE_NAME_UPPER}Handler) List(c *gin.Context) {
	p := pagination.Parse(c)

	${MODULE_NAME_PLURAL_VAR}, total, err := h.${MODULE_NAME_VAR}Service.List(p.Offset, p.PageSize)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, \"failed to list ${MODULE_NAME_PLURAL}\")
		return
	}

	response.PagedSuccess(c, ${MODULE_NAME_PLURAL_VAR}, total, p.Page, p.PageSize)
}

// Update handles ${MODULE_NAME_LOWER} updates.
func (h *${MODULE_NAME_UPPER}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param(\"id\"), 10, 32)
	if err != nil {
		response.BadRequest(c, \"invalid ${MODULE_NAME_LOWER} ID\")
		return
	}

	var req dto.Update${MODULE_NAME_UPPER}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, svcErr := h.${MODULE_NAME_VAR}Service.Update(uint(id), &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, \"failed to update ${MODULE_NAME_LOWER}\")
		return
	}

	response.Success(c, result)
}

// Delete handles ${MODULE_NAME_LOWER} deletion.
func (h *${MODULE_NAME_UPPER}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param(\"id\"), 10, 32)
	if err != nil {
		response.BadRequest(c, \"invalid ${MODULE_NAME_LOWER} ID\")
		return
	}

	if svcErr := h.${MODULE_NAME_VAR}Service.Delete(uint(id)); svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, \"failed to delete ${MODULE_NAME_LOWER}\")
		return
	}

	response.Success(c, nil)
}

// RegisterRoutes registers ${MODULE_NAME_LOWER} routes.
func (h *${MODULE_NAME_UPPER}Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, protectedMiddlewares ...gin.HandlerFunc) {
	${MODULE_NAME_PLURAL_VAR} := r.Group(\"/${MODULE_NAME_PLURAL}\")
	${MODULE_NAME_PLURAL_VAR}.Use(authMiddleware)
	if len(protectedMiddlewares) > 0 {
		${MODULE_NAME_PLURAL_VAR}.Use(protectedMiddlewares...)
	}

	{
		${MODULE_NAME_PLURAL_VAR}.POST(\"\", h.Create)
		${MODULE_NAME_PLURAL_VAR}.GET(\"\", h.List)
		${MODULE_NAME_PLURAL_VAR}.GET(\"/:id\", h.GetByID)
		${MODULE_NAME_PLURAL_VAR}.PUT(\"/:id\", h.Update)
		${MODULE_NAME_PLURAL_VAR}.DELETE(\"/:id\", h.Delete)
	}
}
"
echo "  Created handler/${MODULE_NAME_LOWER}.go"

write_file "$BACKEND_DIR/repository/${MODULE_NAME_LOWER}_test.go" "package repository

import (
	\"testing\"

	\"github.com/xxbbzy/gonext-template/backend/internal/model\"
	\"github.com/xxbbzy/gonext-template/backend/internal/testutil\"
)

func test${MODULE_NAME_UPPER}Repo(t *testing.T) *${MODULE_NAME_UPPER}Repository {
	t.Helper()
	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	return New${MODULE_NAME_UPPER}Repository(db)
}

func Test${MODULE_NAME_UPPER}Repository_CreateAndFindByID(t *testing.T) {
	repo := test${MODULE_NAME_UPPER}Repo(t)

	${MODULE_NAME_VAR} := &model.${MODULE_NAME_UPPER}{Name: \"Test ${MODULE_NAME_UPPER}\"}
	if err := repo.Create(${MODULE_NAME_VAR}); err != nil {
		t.Fatalf(\"Create() error = %v\", err)
	}
	if ${MODULE_NAME_VAR}.ID == 0 {
		t.Fatal(\"Create() did not set ID\")
	}

	found, err := repo.FindByID(${MODULE_NAME_VAR}.ID)
	if err != nil {
		t.Fatalf(\"FindByID() error = %v\", err)
	}
	if found.Name != \"Test ${MODULE_NAME_UPPER}\" {
		t.Fatalf(\"FindByID() name = %q, want %q\", found.Name, \"Test ${MODULE_NAME_UPPER}\")
	}
}

func Test${MODULE_NAME_UPPER}Repository_List(t *testing.T) {
	repo := test${MODULE_NAME_UPPER}Repo(t)

	for _, name := range []string{\"Alpha\", \"Beta\"} {
		if err := repo.Create(&model.${MODULE_NAME_UPPER}{Name: name}); err != nil {
			t.Fatalf(\"Create(%q) error = %v\", name, err)
		}
	}

	${MODULE_NAME_PLURAL_VAR}, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf(\"List() error = %v\", err)
	}
	if total != 2 {
		t.Fatalf(\"List() total = %d, want 2\", total)
	}
	if len(${MODULE_NAME_PLURAL_VAR}) != 2 {
		t.Fatalf(\"List() len = %d, want 2\", len(${MODULE_NAME_PLURAL_VAR}))
	}
}

func Test${MODULE_NAME_UPPER}Repository_UpdateAndSoftDelete(t *testing.T) {
	repo := test${MODULE_NAME_UPPER}Repo(t)

	${MODULE_NAME_VAR} := &model.${MODULE_NAME_UPPER}{Name: \"Original\"}
	if err := repo.Create(${MODULE_NAME_VAR}); err != nil {
		t.Fatalf(\"Create() error = %v\", err)
	}

	${MODULE_NAME_VAR}.Name = \"Updated\"
	if err := repo.Update(${MODULE_NAME_VAR}); err != nil {
		t.Fatalf(\"Update() error = %v\", err)
	}

	found, err := repo.FindByID(${MODULE_NAME_VAR}.ID)
	if err != nil {
		t.Fatalf(\"FindByID() error = %v\", err)
	}
	if found.Name != \"Updated\" {
		t.Fatalf(\"FindByID() name = %q, want %q\", found.Name, \"Updated\")
	}

	if err := repo.SoftDelete(${MODULE_NAME_VAR}.ID); err != nil {
		t.Fatalf(\"SoftDelete() error = %v\", err)
	}
	if _, err := repo.FindByID(${MODULE_NAME_VAR}.ID); err == nil {
		t.Fatal(\"FindByID() after SoftDelete should return error\")
	}
}
"
echo "  Created repository/${MODULE_NAME_LOWER}_test.go"

write_file "$BACKEND_DIR/service/${MODULE_NAME_LOWER}_test.go" "package service

import (
	\"net/http\"
	\"testing\"

	\"github.com/xxbbzy/gonext-template/backend/internal/dto\"
	\"github.com/xxbbzy/gonext-template/backend/internal/model\"
	\"github.com/xxbbzy/gonext-template/backend/internal/repository\"
	\"github.com/xxbbzy/gonext-template/backend/internal/testutil\"
	\"github.com/xxbbzy/gonext-template/backend/pkg/errcode\"
)

func test${MODULE_NAME_UPPER}Service(t *testing.T) *${MODULE_NAME_UPPER}Service {
	t.Helper()
	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := repository.New${MODULE_NAME_UPPER}Repository(db)
	return New${MODULE_NAME_UPPER}Service(repo)
}

func assert${MODULE_NAME_UPPER}AppError(t *testing.T, err error, wantCode, wantStatus int) {
	t.Helper()

	appErr, ok := err.(*errcode.AppError)
	if !ok {
		t.Fatalf(\"expected *errcode.AppError, got %T (%v)\", err, err)
	}
	if appErr.Code != wantCode {
		t.Fatalf(\"error code = %d, want %d\", appErr.Code, wantCode)
	}
	if appErr.HTTPStatus != wantStatus {
		t.Fatalf(\"http status = %d, want %d\", appErr.HTTPStatus, wantStatus)
	}
}

func Test${MODULE_NAME_UPPER}Service_CreateAndGetByID(t *testing.T) {
	svc := test${MODULE_NAME_UPPER}Service(t)

	created, err := svc.Create(&dto.Create${MODULE_NAME_UPPER}Request{Name: \"Created ${MODULE_NAME_UPPER}\"})
	if err != nil {
		t.Fatalf(\"Create() error = %v\", err)
	}

	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf(\"GetByID() error = %v\", err)
	}
	if found.Name != \"Created ${MODULE_NAME_UPPER}\" {
		t.Fatalf(\"GetByID() name = %q, want %q\", found.Name, \"Created ${MODULE_NAME_UPPER}\")
	}
}

func Test${MODULE_NAME_UPPER}Service_List(t *testing.T) {
	svc := test${MODULE_NAME_UPPER}Service(t)

	for _, name := range []string{\"First\", \"Second\"} {
		if _, err := svc.Create(&dto.Create${MODULE_NAME_UPPER}Request{Name: name}); err != nil {
			t.Fatalf(\"Create(%q) error = %v\", name, err)
		}
	}

	${MODULE_NAME_PLURAL_VAR}, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf(\"List() error = %v\", err)
	}
	if total != 2 {
		t.Fatalf(\"List() total = %d, want 2\", total)
	}
	if len(${MODULE_NAME_PLURAL_VAR}) != 2 {
		t.Fatalf(\"List() len = %d, want 2\", len(${MODULE_NAME_PLURAL_VAR}))
	}
}

func Test${MODULE_NAME_UPPER}Service_Update(t *testing.T) {
	svc := test${MODULE_NAME_UPPER}Service(t)

	created, err := svc.Create(&dto.Create${MODULE_NAME_UPPER}Request{Name: \"Original\"})
	if err != nil {
		t.Fatalf(\"Create() error = %v\", err)
	}

	updated, err := svc.Update(created.ID, &dto.Update${MODULE_NAME_UPPER}Request{Name: \"Updated\"})
	if err != nil {
		t.Fatalf(\"Update() error = %v\", err)
	}
	if updated.Name != \"Updated\" {
		t.Fatalf(\"Update() name = %q, want %q\", updated.Name, \"Updated\")
	}
}

func Test${MODULE_NAME_UPPER}Service_Delete(t *testing.T) {
	svc := test${MODULE_NAME_UPPER}Service(t)

	created, err := svc.Create(&dto.Create${MODULE_NAME_UPPER}Request{Name: \"Delete Me\"})
	if err != nil {
		t.Fatalf(\"Create() error = %v\", err)
	}

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf(\"Delete() error = %v\", err)
	}
	if _, err := svc.GetByID(created.ID); err == nil {
		t.Fatal(\"GetByID() after Delete should return error\")
	}
}

func Test${MODULE_NAME_UPPER}Service_GetByID_NotFound(t *testing.T) {
	svc := test${MODULE_NAME_UPPER}Service(t)

	if _, err := svc.GetByID(9999); err == nil {
		t.Fatal(\"GetByID() expected not-found error\")
	} else {
		assert${MODULE_NAME_UPPER}AppError(t, err, errcode.ErrNotFound, http.StatusNotFound)
	}
}
"
echo "  Created service/${MODULE_NAME_LOWER}_test.go"

write_file "$BACKEND_DIR/handler/${MODULE_NAME_LOWER}_test.go" "package handler

import (
	\"bytes\"
	\"encoding/json\"
	\"fmt\"
	\"net/http\"
	\"net/http/httptest\"
	\"testing\"

	\"github.com/gin-gonic/gin\"

	\"github.com/xxbbzy/gonext-template/backend/internal/dto\"
	\"github.com/xxbbzy/gonext-template/backend/internal/model\"
	\"github.com/xxbbzy/gonext-template/backend/internal/repository\"
	\"github.com/xxbbzy/gonext-template/backend/internal/service\"
	\"github.com/xxbbzy/gonext-template/backend/internal/testutil\"
)

func test${MODULE_NAME_UPPER}Handler(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.NewTestDB(t, &model.${MODULE_NAME_UPPER}{})
	repo := repository.New${MODULE_NAME_UPPER}Repository(db)
	svc := service.New${MODULE_NAME_UPPER}Service(repo)
	h := New${MODULE_NAME_UPPER}Handler(svc)

	router := gin.New()
	v1 := router.Group(\"/api/v1\")
	h.RegisterRoutes(v1, func(c *gin.Context) { c.Next() })

	return router
}

func decode${MODULE_NAME_UPPER}Payload(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf(\"unmarshal response: %v\", err)
	}
	return payload
}

func create${MODULE_NAME_UPPER}(t *testing.T, router *gin.Engine, name string) int {
	t.Helper()

	body, err := json.Marshal(dto.Create${MODULE_NAME_UPPER}Request{Name: name})
	if err != nil {
		t.Fatalf(\"marshal request: %v\", err)
	}

	req := httptest.NewRequest(http.MethodPost, \"/api/v1/${MODULE_NAME_PLURAL}\", bytes.NewReader(body))
	req.Header.Set(\"Content-Type\", \"application/json\")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf(\"create status = %d, want %d; body = %s\", resp.Code, http.StatusCreated, resp.Body.String())
	}

	payload := decode${MODULE_NAME_UPPER}Payload(t, resp.Body.Bytes())
	return int(payload[\"data\"].(map[string]any)[\"id\"].(float64))
}

func Test${MODULE_NAME_UPPER}Handler_Create(t *testing.T) {
	router := test${MODULE_NAME_UPPER}Handler(t)

	id := create${MODULE_NAME_UPPER}(t, router, \"New ${MODULE_NAME_UPPER}\")
	if id == 0 {
		t.Fatal(\"Create() should return a non-zero ID\")
	}
}

func Test${MODULE_NAME_UPPER}Handler_CreateValidationError(t *testing.T) {
	router := test${MODULE_NAME_UPPER}Handler(t)

	req := httptest.NewRequest(http.MethodPost, \"/api/v1/${MODULE_NAME_PLURAL}\", bytes.NewReader([]byte(\"{}\")))
	req.Header.Set(\"Content-Type\", \"application/json\")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf(\"status = %d, want %d\", resp.Code, http.StatusBadRequest)
	}
}

func Test${MODULE_NAME_UPPER}Handler_GetByIDAndList(t *testing.T) {
	router := test${MODULE_NAME_UPPER}Handler(t)

	id := create${MODULE_NAME_UPPER}(t, router, \"List Me\")
	create${MODULE_NAME_UPPER}(t, router, \"List Me Too\")

	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf(\"/api/v1/${MODULE_NAME_PLURAL}/%d\", id), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf(\"GetByID() status = %d, want %d\", getResp.Code, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, \"/api/v1/${MODULE_NAME_PLURAL}?page=1&page_size=10\", nil)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)

	if listResp.Code != http.StatusOK {
		t.Fatalf(\"List() status = %d, want %d\", listResp.Code, http.StatusOK)
	}

	payload := decode${MODULE_NAME_UPPER}Payload(t, listResp.Body.Bytes())
	total := payload[\"data\"].(map[string]any)[\"total\"].(float64)
	if total != 2 {
		t.Fatalf(\"List() total = %v, want 2\", total)
	}
}

func Test${MODULE_NAME_UPPER}Handler_UpdateAndDelete(t *testing.T) {
	router := test${MODULE_NAME_UPPER}Handler(t)
	id := create${MODULE_NAME_UPPER}(t, router, \"Original\")

	updateBody, err := json.Marshal(dto.Update${MODULE_NAME_UPPER}Request{Name: \"Updated\"})
	if err != nil {
		t.Fatalf(\"marshal update request: %v\", err)
	}

	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf(\"/api/v1/${MODULE_NAME_PLURAL}/%d\", id), bytes.NewReader(updateBody))
	updateReq.Header.Set(\"Content-Type\", \"application/json\")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)

	if updateResp.Code != http.StatusOK {
		t.Fatalf(\"Update() status = %d, want %d\", updateResp.Code, http.StatusOK)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf(\"/api/v1/${MODULE_NAME_PLURAL}/%d\", id), nil)
	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)

	if deleteResp.Code != http.StatusOK {
		t.Fatalf(\"Delete() status = %d, want %d\", deleteResp.Code, http.StatusOK)
	}
}
"
echo "  Created handler/${MODULE_NAME_LOWER}_test.go"

echo ""
echo "Module '$MODULE_NAME_LOWER' generated successfully!"
echo "Follow-up checklist:"
echo "  [ ] Review api/openapi.yaml before finalizing routes or response shapes"
echo "  [ ] Register New${MODULE_NAME_UPPER}Repository, New${MODULE_NAME_UPPER}Service, and New${MODULE_NAME_UPPER}Handler in backend/cmd/server/providers.go and backend/cmd/server/wire.go"
echo "  [ ] Mount ${MODULE_NAME_UPPER} routes in backend/cmd/server/main.go"
echo "  [ ] Register &model.${MODULE_NAME_UPPER}{} in development AutoMigrate and add backend/migrations when persistence changes ship"
echo "  [ ] Run make gen-types after contract changes; run make gen when generated server/docs artifacts must be refreshed"
echo "  [ ] Run make check"
echo "  [ ] Run make e2e for API or runtime behavior changes"
