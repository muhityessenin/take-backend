package repo

import (
	"gorm.io/gorm"
	"log"
	"warehouse-backend/internal/model"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) CreateUser(user *model.User) error {
	log.Println("SAVING TO DB: username:", user.Username)
	log.Println("SAVING TO DB: password:", user.Password)
	return r.DB.Create(user).Error
}

func (r *UserRepo) GetUserByUserName(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	return &user, err
}
