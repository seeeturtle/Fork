package model

import (
	_ "github.com/lib/pq"
)

type Food struct {
	Name string
}

type DeliciousFood Food
