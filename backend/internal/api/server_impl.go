package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/pagination"
	"github.com/xxbbzy/gonext-template/backend/pkg/requestlog"
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

func derefOr[T any](p *T, fallback T) T {
	if p != nil {
		return *p
	}
	return fallback
}

// toAPIAuthResponse converts dto.AuthResponse → generated AuthResponse.
func toAPIAuthResponse(r *dto.AuthResponse) AuthResponse {
	id := int(r.User.ID)
	return AuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		User: UserResponse{
			Id:       id,
			Username: r.User.Username,
			Email:    r.User.Email,
			Role:     r.User.Role,
		},
	}
}

// toAPIItemResponse converts dto.ItemResponse → generated ItemResponse.
func toAPIItemResponse(r *dto.ItemResponse) ItemResponse {
	id := int(r.ID)
	uid := int(r.UserID)

	var createdAt time.Time
	if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
		createdAt = t
	}

	var updatedAt time.Time
	if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
		updatedAt = t
	}

	return ItemResponse{
		Id:          id,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		UserId:      uid,
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

func requestIDFromContext(ctx context.Context) string {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return ""
	}

	requestID, exists := ginCtx.Get(middleware.RequestIDKey)
	if !exists {
		return ""
	}

	value, ok := requestID.(string)
	if !ok {
		return ""
	}

	return value
}

func errorResponseBody(ctx context.Context, code int, message string) ErrorResponse {
	requestlog.SetErrorCodeFromContext(ctx, code)
	return ErrorResponse{
		Code:    code,
		Data:    nil,
		Message: message,
	}
}

