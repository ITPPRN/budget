package repository

import (
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type userRepositoryDB struct {
	db *gorm.DB
}

func NewUserRepositoryDB(db *gorm.DB) models.UserRepository {
	return &userRepositoryDB{db: db}
}

func (r userRepositoryDB) IsUserExistByID(id string) (bool, error) {

	var count int64
	if err := r.db.Table("user_entities").Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
