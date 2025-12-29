package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
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

func applyJoins(condition *gorm.DB, optional map[string]interface{}, fieldTableMapping map[string]string) *gorm.DB {
	joinedTables := make(map[string]bool) // Track joined tables

	if utils.NeedsJoin(optional, fieldTableMapping, "department_entities") && !joinedTables["department_entities"] {
		condition = condition.Joins("LEFT JOIN department_entities ON department_entities.campaign_id = department_id")
		joinedTables["department_entities"] = true
	}
	return condition
}

func(r userRepositoryDB) GetAll(optional map[string]interface{}, ctx context.Context, offset, size int) ([]models.UserEntity,int,error){

	var users []models.UserEntity
	var totalRecords64 int64

	fieldTableMapping := map[string]string{

		"username": "user_entities.username",
		"email": "user_entities.email",
		"roles":"user_entities.roles",
		"department_name":"department_entities.name",
		"manager_id":"department_entities.manager_id",	
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	condition := r.db.WithContext(ctx)
	condition = applyJoins(condition, optional, fieldTableMapping)

	hasDepartmentFilters := false
	DepartmentFilterConditions := ""
	var DepartmentFilterValues []interface{}

	for field, value := range optional {
		if column, ok := fieldTableMapping[field]; ok && value != nil {
			switch {
			case strings.HasPrefix(column, "department_entities"):
				// กรอง car_budgets
				hasDepartmentFilters = true

				// กรองกรณีอื่น ๆ ของ car_budgets
				if DepartmentFilterConditions != "" {
					DepartmentFilterConditions += " AND "
				}
				DepartmentFilterConditions += fmt.Sprintf("%s = ?", column)
				DepartmentFilterValues = append(DepartmentFilterValues, value)

			default:
				// ใช้ AddCondition สำหรับเงื่อนไขอื่น ๆ
				condition = utils.AddCondition(condition, value, column)
			}
		}
	}

	if err := condition.Model(&models.DepartmentEntity{}).Group("user_entities.id").Count(&totalRecords64).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting user entities: %w", err)
	}

	query := condition.Distinct().Offset(offset).Limit(size)
	if hasDepartmentFilters {
		query = query.Preload("Car", func(db *gorm.DB) *gorm.DB {
			return db.Where(DepartmentFilterConditions, DepartmentFilterValues...)
		})
	} else {
		query = query.Preload("Car")
	}

	

	// คิวรีข้อมูล
	if err := query.Order("id ASC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("error getting campaign: %w", err)
	}

	return users, int(totalRecords64), nil
}