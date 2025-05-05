package repositories

import (
	"database/sql"
	"errors"
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

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := "SELECT id, username, email, password_hash FROM users WHERE email = $1"
	row := r.DB.QueryRow(query, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, errors.New("ошибка базы данных")
	}
	return &user, nil
}
