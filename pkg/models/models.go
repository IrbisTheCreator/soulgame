package models

import (
	"errors"
)

var ErrNoRecord = errors.New("models: подходящей записи не найдено")

type Quests struct {
	Quest_id    int
	Title       string
	Description string
	Price       int
}
