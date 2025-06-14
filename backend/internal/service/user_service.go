package service

import (
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := s.db.Preload("Tenant").First(&user, id).Error
	return &user, err
}
