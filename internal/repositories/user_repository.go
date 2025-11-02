package repositories

import (
	"gorm.io/gorm"
	"message_service/internal/models"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(username string) (*models.User, error) {
	user := &models.User{
		Username: username,
	}
	err := r.db.Create(&user).Error

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByID(id uint) (*models.User, error) {
	var user models.User

	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Delete(user uint) error {
	return r.db.Delete(&models.User{}, user).Error
}
