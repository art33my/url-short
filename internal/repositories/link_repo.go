package repositories

import (
	"database/sql"
	"errors"
	"url-short/internal/models"
)

type LinkRepository struct {
	DB *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{DB: db}
}

var ErrLinkNotFound = errors.New("ссылка не найдена")

func (r *LinkRepository) CreateLink(link *models.Link) error {
	query := `
        INSERT INTO links (user_id, original_url, short_code, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id
    `
	err := r.DB.QueryRow(
		query,
		link.UserID,
		link.OriginalURL,
		link.ShortCode,
	).Scan(&link.ID)
	if err != nil {
		return errors.New("ошибка при создании ссылки")
	}
	return nil
}

func (r *LinkRepository) FindByShortCode(shortCode string) (*models.Link, error) {
	query := `
        SELECT id, user_id, original_url, short_code, created_at 
        FROM links 
        WHERE LOWER(short_code) = LOWER($1)  -- Регистронезависимый поиск
    `
	row := r.DB.QueryRow(query, shortCode)
	var link models.Link
	err := row.Scan(
		&link.ID,
		&link.UserID,
		&link.OriginalURL,
		&link.ShortCode,
		&link.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLinkNotFound
	}
	return &link, err
}

func (r *LinkRepository) IncrementClickCount(shortCode string) error {
	_, err := r.DB.Exec("UPDATE links SET click_count = click_count + 1 WHERE short_code = $1", shortCode)
	return err
}

func (r *LinkRepository) IsShortCodeExist(code string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM links WHERE short_code = $1)",
		code,
	).Scan(&exists)
	return exists, err
}
