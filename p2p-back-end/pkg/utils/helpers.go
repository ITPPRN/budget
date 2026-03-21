package utils

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

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

// NaturalLess compares strings numerically where possible (e.g. "ADM1" < "ADM10")
func NaturalLess(s1, s2 string) bool {
	i, j := 0, 0
	for i < len(s1) && j < len(s2) {
		r1, r2 := s1[i], s2[j]

		// If both are digits, compare as numbers
		if isDigit(r1) && isDigit(r2) {
			num1, nextI := parseNum(s1, i)
			num2, nextJ := parseNum(s2, j)
			if num1 != num2 {
				return num1 < num2
			}
			i, j = nextI, nextJ
			continue
		}

		if r1 != r2 {
			return r1 < r2
		}
		i++
		j++
	}
	return len(s1) < len(s2)
}

func parseNum(s string, start int) (int, int) {
	end := start
	val := 0
	for end < len(s) && isDigit(s[end]) {
		val = val*10 + int(s[end]-'0')
		end++
	}
	return val, end
}

// GetTimestamp returns current time in YYYYMMDDHHMMSS format
func GetTimestamp() string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(time.Now().Format("2006-01-02 15:04:05"), "-", ""), ":", ""), " ", "")
}

// ToDecimal converts float64 to decimal safely
func ToDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
