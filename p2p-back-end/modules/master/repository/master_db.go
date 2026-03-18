package repository

import (
	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
)

type masterRepositoryDB struct {
	db *gorm.DB
}

func NewMasterRepositoryDB(db *gorm.DB) models.MasterRepository {
	return &masterRepositoryDB{db: db}
}
