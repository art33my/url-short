package repositories

import (
	"database/sql"
	"url-short/internal/models"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context, user *models.User) error {
	db := c.MustGet("db").(*sql.DB)
	query := `
        INSERT INTO users (username, email, password_hash, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id
    `
	err := db.QueryRow(query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)
	return err
}

func FindUserByEmail(c *gin.Context, email string) (*models.User, error) {
	db := c.MustGet("db").(*sql.DB)
	query := "SELECT id, username, email, password_hash FROM users WHERE email = $1"
	row := db.QueryRow(query, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash); err != nil {
		return nil, err
	}
	return &user, nil
}
