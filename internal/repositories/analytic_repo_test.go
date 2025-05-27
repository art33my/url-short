package repositories_test

import (
	"testing"
	"url-short/internal/models"
	"url-short/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticRepository_SaveClick(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repositories.NewAnalyticRepository(db)
	click := &models.ClickAnalytic{
		LinkID:     1,
		IPAddress:  "127.0.0.1",
		UserAgent:  "test-agent",
		Location:   "localhost",
		DeviceType: "desktop",
		OS:         "Windows",
		Browser:    "Chrome",
	}

	mock.ExpectExec("INSERT INTO click_analytics").
		WithArgs(
			click.LinkID,
			click.IPAddress,
			click.UserAgent,
			click.Location,
			click.DeviceType,
			click.OS,
			click.Browser,
			click.ClickedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.SaveClick(click)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticRepository_GetAnalytics(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	//repo := repositories.NewAnalyticRepository(db)
	linkID := 1

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM click_analytics WHERE link_id = \\$1").
		WithArgs(linkID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	mock.ExpectQuery("SELECT device_type, COUNT\\(\\*\\) FROM click_analytics WHERE link_id = \\$1 GROUP BY device_type").
		WithArgs(linkID).
		WillReturnRows(sqlmock.NewRows([]string{"device_type", "count"}).
			AddRow("mobile", 60).
			AddRow("desktop", 40))

	mock.ExpectQuery("SELECT browser, COUNT\\(\\*\\) FROM click_analytics WHERE link_id = \\$1 GROUP BY browser").
		WithArgs(linkID).
		WillReturnRows(sqlmock.NewRows([]string{"browser", "count"}).
			AddRow("Chrome", 70).
			AddRow("Firefox", 30))

	mock.ExpectQuery("SELECT location, COUNT\\(\\*\\) FROM click_analytics WHERE link_id = \\$1 GROUP BY location").
		WithArgs(linkID).
		WillReturnRows(sqlmock.NewRows([]string{"location", "count"}).
			AddRow("Moscow", 50).
			AddRow("New York", 50))

	//stats, err := repo.GetAnalytics(linkID)
	//assert.NoError(t, err)
	//assert.Equal(t, 100, stats["total_clicks"])
	//assert.Equal(t, map[string]int{"mobile": 60, "desktop": 40}, stats["devices"])
	//assert.Equal(t, map[string]int{"Chrome": 70, "Firefox": 30}, stats["browsers"])
	//assert.Equal(t, map[string]int{"Moscow": 50, "New York": 50}, stats["locations"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
