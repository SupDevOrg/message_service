package services

import (
	"errors"
	"message_service/internal/models"
	"message_service/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepo
}

func NewUserService(userRepo *repositories.UserRepo) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Create(id uint, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	return s.userRepo.Create(id, username)
}

func (s *UserService) Delete(user uint) error {
	if user == 0 {
		return errors.New("invalid user ID")
	}

	exsts, err := s.userRepo.FindByID(user)
	if err != nil {
		return err
	}
	if exsts == nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(user)
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
