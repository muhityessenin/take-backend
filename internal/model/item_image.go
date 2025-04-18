package model

type ItemImage struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	ItemID uint   `json:"itemId"`
	URL    string `json:"url"`
}
