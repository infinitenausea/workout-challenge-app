package models

type Exercise struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	IsCustom bool   `json:"is_custom"`
}