func loginUserErrorResponse(ctx context.Context, httpStatus, code int, message string) LoginUserResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusBadRequest:
		return LoginUser400JSONResponse{
			Body:    body,
			Headers: LoginUser400ResponseHeaders{XRequestID: requestID},
		}
	case http.StatusUnauthorized:
		return LoginUser401JSONResponse{
			Body:    body,
			Headers: LoginUser401ResponseHeaders{XRequestID: requestID},
		}
	default:
		return LoginUser500JSONResponse{
			Body:    body,
			Headers: LoginUser500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func getProfileErrorResponse(ctx context.Context, httpStatus, code int, message string) GetProfileResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusUnauthorized:
		return GetProfile401JSONResponse{
			Body:    body,
			Headers: GetProfile401ResponseHeaders{XRequestID: requestID},
		}
	case http.StatusNotFound:
		return GetProfile404JSONResponse{
			Body:    body,
			Headers: GetProfile404ResponseHeaders{XRequestID: requestID},
		}
	default:
		return GetProfile500JSONResponse{
			Body:    body,
			Headers: GetProfile500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func refreshTokenErrorResponse(ctx context.Context, httpStatus, code int, message string) RefreshTokenResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusBadRequest:
		return RefreshToken400JSONResponse{
			Body:    body,
			Headers: RefreshToken400ResponseHeaders{XRequestID: requestID},
		}
	case http.StatusUnauthorized:
		return RefreshToken401JSONResponse{
			Body:    body,
			Headers: RefreshToken401ResponseHeaders{XRequestID: requestID},
		}
	default:
		return RefreshToken500JSONResponse{
			Body:    body,
			Headers: RefreshToken500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func registerUserErrorResponse(ctx context.Context, httpStatus, code int, message string) RegisterUserResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusConflict:
		return RegisterUser409JSONResponse{
			Body:    body,
			Headers: RegisterUser409ResponseHeaders{XRequestID: requestID},
		}
	case http.StatusBadRequest:
		return RegisterUser400JSONResponse{
			Body:    body,
			Headers: RegisterUser400ResponseHeaders{XRequestID: requestID},
		}
	default:
		return RegisterUser500JSONResponse{
			Body:    body,
			Headers: RegisterUser500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func listItemsErrorResponse(ctx context.Context, code int, message string) ListItemsResponseObject {
	return ListItems500JSONResponse{
		Body:    errorResponseBody(ctx, code, message),
		Headers: ListItems500ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}
}

func createItemErrorResponse(ctx context.Context, httpStatus, code int, message string) CreateItemResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusBadRequest:
		return CreateItem400JSONResponse{
			Body:    body,
			Headers: CreateItem400ResponseHeaders{XRequestID: requestID},
		}
	default:
		return CreateItem500JSONResponse{
			Body:    body,
			Headers: CreateItem500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func deleteItemErrorResponse(ctx context.Context, httpStatus, code int, message string) DeleteItemResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusNotFound:
		return DeleteItem404JSONResponse{
			Body:    body,
			Headers: DeleteItem404ResponseHeaders{XRequestID: requestID},
		}
	default:
		return DeleteItem500JSONResponse{
			Body:    body,
			Headers: DeleteItem500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func getItemErrorResponse(ctx context.Context, httpStatus, code int, message string) GetItemResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusNotFound:
		return GetItem404JSONResponse{
			Body:    body,
			Headers: GetItem404ResponseHeaders{XRequestID: requestID},
		}
	default:
		return GetItem500JSONResponse{
			Body:    body,
			Headers: GetItem500ResponseHeaders{XRequestID: requestID},
		}
	}
}

func updateItemErrorResponse(ctx context.Context, httpStatus, code int, message string) UpdateItemResponseObject {
	body := errorResponseBody(ctx, code, message)
	requestID := requestIDFromContext(ctx)

	switch httpStatus {
	case http.StatusBadRequest:
		return UpdateItem400JSONResponse{
			Body:    body,
			Headers: UpdateItem400ResponseHeaders{XRequestID: requestID},
		}
	case http.StatusNotFound:
		return UpdateItem404JSONResponse{
			Body:    body,
			Headers: UpdateItem404ResponseHeaders{XRequestID: requestID},
		}
	default:
		return UpdateItem500JSONResponse{
			Body:    body,
			Headers: UpdateItem500ResponseHeaders{XRequestID: requestID},
		}
	}
}

// ==================== Auth ====================

func (s *Server) RegisterUser(ctx context.Context, request RegisterUserRequestObject) (RegisterUserResponseObject, error) {
	dtoReq := dto.RegisterRequest{
		Username: request.Body.Username,
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	if err := validate(dtoReq); err != nil {
		return registerUserErrorResponse(ctx, http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.Register(&dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return registerUserErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return RegisterUser201JSONResponse{
		Body: AuthSuccessResponse{
			Code:    AuthSuccessResponseCodeN0,
			Data:    toAPIAuthResponse(result),
			Message: AuthSuccessResponseMessageSuccess,
		},
		Headers: RegisterUser201ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, request LoginUserRequestObject) (LoginUserResponseObject, error) {
	dtoReq := dto.LoginRequest{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}
	if err := validate(dtoReq); err != nil {
		return loginUserErrorResponse(ctx, http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.Login(&dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return loginUserErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return LoginUser200JSONResponse{
		Body: AuthSuccessResponse{
			Code:    AuthSuccessResponseCodeN0,
			Data:    toAPIAuthResponse(result),
			Message: AuthSuccessResponseMessageSuccess,
		},
		Headers: LoginUser200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, request RefreshTokenRequestObject) (RefreshTokenResponseObject, error) {
	dtoReq := dto.RefreshRequest{
		RefreshToken: request.Body.RefreshToken,
	}
	if err := validate(dtoReq); err != nil {
		return refreshTokenErrorResponse(ctx, http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.authService.RefreshToken(dtoReq.RefreshToken)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return refreshTokenErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return RefreshToken200JSONResponse{
		Body: AuthSuccessResponse{
			Code:    AuthSuccessResponseCodeN0,
			Data:    toAPIAuthResponse(result),
			Message: AuthSuccessResponseMessageSuccess,
		},
		Headers: RefreshToken200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, _ GetProfileRequestObject) (GetProfileResponseObject, error) {
	// user_id is set by the auth middleware in gin.Context.
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		// When using strict server, the context is stdlib context.
		// We need to reach into gin.Context which the strict handler passes as the Go context.
		return getProfileErrorResponse(ctx, http.StatusUnauthorized, 401, "unauthorized"), nil
	}

	userIDVal, exists := ginCtx.Get("user_id")
	if !exists {
		return getProfileErrorResponse(ctx, http.StatusUnauthorized, 401, "unauthorized"), nil
	}
	uid, ok := userIDVal.(uint)
	if !ok {
		return getProfileErrorResponse(ctx, http.StatusUnauthorized, 401, "unauthorized"), nil
	}

	profile, err := s.authService.GetProfile(uid)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return getProfileErrorResponse(ctx, httpStatus, code, msg), nil
	}

	id := int(profile.ID)
	return GetProfile200JSONResponse{
		Body: UserSuccessResponse{
			Code: UserSuccessResponseCode(0),
			Data: UserResponse{
				Id:       id,
				Username: profile.Username,
				Email:    profile.Email,
				Role:     profile.Role,
			},
			Message: UserSuccessResponseMessage("success"),
		},
		Headers: GetProfile200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

// ==================== Items ====================

func (s *Server) ListItems(ctx context.Context, request ListItemsRequestObject) (ListItemsResponseObject, error) {
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
		code, _, msg := handleServiceError(err)
		return listItemsErrorResponse(ctx, code, msg), nil
	}

	// Convert dto items to API items
	apiItems := make([]ItemResponse, len(items))
	for i, item := range items {
		apiItems[i] = toAPIItemResponse(&item)
	}

	totalInt := int(total)
	totalPages := totalInt / p.PageSize
	if totalInt%p.PageSize > 0 {
		totalPages++
	}

	return ListItems200JSONResponse{
		Body: PagedItemsSuccessResponse{
			Code: PagedItemsSuccessResponseCodeN0,
			Data: PagedItemsResponse{
				Items:      apiItems,
				Total:      totalInt,
				Page:       p.Page,
				PageSize:   p.PageSize,
				TotalPages: totalPages,
			},
			Message: PagedItemsSuccessResponseMessageSuccess,
		},
		Headers: ListItems200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
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
		return createItemErrorResponse(ctx, http.StatusBadRequest, 400, err.Error()), nil
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
		return createItemErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return CreateItem201JSONResponse{
		Body: ItemSuccessResponse{
			Code:    ItemSuccessResponseCodeN0,
			Data:    toAPIItemResponse(result),
			Message: ItemSuccessResponseMessageSuccess,
		},
		Headers: CreateItem201ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) GetItem(ctx context.Context, request GetItemRequestObject) (GetItemResponseObject, error) {
	result, err := s.itemService.GetByID(uint(request.Id))
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return getItemErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return GetItem200JSONResponse{
		Body: ItemSuccessResponse{
			Code:    ItemSuccessResponseCodeN0,
			Data:    toAPIItemResponse(result),
			Message: ItemSuccessResponseMessageSuccess,
		},
		Headers: GetItem200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) UpdateItem(ctx context.Context, request UpdateItemRequestObject) (UpdateItemResponseObject, error) {
	dtoReq := dto.UpdateItemRequest{
		Title:       derefOr(request.Body.Title, ""),
		Description: derefOr(request.Body.Description, ""),
	}
	if request.Body.Status != nil {
		dtoReq.Status = string(*request.Body.Status)
	}

	if err := validate(dtoReq); err != nil {
		return updateItemErrorResponse(ctx, http.StatusBadRequest, 400, err.Error()), nil
	}

	result, err := s.itemService.Update(uint(request.Id), &dtoReq)
	if err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return updateItemErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return UpdateItem200JSONResponse{
		Body: ItemSuccessResponse{
			Code:    ItemSuccessResponseCodeN0,
			Data:    toAPIItemResponse(result),
			Message: ItemSuccessResponseMessageSuccess,
		},
		Headers: UpdateItem200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}

func (s *Server) DeleteItem(ctx context.Context, request DeleteItemRequestObject) (DeleteItemResponseObject, error) {
	if err := s.itemService.Delete(uint(request.Id)); err != nil {
		code, httpStatus, msg := handleServiceError(err)
		return deleteItemErrorResponse(ctx, httpStatus, code, msg), nil
	}

	return DeleteItem200JSONResponse{
		Body: EmptySuccessResponse{
			Code:    EmptySuccessResponseCodeN0,
			Data:    nil,
			Message: EmptySuccessResponseMessageSuccess,
		},
		Headers: DeleteItem200ResponseHeaders{XRequestID: requestIDFromContext(ctx)},
	}, nil
}
