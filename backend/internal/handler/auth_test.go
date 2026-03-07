package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

func testAuthHandler(t *testing.T) (*AuthHandler, *gorm.DB, *pkgjwt.Manager) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	jwtManager, err := pkgjwt.NewManager("test-secret", "15m", "24h")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtManager)
	return NewAuthHandler(authService), db, jwtManager
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

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

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

func TestAuthGetProfileReturnsFullProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authHandler, db, _ := testAuthHandler(t)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &model.User{
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: string(passwordHash),
		Role:         "admin",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/profile", nil)
	resp := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(resp)
	ctx.Request = req
	ctx.Set("user_id", user.ID)

	authHandler.GetProfile(ctx)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

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
}
