package repositories

import (
	"database/sql"
	"errors"
	"url-short/internal/models"
)

type AnalyticRepository struct {
	DB *sql.DB
}

func NewAnalyticRepository(db *sql.DB) *AnalyticRepository {
	return &AnalyticRepository{DB: db}
}

func (r *AnalyticRepository) SaveClick(click *models.ClickAnalytic) error {
	query := `
        INSERT INTO click_analytics 
        (link_id, ip_address, user_agent, location, device_type, os, browser, clicked_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.DB.Exec(
		query,
		click.LinkID,
		click.IPAddress,
		click.UserAgent,
		click.Location,
		click.DeviceType,
		click.OS,
		click.Browser,
		click.ClickedAt,
	)
	return err
}

func (r *AnalyticRepository) GetAnalytics(linkID int) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var total int
	err := r.DB.QueryRow(
		"SELECT COUNT(*) FROM click_analytics WHERE link_id = $1",
		linkID,
	).Scan(&total)
	if err != nil {
		return nil, errors.New("ошибка получения общего числа кликов")
	}
	result["total_clicks"] = total

	devices := make(map[string]int)
	rows, err := r.DB.Query(
		"SELECT device_type, COUNT(*) FROM click_analytics WHERE link_id = $1 GROUP BY device_type",
		linkID,
	)
	if err != nil {
		return nil, errors.New("ошибка получения данных по устройствам")
	}
	defer rows.Close()

	for rows.Next() {
		var device string
		var count int
		if err := rows.Scan(&device, &count); err != nil {
			continue
		}
		devices[device] = count
	}
	result["devices"] = devices

	browsers := make(map[string]int)
	rows, err = r.DB.Query(
		"SELECT browser, COUNT(*) FROM click_analytics WHERE link_id = $1 GROUP BY browser",
		linkID,
	)
	if err != nil {
		return nil, errors.New("ошибка получения данных по браузерам")
	}
	defer rows.Close()

	for rows.Next() {
		var browser string
		var count int
		if err := rows.Scan(&browser, &count); err != nil {
			continue
		}
		browsers[browser] = count
	}
	result["browsers"] = browsers

	locations := make(map[string]int)
	rows, err = r.DB.Query(
		"SELECT location, COUNT(*) FROM click_analytics WHERE link_id = $1 GROUP BY location",
		linkID,
	)
	if err != nil {
		return nil, errors.New("ошибка получения данных по локациям")
	}
	defer rows.Close()

	for rows.Next() {
		var location string
		var count int
		if err := rows.Scan(&location, &count); err != nil {
			continue
		}
		locations[location] = count
	}
	result["locations"] = locations

	return result, nil
}
