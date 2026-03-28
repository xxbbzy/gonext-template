package service

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/testutil"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

func newAuthServiceForTest(t *testing.T, accessTTL, refreshTTL string) (*AuthService, *repository.UserRepository, *pkgjwt.Manager) {
	t.Helper()

	db := testutil.NewTestDB(t, &model.User{})
	userRepo := repository.NewUserRepository(db)

	jwtManager, err := pkgjwt.NewManager("test-secret", accessTTL, refreshTTL)
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	return NewAuthService(userRepo, jwtManager), userRepo, jwtManager
}

func mustCreateUser(t *testing.T, userRepo *repository.UserRepository, username, email, password, role string) *model.User {
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

func assertAppError(t *testing.T, err error, wantCode, wantStatus int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected app error code=%d status=%d, got nil", wantCode, wantStatus)
	}

	appErr, ok := err.(*errcode.AppError)
	if !ok {
		t.Fatalf("expected *errcode.AppError, got %T (%v)", err, err)
	}

	if appErr.Code != wantCode {
		t.Fatalf("error code = %d, want %d", appErr.Code, wantCode)
	}
	if appErr.HTTPStatus != wantStatus {
		t.Fatalf("http status = %d, want %d", appErr.HTTPStatus, wantStatus)
	}
}

func TestAuthService_Register(t *testing.T) {
	testCases := []struct {
		name       string
		prepare    func(t *testing.T, userRepo *repository.UserRepository)
		req        dto.RegisterRequest
		wantErr    bool
		wantCode   int
		wantStatus int
	}{
		{
			name: "register success",
			req: dto.RegisterRequest{
				Username: "alice",
				Email:    "alice@example.com",
				Password: "secret123",
			},
		},
		{
			name: "duplicate email fails",
			prepare: func(t *testing.T, userRepo *repository.UserRepository) {
				mustCreateUser(t, userRepo, "alice_existing", "alice@example.com", "secret123", "user")
			},
			req: dto.RegisterRequest{
				Username: "alice2",
				Email:    "alice@example.com",
				Password: "secret456",
			},
			wantErr:    true,
			wantCode:   errcode.ErrEmailExists,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, userRepo, _ := newAuthServiceForTest(t, "15m", "24h")
			if tc.prepare != nil {
				tc.prepare(t, userRepo)
			}

			resp, err := svc.Register(&tc.req)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("Register() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Register() response is nil")
			}
			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Fatal("Register() should return non-empty access/refresh token")
			}
			if resp.User.Email != tc.req.Email {
				t.Fatalf("user email = %q, want %q", resp.User.Email, tc.req.Email)
			}
			if resp.User.Username != tc.req.Username {
				t.Fatalf("username = %q, want %q", resp.User.Username, tc.req.Username)
			}
			if resp.User.Role != "user" {
				t.Fatalf("role = %q, want %q", resp.User.Role, "user")
			}

			stored, findErr := userRepo.FindByEmail(tc.req.Email)
			if findErr != nil {
				t.Fatalf("FindByEmail() error = %v", findErr)
			}
			if stored.PasswordHash == tc.req.Password {
				t.Fatal("stored password hash should not equal plaintext password")
			}
			if cmpErr := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(tc.req.Password)); cmpErr != nil {
				t.Fatalf("stored hash does not match plaintext password: %v", cmpErr)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	testCases := []struct {
		name          string
		prepare       func(t *testing.T, userRepo *repository.UserRepository)
		req           dto.LoginRequest
		wantErr       bool
		wantCode      int
		wantStatus    int
		wantUserEmail string
	}{
		{
			name: "login success",
			prepare: func(t *testing.T, userRepo *repository.UserRepository) {
				mustCreateUser(t, userRepo, "bob", "bob@example.com", "secret123", "admin")
			},
			req: dto.LoginRequest{
				Email:    "bob@example.com",
				Password: "secret123",
			},
			wantUserEmail: "bob@example.com",
		},
		{
			name: "wrong password fails",
			prepare: func(t *testing.T, userRepo *repository.UserRepository) {
				mustCreateUser(t, userRepo, "bob", "bob@example.com", "secret123", "admin")
			},
			req: dto.LoginRequest{
				Email:    "bob@example.com",
				Password: "wrong-password",
			},
			wantErr:    true,
			wantCode:   errcode.ErrInvalidCreds,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent email fails",
			req: dto.LoginRequest{
				Email:    "nobody@example.com",
				Password: "secret123",
			},
			wantErr:    true,
			wantCode:   errcode.ErrInvalidCreds,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, userRepo, _ := newAuthServiceForTest(t, "15m", "24h")
			if tc.prepare != nil {
				tc.prepare(t, userRepo)
			}

			resp, err := svc.Login(&tc.req)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("Login() error = %v", err)
			}
			if resp == nil {
				t.Fatal("Login() response is nil")
			}
			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Fatal("Login() should return non-empty access/refresh token")
			}
			if resp.User.Email != tc.wantUserEmail {
				t.Fatalf("user email = %q, want %q", resp.User.Email, tc.wantUserEmail)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	testCases := []struct {
		name       string
		accessTTL  string
		refreshTTL string
		prepare    func(t *testing.T, userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) string
		wait       time.Duration
		wantErr    bool
		wantCode   int
		wantStatus int
	}{
		{
			name:       "refresh success",
			accessTTL:  "15m",
			refreshTTL: "24h",
			prepare: func(t *testing.T, userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) string {
				user := mustCreateUser(t, userRepo, "carol", "carol@example.com", "secret123", "user")
				token, err := jwtManager.GenerateRefreshToken(user.ID, user.Role)
				if err != nil {
					t.Fatalf("GenerateRefreshToken() error = %v", err)
				}
				return token
			},
		},
		{
			name:       "invalid token fails",
			accessTTL:  "15m",
			refreshTTL: "24h",
			prepare: func(_ *testing.T, _ *repository.UserRepository, _ *pkgjwt.Manager) string {
				return "invalid-token"
			},
			wantErr:    true,
			wantCode:   errcode.ErrTokenInvalid,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token fails",
			accessTTL:  "15m",
			refreshTTL: "2s",
			prepare: func(t *testing.T, userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) string {
				user := mustCreateUser(t, userRepo, "dave", "dave@example.com", "secret123", "user")
				token, err := jwtManager.GenerateRefreshToken(user.ID, user.Role)
				if err != nil {
					t.Fatalf("GenerateRefreshToken() error = %v", err)
				}
				return token
			},
			wait:       2500 * time.Millisecond,
			wantErr:    true,
			wantCode:   errcode.ErrTokenExpired,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, userRepo, jwtManager := newAuthServiceForTest(t, tc.accessTTL, tc.refreshTTL)
			refreshToken := tc.prepare(t, userRepo, jwtManager)

			if tc.wait > 0 {
				time.Sleep(tc.wait)
			}

			resp, err := svc.RefreshToken(refreshToken)
			if tc.wantErr {
				assertAppError(t, err, tc.wantCode, tc.wantStatus)
				return
			}

			if err != nil {
				t.Fatalf("RefreshToken() error = %v", err)
			}
			if resp == nil {
				t.Fatal("RefreshToken() response is nil")
			}
			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Fatal("RefreshToken() should return non-empty access/refresh token")
			}
		})
	}
}

func TestAuthService_GetProfile(t *testing.T) {
	svc, userRepo, _ := newAuthServiceForTest(t, "15m", "24h")
	user := mustCreateUser(t, userRepo, "erin", "erin@example.com", "secret123", "admin")

	profile, err := svc.GetProfile(user.ID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if profile.ID != user.ID {
		t.Fatalf("profile id = %d, want %d", profile.ID, user.ID)
	}
	if profile.Email != user.Email {
		t.Fatalf("profile email = %q, want %q", profile.Email, user.Email)
	}
	if profile.Role != user.Role {
		t.Fatalf("profile role = %q, want %q", profile.Role, user.Role)
	}

	_, err = svc.GetProfile(9999)
	assertAppError(t, err, errcode.ErrNotFound, http.StatusNotFound)
}
