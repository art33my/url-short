package repositories

import (
	"database/sql"
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
        INSERT INTO click_analytics (
            link_id, 
            ip_address, 
            user_agent, 
            location, 
            device_type, 
            os, 
            browser
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
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
	)
	return err
}
func (r *AnalyticRepository) GetAnalytics(linkID int) ([]models.ClickAnalytic, error) {
	query := `
        SELECT 
            ip_address, 
            location, 
            device_type, 
            os, 
            browser, 
            clicked_at 
        FROM click_analytics 
        WHERE link_id = $1
    `

	rows, err := r.DB.Query(query, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []models.ClickAnalytic
	for rows.Next() {
		var ca models.ClickAnalytic
		err := rows.Scan(
			&ca.IPAddress,
			&ca.Location,
			&ca.DeviceType,
			&ca.OS,
			&ca.Browser,
			&ca.ClickedAt,
		)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, ca)
	}

	return analytics, nil
}
