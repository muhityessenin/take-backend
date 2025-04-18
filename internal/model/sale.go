package model

import "time"

type Sale struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ItemID     uint      `json:"itemId"`
	Item       Item      `gorm:"foreignKey:ItemID"`
	SoldAt     time.Time `json:"soldAt"`     // дата продажи
	Quantity   int       `json:"quantity"`   // количество
	TotalPrice int       `json:"totalPrice"` // общая сумма
	Customer   string    `json:"customer"`   // кому продано
}
