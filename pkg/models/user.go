package models

type User struct {
	User_id int    `json:"user_id"`
	Phone   string `json:"phone"`
	Pass    string `json:"pass"`
	Dostup  int    `json:"dostup"`
	Level   int    `json:"level"`
	Exp     int    `json:"exp"`
	Balance int    `json:"balance"`
}
