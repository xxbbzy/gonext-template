package repository

import (
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"gorm.io/gorm"
)

// ItemRepository handles item data access.
type ItemRepository struct {
	db *gorm.DB
}

// NewItemRepository creates a new ItemRepository.
func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create creates a new item.
func (r *ItemRepository) Create(item *model.Item) error {
	return r.db.Create(item).Error
}

// FindByID finds an item by ID.
func (r *ItemRepository) FindByID(id uint) (*model.Item, error) {
	var item model.Item
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// List returns paginated items with optional keyword search.
func (r *ItemRepository) List(offset, limit int, keyword, status string) ([]model.Item, int64, error) {
	var items []model.Item
	var total int64

	query := r.db.Model(&model.Item{})

	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Update updates an item.
func (r *ItemRepository) Update(item *model.Item) error {
	return r.db.Save(item).Error
}

// SoftDelete soft-deletes an item.
func (r *ItemRepository) SoftDelete(id uint) error {
	return r.db.Delete(&model.Item{}, id).Error
}
