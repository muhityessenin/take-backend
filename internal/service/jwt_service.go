package service

import (
	"github.com/golang-jwt/jwt/v5"
	"warehouse-backend/pkg/utils"
)

type JWTService struct{}

func NewJWTService() *JWTService {
	return &JWTService{}
}

func (s *JWTService) ValidateToken(token string) (jwt.MapClaims, error) {
	return utils.ValidateToken(token)
}
