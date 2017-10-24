package model

import (
	_ "github.com/lib/pq"
)

type Food struct {
	Name      string
	Delicious bool
}

type DeliciousFood Food

type Lunch struct {
	Date  string
	Foods []Food
}
