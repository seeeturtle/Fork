package model

import (
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

func GetDeliciousFoods(db *sql.DB, t *time.Time) ([]string, error) {
	var delicious_foods []string
	err := db.QueryRow("SELECT delicious_foods lunch WHERE date = $1", t.Format("20060102")).Scan(&delicious_foods)
	if err != nil {
		return []string{}, err
	}
	return delicious_foods, nil
}

func GetFoods(db *sql.DB, t *time.Time) ([]string, error) {
	var foods []string
	err := db.QueryRow("SELECT foods FROM lunch WHERE date = $1", t.Format("20060102")).Scan(&foods)
	if err != nil {
		return []string{}, err
	}
	return foods, nil
}
