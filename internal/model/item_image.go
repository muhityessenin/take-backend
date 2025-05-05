package model

type ItemImage struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	ItemID uint   `gorm:"index" json:"itemId"`
	URL    string `json:"url"`
}
