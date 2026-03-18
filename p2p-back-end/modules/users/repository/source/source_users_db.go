package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

type sourceUsersRepositoryDB struct {
	db *gorm.DB
}

func NewSourceUsersRepositoryDB(db *gorm.DB) models.SourceUserRepository {
	return &sourceUsersRepositoryDB{db: db}
}

func (r *sourceUsersRepositoryDB) GetUsers(ctx context.Context, lastID uint, limit int) ([]models.CentralUser, error) {

	var users []models.CentralUser

	err := r.db.WithContext(ctx).Where("id > ?", lastID).
		Order("id ASC").
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("sourceUserRepo.GetUsers: %w", err)
	}

	if len(users) > 0 {
		logs.Infof("[SOURCE-VERIFY] Connected to Master DB. Fetched %d users. Sample -> ID: %d, Username: %s, CompanyID: %d", len(users), users[0].UserID, users[0].Username, users[0].CompanyID)
	} else {
		logs.Warn("[SOURCE-VERIFY] No users found in Master DB table.")
	}

	return users, nil
}

func (r *sourceUsersRepositoryDB) FindByUsername(ctx context.Context, username string) (*models.CentralUser, error) {
	var user models.CentralUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("sourceUserRepo.FindByUsername: %w", err)
	}
	return &user, nil
}
