package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
)

// ItemService handles item business logic.
type ItemService struct {
	itemRepo *repository.ItemRepository
}

// NewItemService creates a new ItemService.
func NewItemService(itemRepo *repository.ItemRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo}
}

// Create creates a new item.
func (s *ItemService) Create(req *dto.CreateItemRequest, userID uint) (*dto.ItemResponse, error) {
	item := &model.Item{
		Title:       req.Title,
		Description: req.Description,
		Status:      "active",
		UserID:      userID,
	}

	if req.Status != "" {
		item.Status = req.Status
	}

	if err := s.itemRepo.Create(item); err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.toResponse(item), nil
}

// GetByID retrieves an item by ID.
func (s *ItemService) GetByID(id uint) (*dto.ItemResponse, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFoundMsg
		}
		return nil, errcode.ErrInternalServer
	}
	return s.toResponse(item), nil
}

// List returns paginated items.
func (s *ItemService) List(offset, limit int, keyword, status string) ([]dto.ItemResponse, int64, error) {
	items, total, err := s.itemRepo.List(offset, limit, keyword, status)
	if err != nil {
		return nil, 0, errcode.ErrInternalServer
	}

	responses := make([]dto.ItemResponse, len(items))
	for i, item := range items {
		responses[i] = *s.toResponse(&item)
	}

	return responses, total, nil
}

// Update updates an existing item.
func (s *ItemService) Update(id uint, req *dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFoundMsg
		}
		return nil, errcode.ErrInternalServer
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Status != "" {
		item.Status = req.Status
	}

	if err := s.itemRepo.Update(item); err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.toResponse(item), nil
}

// Delete soft-deletes an item.
func (s *ItemService) Delete(id uint) error {
	if _, err := s.itemRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFoundMsg
		}
		return errcode.ErrInternalServer
	}

	return s.itemRepo.SoftDelete(id)
}

func (s *ItemService) toResponse(item *model.Item) *dto.ItemResponse {
	return &dto.ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Status:      item.Status,
		UserID:      item.UserID,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
