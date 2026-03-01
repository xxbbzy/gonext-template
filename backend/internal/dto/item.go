package dto

// CreateItemRequest represents the create item request body.
type CreateItemRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdateItemRequest represents the update item request body.
type UpdateItemRequest struct {
	Title       string `json:"title" binding:"omitempty,max=200"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ItemResponse represents item data in responses.
type ItemResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserID      uint   `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListItemsQuery represents the query parameters for listing items.
type ListItemsQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
}
