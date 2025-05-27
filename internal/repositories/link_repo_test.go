package repositories_test

import (
	"testing"
	"url-short/internal/models"
	"url-short/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestLinkRepository_CreateLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	repo := repositories.NewLinkRepository(db)
	link := &models.Link{
		UserID:      1,
		OriginalURL: "https://example.com",
		ShortCode:   "test123",
	}

	mock.ExpectQuery("INSERT INTO links").
		WithArgs(link.UserID, link.OriginalURL, link.ShortCode).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.CreateLink(link)
	assert.NoError(t, err)
	assert.Equal(t, 1, link.ID)
}

func TestLinkRepository_FindByShortCode(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewLinkRepository(db)
	expectedLink := &models.Link{
		ID:          1,
		UserID:      1,
		OriginalURL: "https://example.com",
		ShortCode:   "test123",
	}

	mock.ExpectQuery("SELECT id, user_id, original_url, short_code, created_at FROM links WHERE short_code = ?").
		WithArgs("test123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "original_url", "short_code", "created_at"}).
			AddRow(expectedLink.ID, expectedLink.UserID, expectedLink.OriginalURL, expectedLink.ShortCode, expectedLink.CreatedAt))

	link, err := repo.FindByShortCode("test123")
	assert.NoError(t, err)
	assert.Equal(t, expectedLink, link)
}
