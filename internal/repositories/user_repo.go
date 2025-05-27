package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"url-short/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id
    `
	err := r.DB.QueryRow(query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)

	if err != nil {
		return errors.New("ошибка при создании пользователя")
	}
	return nil
}

var ErrUserNotFound = errors.New("пользователь не найден")

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.DB.QueryRow(
		"SELECT id, username, email, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}
	return user, nil
}
