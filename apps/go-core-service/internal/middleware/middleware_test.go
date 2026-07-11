package middleware

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
// 	"github.com/stretchr/testify/assert"
// )

// type mockUserRepo struct{}

// func (m *mockUserRepo) CreateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error) {
// 	return nil, nil
// }
// func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*UserEntity.User, error) {
// 	return nil, nil
// }
// func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*UserEntity.User, error) {
// 	return &UserEntity.User{
// 		ID:     uuid.MustParse(id),
// 		Email:  "test@example.com",
// 		Role:   UserEntity.RoleCustomer,
// 		Status: UserEntity.StatusActive,
// 	}, nil
// }
// func (m *mockUserRepo) UpdateUserStatus(ctx context.Context, id string, status UserEntity.Status) error {
// 	return nil
// }
// func (m *mockUserRepo) CheckEmailExists(ctx context.Context, email string) (bool, error) {
// 	return false, nil
// }
// func (m *mockUserRepo) CheckPhoneExists(ctx context.Context, phone string, excludeUserID string) (bool, error) {
// 	return false, nil
// }
// func (m *mockUserRepo) WriteAuditLog(ctx context.Context, content string, logType string) error {
// 	return nil
// }
// func (m *mockUserRepo) UpdateUser(ctx context.Context, user *UserEntity.User) (*UserEntity.User, error) {
// 	return nil, nil
// }
// func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
// 	return nil
// }
// func (m *mockUserRepo) ListUsers(ctx context.Context, page, limit int, role, status, search string) ([]*UserEntity.User, int64, error) {
// 	return nil, 0, nil
// }

// func TestAuthAndRoleMiddleware(t *testing.T) {
// 	gin.SetMode(gin.TestMode)
// 	os.Setenv("JWT_SECRET", "testsecret")
// 	defer os.Unsetenv("JWT_SECRET")

// 	t.Run("Valid Token - Allowed Role", func(t *testing.T) {
// 		userID := uuid.New().String()
// 		token, err := utils.GenerateAccessToken(userID, "admin@example.com", "ADMIN")
// 		assert.NoError(t, err)

// 		r := gin.New()
// 		repo := &mockUserRepo{}
// 		r.Use(AuthMiddleware(repo))
// 		r.Use(RoleMiddleware("ADMIN"))
// 		r.GET("/test", func(c *gin.Context) {
// 			c.JSON(http.StatusOK, gin.H{
// 				"user_id": utils.GetCurrentUserID(c),
// 				"email":   utils.GetCurrentEmail(c),
// 				"role":    utils.GetCurrentRole(c),
// 			})
// 		})

// 		req, _ := http.NewRequest("GET", "/test", nil)
// 		req.Header.Set("Authorization", "Bearer "+token)
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusOK, w.Code)
// 		assert.Contains(t, w.Body.String(), userID)
// 		assert.Contains(t, w.Body.String(), "admin@example.com")
// 		assert.Contains(t, w.Body.String(), "ADMIN")
// 	})

// 	t.Run("Valid Token - Disallowed Role", func(t *testing.T) {
// 		userID := uuid.New().String()
// 		token, err := utils.GenerateAccessToken(userID, "user@example.com", "USER")
// 		assert.NoError(t, err)

// 		r := gin.New()
// 		repo := &mockUserRepo{}
// 		r.Use(AuthMiddleware(repo))
// 		r.Use(RoleMiddleware("ADMIN"))
// 		r.GET("/test", func(c *gin.Context) {
// 			c.JSON(http.StatusOK, gin.H{"status": "ok"})
// 		})

// 		req, _ := http.NewRequest("GET", "/test", nil)
// 		req.Header.Set("Authorization", "Bearer "+token)
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusForbidden, w.Code)
// 	})

// 	t.Run("Missing Authorization Header", func(t *testing.T) {
// 		r := gin.New()
// 		repo := &mockUserRepo{}
// 		r.Use(AuthMiddleware(repo))
// 		r.GET("/test", func(c *gin.Context) {
// 			c.JSON(http.StatusOK, gin.H{"status": "ok"})
// 		})

// 		req, _ := http.NewRequest("GET", "/test", nil)
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)

// 		assert.Equal(t, http.StatusUnauthorized, w.Code)
// 	})
// }
