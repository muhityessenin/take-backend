package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	_ "os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"warehouse-backend/internal/model"
	"warehouse-backend/internal/repo"
)

type ItemHandler struct {
	Repo *repo.ItemRepository
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{
		Repo: repo.NewItemRepository(db),
	}
}

func uploadToCloudflare(file multipart.File, filename string) (string, error) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, _ := writer.CreateFormFile("file", filename)
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.cloudflare.com/client/v4/accounts/e39dcb277e03d5eacfdfad578343290d/images/v1", &b)
	req.Header.Set("Authorization", "Bearer I5csC_rU3Su8IyluZ4g0Ruy1OR8Eb5r0GluC6SnW")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			Variants []string `json:"variants"`
		} `json:"result"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Println("Cloudflare response body:", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Println("Ошибка при парсинге JSON:", err)
		return "", err
	}

	if !result.Success || len(result.Result.Variants) == 0 {
		log.Println("Upload to Cloudflare failed or empty variants:", result)
		return "", errors.New("cloudflare upload failed")
	}

	log.Println("Cloudflare image uploaded:", result.Result.Variants[0])
	return result.Result.Variants[0], nil
}

func (h *ItemHandler) AddItem(c *gin.Context) {
	name := c.PostForm("name")
	partNumber := c.PostForm("partNumber")
	brand := c.PostForm("brand")
	modelName := c.PostForm("model")
	stock, _ := strconv.Atoi(c.PostForm("stock"))
	price, _ := strconv.Atoi(c.PostForm("price"))
	wholesalePrice, _ := strconv.Atoi(c.PostForm("wholesalePrice"))

	item := model.Item{
		Name:           name,
		PartNumber:     partNumber,
		Brand:          brand,
		Model:          modelName,
		Stock:          stock,
		Price:          price,
		WholesalePrice: wholesalePrice,
	}

	form, _ := c.MultipartForm()
	files := form.File["images"]

	for _, fileHeader := range files {
		file, _ := fileHeader.Open()
		url, err := uploadToCloudflare(file, fileHeader.Filename)
		file.Close()
		if err != nil {
			log.Println("Ошибка при загрузке фото:", err)
		} else {
			item.Images = append(item.Images, model.ItemImage{URL: url})
		}

	}

	if err := h.Repo.AddItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ItemHandler) GetItems(c *gin.Context) {
	brand := c.Query("brand")

	var (
		items []model.Item
		err   error
	)

	if brand != "" {
		items, err = h.Repo.GetItemsByBrand(brand)
	} else {
		items, err = h.Repo.GetAllItems()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список товаров"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ItemHandler) MakeSale(c *gin.Context) {
	var req struct {
		ItemID   uint   `json:"itemId"`
		Quantity int    `json:"quantity"`
		Customer string `json:"customer"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	sale, err := h.Repo.MakeSale(req.ItemID, req.Quantity, req.Customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sale)
}

func (h *ItemHandler) GetTodaySales(c *gin.Context) {
	sales, err := h.Repo.GetTodaySales()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить продажи за сегодня"})
		return
	}
	c.JSON(http.StatusOK, sales)
}

func (h *ItemHandler) GetTop5BestSellers(c *gin.Context) {
	top, err := h.Repo.GetTop5BestSellers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить топ продаж"})
		return
	}
	c.JSON(http.StatusOK, top)
}

func (h *ItemHandler) GetSales(c *gin.Context) {
	brand := c.Query("brand")

	var (
		sales []model.Sale
		err   error
	)

	if brand != "" {
		sales, err = h.Repo.GetSalesByBrand(brand)
	} else {
		sales, err = h.Repo.GetTodaySales() // или можешь сделать GetAllSales()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить продажи"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (h *ItemHandler) UpdateItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	item := make(map[string]interface{})
	if err := c.Bind(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format"})
		return
	}

	form, _ := c.MultipartForm()
	files := form.File["images"]
	var images []model.ItemImage

	for _, fileHeader := range files {
		file, _ := fileHeader.Open()
		defer file.Close()
		url, err := uploadToCloudflare(file, fileHeader.Filename)
		if err == nil {
			images = append(images, model.ItemImage{URL: url})
		}
	}

	if len(images) > 0 {
		item["images"] = images
	}

	updatedItem, err := h.Repo.UpdateItem(uint(id), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedItem)
}
