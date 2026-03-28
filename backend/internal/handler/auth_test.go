package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

func testAuthHandler(t *testing.T) (*AuthHandler, *repository.UserRepository, *pkgjwt.Manager) {
	t.Helper()

	db := testutil.NewTestDB(t, &model.User{})

	jwtManager, err := pkgjwt.NewManager("test-secret", "15m", "24h")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtManager)
	return NewAuthHandler(authService), userRepo, jwtManager
}

func mustCreateAuthUser(t *testing.T, userRepo *repository.UserRepository, username, email, password, role string) *model.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func decodeAuthPayload(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return payload
}

func TestAuthRegisterReturnsCreatedUserPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, _, _ := testAuthHandler(t)

	body, err := json.Marshal(dto.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req

	authHandler.Register(ctx)

	if resp.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusCreated)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	data := payload["data"].(map[string]any)
	user := data["user"].(map[string]any)
	if user["username"] != "alice" {
		t.Fatalf("username = %v, want alice", user["username"])
	}
	if user["email"] != "alice@example.com" {
		t.Fatalf("email = %v, want alice@example.com", user["email"])
	}
	if user["role"] != "user" {
		t.Fatalf("role = %v, want user", user["role"])
	}
}

func TestAuthRegisterDuplicateEmailReturnsConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, _, _ := testAuthHandler(t)

	body, _ := json.Marshal(dto.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstResp := httptest.NewRecorder()
	firstCtx, _ := gin.CreateTestContext(firstResp)
	firstCtx.Request = firstReq
	authHandler.Register(firstCtx)

	if firstResp.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d", firstResp.Code, http.StatusCreated)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondResp := httptest.NewRecorder()
	secondCtx, _ := gin.CreateTestContext(secondResp)
	secondCtx.Request = secondReq
	authHandler.Register(secondCtx)

	if secondResp.Code != http.StatusConflict {
		t.Fatalf("second register status = %d, want %d", secondResp.Code, http.StatusConflict)
	}

	payload := decodeAuthPayload(t, secondResp.Body.Bytes())
	if payload["code"] != float64(errcode.ErrEmailExists) {
		t.Fatalf("error code = %v, want %d", payload["code"], errcode.ErrEmailExists)
	}
}

func TestAuthLoginReturnsAuthPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, userRepo, _ := testAuthHandler(t)
	mustCreateAuthUser(t, userRepo, "bob", "bob@example.com", "secret123", "admin")

	body, err := json.Marshal(dto.LoginRequest{
		Email:    "bob@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req

	authHandler.Login(ctx)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	if payload["code"] != float64(0) {
		t.Fatalf("code = %v, want 0", payload["code"])
	}

	data := payload["data"].(map[string]any)
	if data["access_token"] == "" || data["refresh_token"] == "" {
		t.Fatal("tokens should not be empty")
	}
	user := data["user"].(map[string]any)
	if user["email"] != "bob@example.com" {
		t.Fatalf("email = %v, want bob@example.com", user["email"])
	}
}

func TestAuthLoginWrongPasswordReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, userRepo, _ := testAuthHandler(t)
	mustCreateAuthUser(t, userRepo, "bob", "bob@example.com", "secret123", "admin")

	body, _ := json.Marshal(dto.LoginRequest{
		Email:    "bob@example.com",
		Password: "wrong-password",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req

	authHandler.Login(ctx)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	if payload["code"] != float64(errcode.ErrInvalidCreds) {
		t.Fatalf("code = %v, want %d", payload["code"], errcode.ErrInvalidCreds)
	}
}

func TestAuthGetProfileReturnsFullProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, userRepo, _ := testAuthHandler(t)
	user := mustCreateAuthUser(t, userRepo, "bob", "bob@example.com", "secret123", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/profile", nil)
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req
	ctx.Set("user_id", user.ID)

	authHandler.GetProfile(ctx)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	data := payload["data"].(map[string]any)
	if data["username"] != "bob" {
		t.Fatalf("username = %v, want bob", data["username"])
	}
	if data["email"] != "bob@example.com" {
		t.Fatalf("email = %v, want bob@example.com", data["email"])
	}
	if data["role"] != "admin" {
		t.Fatalf("role = %v, want admin", data["role"])
	}
}

func TestRegisterHealthRoutesLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router, func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	if payload["status"] != "alive" {
		t.Fatalf("status payload = %v, want alive", payload["status"])
	}
}

func TestRegisterHealthRoutesReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router, func() bool { return false })

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusServiceUnavailable)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	if payload["status"] != "not ready" {
		t.Fatalf("status payload = %v, want not ready", payload["status"])
	}
}

func TestRegisterHealthRoutesReadinessReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router, func() bool { return true })

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	payload := decodeAuthPayload(t, resp.Body.Bytes())
	if payload["status"] != "ready" {
		t.Fatalf("status payload = %v, want ready", payload["status"])
	}
}
