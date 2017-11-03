package model

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

const (
	timeForm string = "20060102"
	LocForm  string = "Asia/Seoul"
)

var (
	Lunches *LunchesModel
)

type Food struct {
	Name      string
	delicious bool
}

type Modeler interface {
	Get(...interface{}) (interface{}, error)
	GetDelicious(...interface{}) (interface{}, error)
	New(...interface{}) interface{}
}

type Lunch struct {
	Date  string
	Foods []Food
}

type LunchesModel struct {
	Value []Lunch
	DB    *sql.DB
}

func (l *LunchesModel) New(db *sql.DB) *LunchesModel {
	lunches := &LunchesModel{
		Value: []Lunch{},
		DB:    db,
	}
	return lunches
}

func (l *LunchesModel) Get(startDate, endDate string) ([]Lunch, error) {
	var lunches LunchesModel
	loc, _ := time.LoadLocation(LocForm)
	startTime, _ := time.ParseInLocation(timeForm, startDate, loc)
	endTime, _ := time.ParseInLocation(timeForm, endDate, loc)
	if startDate != endDate {
		for d := startTime; d.Before(endTime) || d.Equal(endTime); d = d.AddDate(0, 0, 1) {
			lunch, err := l.getADay(d)
			if err != nil {
				continue
			}
			lunches.Value = append(lunches.Value, lunch)
		}
		return lunches.Value, nil
	}
	lunch, err := l.getADay(startTime)
	if err != nil {
		return lunches.Value, err
	}
	lunches.Value = append(lunches.Value, lunch)
	return lunches.Value, nil
}

func (l *LunchesModel) GetDelicious(startDate, endDate string) ([]Lunch, error) {
	var deliciousLunches []Lunch

	lunches, err := l.Get(startDate, endDate)

	if err != nil {
		return deliciousLunches, err
	}

	for _, lunch := range lunches {
		var deliciousLunch Lunch
		for _, food := range lunch.Foods {
			if food.delicious {
				deliciousLunch.Foods = append(deliciousLunch.Foods, food)
			}
		}
		deliciousLunches = append(deliciousLunches, deliciousLunch)
	}
	return deliciousLunches, nil
}

func (l *LunchesModel) getADay(d time.Time) (Lunch, error) {
	var (
		lunch Lunch
		foods []Food
	)
	date := d.Format(timeForm)
	var lunchID int
	err := l.DB.QueryRow("SELECT lunch_id FROM lunches WHERE date=$1", date).Scan(&lunchID)

	if err != nil {
		return lunch, err
	}

	query := `
	SELECT f.food_name, f.delicious
	FROM foods as f
	NATURAL JOIN
	(SELECT food_id FROM lunches_foods
	WHERE lunch_id=$1) as food_ids
	`
	rows, err := l.DB.Query(query, lunchID)

	if err != nil {
		return lunch, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			food      Food
			foodName  string
			delicious bool
		)

		err := rows.Scan(&foodName, &delicious)

		if err != nil {
			return lunch, err
		}

		food.Name = foodName
		food.delicious = delicious
		foods = append(foods, food)
	}

	err = rows.Err()
	if err != nil {
		return lunch, err
	}

	lunch.Date = date
	lunch.Foods = foods

	return lunch, nil
}
