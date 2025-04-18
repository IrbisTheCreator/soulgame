package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: подходящей записи не найдено")

type Quests struct {
	Quest_id    int
	Title       string
	Description string
	Price       int
	Active      bool
	Exp         int
	Delet       bool
}

type Shop struct {
	Shop_id     int
	Title       string
	Description string
	Count       int
	Price       int
}

type Adm struct {
	Take_id  int
	Phone    string
	Title    string
	Now      int
	User_id  int
	Price    int
	Exp      int
	Quest_id int
}

type HistoryBought struct {
	Purc_id      int
	Phone        string
	PurchaseDate time.Time
	ItemTitle    string
	ClubName     string
	IsIssued     bool
}

type Log struct {
	Log_id int
	Phone  string
	Title  string
	Time   time.Time
}
