package utils_test

import (
	"testing"
	"url-short/internal/repositories"
	"url-short/internal/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUniqueShortCode(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	linkRepo := repositories.NewLinkRepository(db)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	code, err := utils.GenerateUniqueShortCode(linkRepo)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(code))
}

func TestGenerateRandomCode(t *testing.T) {
	code, err := utils.GenerateRandomCode(8)
	assert.NoError(t, err)
	assert.Equal(t, 8, len(code))
}
