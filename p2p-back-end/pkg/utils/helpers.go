package utils

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func NeedsJoin(optional map[string]interface{}, fieldTableMapping map[string]string, tablePrefix string) bool {
	for field := range optional {
		if column, ok := fieldTableMapping[field]; ok && strings.HasPrefix(column, tablePrefix) {
			return true // ถ้าพบฟิลด์ที่ต้องการ tablePrefix จะ return true
		}
	}
	return false
}

func AddCondition(condition *gorm.DB, value interface{}, query string) *gorm.DB {
	switch v := value.(type) {
	case string:
		// ใช้ ILIKE สำหรับการค้นหา string
		return condition.Where(query+" ILIKE ?", "%"+v+"%")
	case []string:
		// ใช้ WHERE IN สำหรับ slice ของ string (GORM handles slices automatically with IN)
		return condition.Where(query+" IN ?", v)
	default:
		// กรณีค่าอื่นๆ
		return condition.Where(query, value)
	}
}

func AddconditionReqArray(condition *gorm.DB, value interface{}, query string) *gorm.DB {
	switch v := value.(type) {
	case string:
		// ปกติกรณีค้น string ตรงๆ เช่น keyword
		return condition.Where(query+" ILIKE ?", "%"+v+"%")

	case []string:
		// กรณีนี้เลย: string array เช็คกับฟิลด์ string
		return condition.Where(query+" IN ?", v)

	default:
		return condition.Where(query+" = ?", value)
	}
}

func ClearAllCache(rdb redis.Cmdable) error {
	ctx := context.Background()
	status := rdb.FlushAll(ctx)
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}

// /////// auth ///////////
func ConvertInterfaceSliceToStringSlice(slice []interface{}) []string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		// ต้องมั่นใจว่า v สามารถแปลงเป็น string ได้
		strSlice[i] = v.(string)
	}
	return strSlice
}

func GetSafeString(claims map[string]interface{}, key string) string {
	if v, ok := claims[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
