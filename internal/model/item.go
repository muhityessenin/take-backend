package model

type Item struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	Name           string      `json:"name"`
	PartNumber     string      `json:"partNumber"`
	Brand          string      `json:"brand"`
	Model          string      `json:"model"`
	Stock          int         `json:"stock"`
	Price          int         `json:"price"`
	WholesalePrice int         `json:"wholesalePrice"`
	Images         []ItemImage `gorm:"foreignKey:ItemID" json:"images"`
	Sales          []Sale      `gorm:"foreignKey:ItemID"`
}
