package repo

import (
	"fmt"
	"gorm.io/gorm"
	"time"
	"warehouse-backend/internal/model"
)

type ItemRepository struct {
	DB *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{DB: db}
}

func (r *ItemRepository) AddItem(item *model.Item) error {
	return r.DB.Create(item).Error
}

func (r *ItemRepository) UpdateStock(id uint, quantity int) (*model.Item, error) {
	var item model.Item
	if err := r.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	if item.Stock < quantity {
		return nil, gorm.ErrInvalidData
	}
	item.Stock -= quantity
	if err := r.DB.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepository) GetAllItems() ([]model.Item, error) {
	var items []model.Item
	err := r.DB.Preload("Images").Find(&items).Error
	return items, err
}
func (r *ItemRepository) GetItemsByBrand(brand string) ([]model.Item, error) {
	var items []model.Item
	err := r.DB.Preload("Images").Where("brand = ?", brand).Find(&items).Error
	return items, err
}

func (r *ItemRepository) MakeSale(itemID uint, quantity int, customer string) (*model.Sale, error) {
	var item model.Item
	if err := r.DB.First(&item, itemID).Error; err != nil {
		return nil, err
	}

	if item.Stock < quantity {
		return nil, fmt.Errorf("недостаточно товара на складе")
	}

	// уменьшаем количество
	item.Stock -= quantity
	if err := r.DB.Save(&item).Error; err != nil {
		return nil, err
	}

	// создаём продажу
	sale := model.Sale{
		ItemID:     itemID,
		Quantity:   quantity,
		TotalPrice: item.Price * quantity,
		Customer:   customer,
		SoldAt:     time.Now(),
	}

	if err := r.DB.Create(&sale).Error; err != nil {
		return nil, err
	}

	return &sale, nil
}

func (r *ItemRepository) GetTodaySales() ([]model.Sale, error) {
	var sales []model.Sale

	today := time.Now().Truncate(24 * time.Hour) // начало дня (00:00)
	err := r.DB.Preload("Item").
		Where("sold_at >= ?", today).
		Order("sold_at desc").
		Find(&sales).Error

	return sales, err
}

func (r *ItemRepository) GetTop5BestSellers() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	oneWeekAgo := time.Now().AddDate(0, 0, -7)

	// Используем raw SQL с join, group by и order
	rows, err := r.DB.Table("sales").
		Select("items.name, items.part_number, SUM(sales.quantity) as total_sold").
		Joins("JOIN items ON sales.item_id = items.id").
		Where("sales.sold_at >= ?", oneWeekAgo).
		Group("items.id, items.name, items.part_number").
		Order("total_sold DESC").
		Limit(5).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, partNumber string
		var totalSold int
		if err := rows.Scan(&name, &partNumber, &totalSold); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"name":       name,
			"partNumber": partNumber,
			"totalSold":  totalSold,
		})
	}

	return results, nil
}

func (r *ItemRepository) GetSalesByBrand(brand string) ([]model.Sale, error) {
	var sales []model.Sale
	err := r.DB.
		Joins("JOIN items ON sales.item_id = items.id").
		Where("items.brand = ?", brand).
		Preload("Item").
		Order("sales.sold_at DESC").
		Find(&sales).Error
	return sales, err
}
