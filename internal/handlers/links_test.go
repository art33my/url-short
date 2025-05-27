package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"url-short/internal/handlers"
	"url-short/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupLinkHandler(t *testing.T) (*handlers.LinkHandler, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	linkRepo := repositories.NewLinkRepository(db)
	analyticRepo := repositories.NewAnalyticRepository(db)

	return &handlers.LinkHandler{
		LinkRepo:     linkRepo,
		AnalyticRepo: analyticRepo,
	}, mock, db
}

func TestCreateShortLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		requestBody  string
		mockClosure  func(mock sqlmock.Sqlmock)
		expectedCode int
		expectedBody string
	}{
		{
			name:        "Success with generated code",
			requestBody: `{"original_url": "https://example.com"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS\\(.*").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectQuery("INSERT INTO links").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedCode: http.StatusOK,
			expectedBody: `"short_code"`,
		},
		{
			name:        "Success with custom code",
			requestBody: `{"original_url": "https://example.com", "custom_code": "mycode"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS\\(.*").
					WithArgs("mycode").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectQuery("INSERT INTO links").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedCode: http.StatusOK,
			expectedBody: `"short_code":"mycode"`,
		},
		{
			name:         "Invalid URL",
			requestBody:  `{"original_url": "invalid-url"}`,
			mockClosure:  func(mock sqlmock.Sqlmock) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `"error":"Некорректный URL"`,
		},
		{
			name:        "Custom code exists",
			requestBody: `{"original_url": "https://example.com", "custom_code": "taken"}`,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS\\(.*").
					WithArgs("taken").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedCode: http.StatusConflict,
			expectedBody: `"error":"Код уже занят"`,
		},
		{
			name:         "Invalid custom code",
			requestBody:  `{"original_url": "https://example.com", "custom_code": "!@#$"}`,
			mockClosure:  func(mock sqlmock.Sqlmock) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `Код должен содержать`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, db := setupLinkHandler(t)
			defer db.Close()

			tt.mockClosure(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(
				"POST",
				"/api/links",
				strings.NewReader(tt.requestBody),
			)
			c.Set("userID", 1)

			handler.CreateShortLink(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		shortCode      string
		mockClosure    func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:      "Success",
			shortCode: "valid",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, original_url, short_code, created_at FROM links WHERE short_code = \\$1").
					WithArgs("valid").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "user_id", "original_url", "short_code", "created_at"}).
							AddRow(1, 1, "https://example.com", "valid", time.Now()),
					)

				mock.ExpectExec("UPDATE links SET click_count = click_count \\+ 1 WHERE short_code = \\$1").
					WithArgs("valid").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("INSERT INTO click_analytics").
					WithArgs(
						1,
						sqlmock.AnyArg(), // IP
						sqlmock.AnyArg(), // User-Agent
						sqlmock.AnyArg(), // Location
						sqlmock.AnyArg(), // Device
						sqlmock.AnyArg(), // OS
						sqlmock.AnyArg(), // Browser
						sqlmock.AnyArg(), // Timestamp
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			name:      "Link not found",
			shortCode: "invalid",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, original_url, short_code, created_at FROM links WHERE short_code = \\$1").
					WithArgs("invalid").
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mock, db := setupLinkHandler(t)
			defer db.Close()

			tt.mockClosure(mock)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/"+tt.shortCode, nil)
			c.Params = gin.Params{{Key: "short_code", Value: tt.shortCode}}

			handler.Redirect(c)

			t.Logf("Response body: %s", w.Body.String())
			assert.Equal(t, tt.expectedStatus, w.Code)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}
