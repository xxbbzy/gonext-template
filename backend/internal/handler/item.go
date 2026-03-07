package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/pagination"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

// ItemHandler handles item HTTP requests.
type ItemHandler struct {
	itemService *service.ItemService
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

// Create handles item creation.
// @Summary Create a new item
// @Tags Items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateItemRequest true "Item data"
// @Success 201 {object} response.Response{data=dto.ItemResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/items [post]
func (h *ItemHandler) Create(c *gin.Context) {
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	result, err := h.itemService.Create(&req, uid)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to create item")
		return
	}

	response.Created(c, result)
}

// GetByID handles getting an item by ID.
// @Summary Get an item by ID
// @Tags Items
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} response.Response{data=dto.ItemResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/items/{id} [get]
func (h *ItemHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	result, svcErr := h.itemService.GetByID(uint(id))
	if svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to get item")
		return
	}

	response.Success(c, result)
}

// List handles listing items with pagination.
// @Summary List items with pagination and search
// @Tags Items
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param keyword query string false "Search keyword"
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Response{data=response.PagedData}
// @Router /api/v1/items [get]
func (h *ItemHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	keyword := c.Query("keyword")
	status := c.Query("status")

	items, total, err := h.itemService.List(p.Offset, p.PageSize, keyword, status)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to list items")
		return
	}

	response.PagedSuccess(c, items, total, p.Page, p.PageSize)
}

// Update handles item updates.
// @Summary Update an item
// @Tags Items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param body body dto.UpdateItemRequest true "Update data"
// @Success 200 {object} response.Response{data=dto.ItemResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/items/{id} [put]
func (h *ItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, svcErr := h.itemService.Update(uint(id), &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to update item")
		return
	}

	response.Success(c, result)
}

// Delete handles item deletion.
// @Summary Delete an item (soft delete)
// @Tags Items
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/items/{id} [delete]
func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid item ID")
		return
	}

	if svcErr := h.itemService.Delete(uint(id)); svcErr != nil {
		if appErr, ok := svcErr.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to delete item")
		return
	}

	response.Success(c, nil)
}

// RegisterRoutes registers item routes.
func (h *ItemHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, protectedMiddlewares ...gin.HandlerFunc) {
	items := r.Group("/items")
	items.Use(authMiddleware)
	if len(protectedMiddlewares) > 0 {
		items.Use(protectedMiddlewares...)
	}
	{
		items.POST("", h.Create)
		items.GET("", h.List)
		items.GET("/:id", h.GetByID)
		items.PUT("/:id", h.Update)
		items.DELETE("/:id", h.Delete)
	}
}
