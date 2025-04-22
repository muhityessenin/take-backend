package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"warehouse-backend/internal/model"
	"warehouse-backend/internal/repo"

	"warehouse-backend/pkg/utils"
)

type UserService struct {
	Repo *repo.UserRepo
}

func NewUserService(r *repo.UserRepo) *UserService {
	return &UserService{Repo: r}
}

func (s *UserService) Register(user *model.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	log.Println("REGISTER username:", user.Username)
	log.Println("REGISTER password hash:", user.Password)
	user.Password = string(hash)
	return s.Repo.CreateUser(user)
}

func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.Repo.GetUserByUserName(username)
	if err != nil {
		return "", errors.New("user not found")
	}

	log.Println("DEBUG: Username:", username)
	log.Println("DEBUG: Input password:", password)
	log.Println("DEBUG: DB hash:", user.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("DEBUG: bcrypt comparison failed:", err) // <== ДОБАВЬ ЭТО
		return "", errors.New("invalid credentials")
	}
	return utils.GenerateJWT(user.ID, user.Username)
}
