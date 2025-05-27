package handlers_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-short/internal/config"
	"url-short/internal/handlers"
	"url-short/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthHandler(t *testing.T) (*handlers.AuthHandler, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	cfg := &config.Config{
		JWTSecret: "test-secret-1234567890",
	}

	return handlers.NewAuthHandler(userRepo, cfg), mock, db
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		requestBody  string
		mockClosure  func(mock sqlmock.Sqlmock)
		expectedCode int
		expectedBody string
	}{
		{
			name:        "Success",
			requestBody: `{"username": "testuser", "email": "test@example.com", "password": "123456"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "test@example.com", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedCode: http.StatusCreated,
			expectedBody: `"message":"Пользователь создан"`,
		},
		{
			name:         "Invalid JSON",
			requestBody:  `{invalid-json}`,
			mockClosure:  func(mock sqlmock.Sqlmock) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `"error":"Неверные данные"`,
		},
		{
			name:        "Duplicate Email",
			requestBody: `{"username": "testuser", "email": "exists@example.com", "password": "123456"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `"error":"Ошибка при создании пользователя"`,
		},
		{
			name:         "Weak Password",
			requestBody:  `{"username": "testuser", "email": "test@example.com", "password": "1"}`,
			mockClosure:  func(mock sqlmock.Sqlmock) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `"error":"Неверные данные"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, db := setupAuthHandler(t)
			defer db.Close()

			tt.mockClosure(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(
				"POST",
				"/api/register",
				strings.NewReader(tt.requestBody),
			)
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Register(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validPassword := "correct-password-123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		requestBody  string
		mockClosure  func(mock sqlmock.Sqlmock)
		expectedCode int
		checkToken   bool
	}{
		{
			name:        "Success",
			requestBody: `{"email": "correct@example.com", "password": "` + validPassword + `"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE email = ?").
					WithArgs("correct@example.com").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
							AddRow(1, "testuser", "correct@example.com", hashedPassword),
					)
			},
			expectedCode: http.StatusOK,
			checkToken:   true,
		},
		{
			name:        "User Not Found",
			requestBody: `{"email": "notfound@example.com", "password": "any"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE email = ?").
					WithArgs("notfound@example.com").
					WillReturnError(repositories.ErrUserNotFound)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:        "Invalid Password",
			requestBody: `{"email": "correct@example.com", "password": "wrong-password"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE email = ?").
					WithArgs("correct@example.com").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
							AddRow(1, "testuser", "correct@example.com", "invalid-hash"),
					)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:        "Database Error",
			requestBody: `{"email": "error@example.com", "password": "any"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT ...").
					WillReturnError(errors.New("database error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, db := setupAuthHandler(t)
			defer db.Close()

			tt.mockClosure(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(
				"POST",
				"/api/login",
				strings.NewReader(tt.requestBody),
			)
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Login(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.checkToken {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tokenString := response["token"]
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret-1234567890"), nil
				})

				assert.NoError(t, err)
				assert.True(t, token.Valid)
				claims := token.Claims.(jwt.MapClaims)
				assert.Equal(t, float64(1), claims["user_id"])
			}
		})
	}
}
