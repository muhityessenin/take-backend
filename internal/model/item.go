package model

type Item struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	Name       string      `json:"name"`
	PartNumber string      `json:"partNumber"`
	Brand      string      `json:"brand"`
	Stock      int         `json:"stock"`
	Price      int         `json:"price"`
	Images     []ItemImage `gorm:"foreignKey:ItemID" json:"images"`
	Sales      []Sale      `gorm:"foreignKey:ItemID"`
}
