package repositories_test

import (
	"database/sql"
	"testing"
	"url-short/internal/models"
	"url-short/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewUserRepository(db)
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Username, user.Email, user.PasswordHash).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewUserRepository(db)
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	mock.ExpectQuery("INSERT INTO users").
		WillReturnError(sql.ErrNoRows)

	err := repo.Create(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при создании пользователя")
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewUserRepository(db)
	expectedUser := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE email = ?").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.PasswordHash))

	user, err := repo.FindByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewUserRepository(db)

	mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE email = ?").
		WithArgs("wrong@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.FindByEmail("wrong@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "пользователь не найден")
}
