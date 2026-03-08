package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/pagination"
)

// Server implements StrictServerInterface, bridging generated types to existing services.
type Server struct {
	authService *service.AuthService
	itemService *service.ItemService
}

// NewServer creates a new Server.
func NewServer(authService *service.AuthService, itemService *service.ItemService) *Server {
	return &Server{
		authService: authService,
		itemService: itemService,
	}
}

// validate uses Gin's validator engine so that `binding:` struct tags are honored,
// preserving exact same validation behavior as c.ShouldBindJSON().
func validate(obj any) error {
	return binding.Validator.ValidateStruct(obj)
}

// --- helpers for the standard {code, data, message} envelope ---

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func derefOr[T any](p *T, fallback T) T {
	if p != nil {
		return *p
	}
	return fallback
}

// toAPIAuthResponse converts dto.AuthResponse → generated AuthResponse.
func toAPIAuthResponse(r *dto.AuthResponse) *AuthResponse {
	id := int(r.User.ID)
	return &AuthResponse{
		AccessToken:  &r.AccessToken,
		RefreshToken: &r.RefreshToken,
		User: &UserResponse{
			Id:       &id,
			Username: &r.User.Username,
			Email:    &r.User.Email,
			Role:     &r.User.Role,
		},
	}
}

// toAPIItemResponse converts dto.ItemResponse → generated ItemResponse.
func toAPIItemResponse(r *dto.ItemResponse) *ItemResponse {
	id := int(r.ID)
	uid := int(r.UserID)

	var createdAt *time.Time
	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		createdAt = &t
	}

	var updatedAt *time.Time
	if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
		updatedAt = &t
	}

	return &ItemResponse{
		Id:          &id,
		Title:       &r.Title,
		Description: &r.Description,
		Status:      &r.Status,
		UserId:      &uid,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// handleServiceError maps an *errcode.AppError to the appropriate HTTP error response.
func handleServiceError(err error) (code int, httpStatus int, message string) {
	if appErr, ok := err.(*errcode.AppError); ok {
		return appErr.Code, appErr.HTTPStatus, appErr.Message
	}
	return 500, http.StatusInternalServerError, "internal server error"
}

func loginUserErrorResponse(httpStatus, code int, message string) LoginUserResponseObject {
	switch httpStatus {
	case http.StatusBadRequest:
		return LoginUser400JSONResponse{Code: code, Message: message}
	case http.StatusUnauthorized:
		return LoginUser401JSONResponse{Code: code, Message: message}
	default:
		return LoginUser500JSONResponse{Code: code, Message: message}
	}
}

func getProfileErrorResponse(httpStatus, code int, message string) GetProfileResponseObject {
	switch httpStatus {
	case http.StatusUnauthorized:
		return GetProfile401JSONResponse{Code: code, Message: message}
	case http.StatusNotFound:
		return GetProfile404JSONResponse{Code: code, Message: message}
	default:
		return GetProfile500JSONResponse{Code: code, Message: message}
	}
}

func refreshTokenErrorResponse(httpStatus, code int, message string) RefreshTokenResponseObject {
	switch httpStatus {
	case http.StatusBadRequest:
		return RefreshToken400JSONResponse{Code: code, Message: message}
	case http.StatusUnauthorized:
		return RefreshToken401JSONResponse{Code: code, Message: message}
	default:
		return RefreshToken500JSONResponse{Code: code, Message: message}
	}
}

func registerUserErrorResponse(httpStatus, code int, message string) RegisterUserResponseObject {
	switch httpStatus {
	case http.StatusConflict:
		return RegisterUser409JSONResponse{Code: code, Message: message}
	case http.StatusBadRequest:
		return RegisterUser400JSONResponse{Code: code, Message: message}
	default:
		return RegisterUser500JSONResponse{Code: code, Message: message}
	}
}

func listItemsErrorResponse(httpStatus, code int, message string) ListItemsResponseObject {
	return ListItems500JSONResponse{Code: code, Message: message}
}

func createItemErrorResponse(httpStatus, code int, message string) CreateItemResponseObject {
	switch httpStatus {
	case http.StatusBadRequest:
		return CreateItem400JSONResponse{Code: code, Message: message}
	default:
		return CreateItem500JSONResponse{Code: code, Message: message}
	}
}

func deleteItemErrorResponse(httpStatus, code int, message string) DeleteItemResponseObject {
	switch httpStatus {
	case http.StatusNotFound:
		return DeleteItem404JSONResponse{Code: code, Message: message}
	default:
		return DeleteItem500JSONResponse{Code: code, Message: message}
	}
}

func getItemErrorResponse(httpStatus, code int, message string) GetItemResponseObject {
	switch httpStatus {
	case http.StatusNotFound:
		return GetItem404JSONResponse{Code: code, Message: message}
	default:
		return GetItem500JSONResponse{Code: code, Message: message}
	}
}

func updateItemErrorResponse(httpStatus, code int, message string) UpdateItemResponseObject {
	switch httpStatus {
	case http.StatusBadRequest:
		return UpdateItem400JSONResponse{Code: code, Message: message}
	case http.StatusNotFound:
		return UpdateItem404JSONResponse{Code: code, Message: message}
	default:
		return UpdateItem500JSONResponse{Code: code, Message: message}
	}
}

// ==================== Auth ====================

func (s *Server) RegisterUser(_ context.Context, request RegisterUserRequestObject) (RegisterUserResponseObject, error) {
	dtoReq := dto.RegisterRequest{
		Username: request.Body.Username,
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	if err := validate(dtoReq); err != nil {
		return registerUserErrorResponse(http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.Register(&dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return registerUserErrorResponse(httpStatus, code, msg), nil
	}

	return RegisterUser201JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIAuthResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) LoginUser(_ context.Context, request LoginUserRequestObject) (LoginUserResponseObject, error) {
	dtoReq := dto.LoginRequest{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	if err := validate(dtoReq); err != nil {
		return loginUserErrorResponse(http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.Login(&dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return loginUserErrorResponse(httpStatus, code, msg), nil
	}

	return LoginUser200JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIAuthResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) RefreshToken(_ context.Context, request RefreshTokenRequestObject) (RefreshTokenResponseObject, error) {
	dtoReq := dto.RefreshRequest{
		RefreshToken: request.Body.RefreshToken,
	}
	if err := validate(dtoReq); err != nil {
		return refreshTokenErrorResponse(http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.RefreshToken(dtoReq.RefreshToken)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return refreshTokenErrorResponse(httpStatus, code, msg), nil
	}

	return RefreshToken200JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIAuthResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, _ GetProfileRequestObject) (GetProfileResponseObject, error) {
	// user_id is set by the auth middleware in gin.Context.
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		// When using strict server, the context is stdlib context.
		// We need to reach into gin.Context which the strict handler passes as the Go context.
		return getProfileErrorResponse(http.StatusUnauthorized, 401, "unauthorized"), nil
	}

	userIDVal, exists := ginCtx.Get("user_id")
	if !exists {
		return getProfileErrorResponse(http.StatusUnauthorized, 401, "unauthorized"), nil
	}
	uid, ok := userIDVal.(uint)
	if !ok {
		return getProfileErrorResponse(http.StatusUnauthorized, 401, "unauthorized"), nil
	}

	profile, err := s.authService.GetProfile(uid)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return getProfileErrorResponse(httpStatus, code, msg), nil
	}

	id := int(profile.ID)
	return GetProfile200JSONResponse{
		Code: intPtr(0),
		Data: &UserResponse{
			Id:       &id,
			Username: &profile.Username,
			Email:    &profile.Email,
			Role:     &profile.Role,
		},
		Message: strPtr("success"),
	}, nil
}

// ==================== Items ====================

func (s *Server) ListItems(_ context.Context, request ListItemsRequestObject) (ListItemsResponseObject, error) {
	p := pagination.NewParams(
		derefOr(request.Params.Page, 1),
		derefOr(request.Params.PageSize, 10),
	)

	keyword := derefOr(request.Params.Keyword, "")
	status := ""
	if request.Params.Status != nil {
		status = string(*request.Params.Status)
	}

	items, total, err := s.itemService.List(p.Offset, p.PageSize, keyword, status)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return listItemsErrorResponse(httpStatus, code, msg), nil
	}

	// Convert dto items to API items
	apiItems := make([]ItemResponse, len(items))
	for i, item := range items {
		apiItems[i] = *toAPIItemResponse(&item)
	}

	totalInt := int(total)
	totalPages := totalInt / p.PageSize
	if totalInt%p.PageSize > 0 {
		totalPages++
	}

	return ListItems200JSONResponse{
		Code: intPtr(0),
		Data: &PagedItemsResponse{
			Items:      &apiItems,
			Total:      &totalInt,
			Page:       &p.Page,
			PageSize:   &p.PageSize,
			TotalPages: &totalPages,
		},
		Message: strPtr("success"),
	}, nil
}

func (s *Server) CreateItem(ctx context.Context, request CreateItemRequestObject) (CreateItemResponseObject, error) {
	dtoReq := dto.CreateItemRequest{
		Title:       request.Body.Title,
		Description: derefOr(request.Body.Description, ""),
	}
	if request.Body.Status != nil {
		dtoReq.Status = string(*request.Body.Status)
	}

	if err := validate(dtoReq); err != nil {
		return createItemErrorResponse(http.StatusBadRequest, 400, err.Error()), nil
	}

	// Get user_id from gin context
	var userID uint
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if uid, exists := ginCtx.Get("user_id"); exists {
			if v, ok := uid.(uint); ok {
				userID = v
			}
		}
	}

	result, err := s.itemService.Create(&dtoReq, userID)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return createItemErrorResponse(httpStatus, code, msg), nil
	}

	return CreateItem201JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIItemResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) GetItem(_ context.Context, request GetItemRequestObject) (GetItemResponseObject, error) {
	result, err := s.itemService.GetByID(uint(request.Id))
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return getItemErrorResponse(httpStatus, code, msg), nil
	}

	return GetItem200JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIItemResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) UpdateItem(_ context.Context, request UpdateItemRequestObject) (UpdateItemResponseObject, error) {
	dtoReq := dto.UpdateItemRequest{
		Title:       derefOr(request.Body.Title, ""),
		Description: derefOr(request.Body.Description, ""),
	}
	if request.Body.Status != nil {
		dtoReq.Status = string(*request.Body.Status)
	}

	if err := validate(dtoReq); err != nil {
		return updateItemErrorResponse(http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.itemService.Update(uint(request.Id), &dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return updateItemErrorResponse(httpStatus, code, msg), nil
	}

	return UpdateItem200JSONResponse{
		Code:    intPtr(0),
		Data:    toAPIItemResponse(result),
		Message: strPtr("success"),
	}, nil
}

func (s *Server) DeleteItem(_ context.Context, request DeleteItemRequestObject) (DeleteItemResponseObject, error) {
	if err := s.itemService.Delete(uint(request.Id)); err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return deleteItemErrorResponse(httpStatus, code, msg), nil
	}

	return DeleteItem200JSONResponse{
		Code:    intPtr(0),
		Message: strPtr("success"),
	}, nil
}
