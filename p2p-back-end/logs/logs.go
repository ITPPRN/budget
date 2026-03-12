package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Loginit() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = customTimeEncoder
	config.EncoderConfig.StacktraceKey = ""

	var err error
	Logger, err = config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := Logger.Sync(); err != nil {
            handleSyncError(err)
        }
	}()

}

func handleSyncError(err error) {
	// รายการ Error ที่ "ไม่อันตราย" สำหรับการ Sync ลง Terminal/Console
	ignoredErrors := []string{
		"inappropriate ioctl for device",
		"invalid argument",
		"bad file descriptor",
		"enotty",
	}

	for _, msg := range ignoredErrors {
		if containsIgnoreCase(err.Error(), msg) {
			return // เป็น Error ปกติของระบบ terminal ให้ข้ามไปเลย
		}
	}

	// ถ้าหลุดจากข้างบนมา แสดงว่าเป็น Error ที่อาจจะร้ายแรงจริงๆ ให้พิมพ์ออกมาดู
	fmt.Printf("Logger Sync Error (Potential Issue): %v\n", err)
}

func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func Info(message string, fields ...zap.Field) {
	Logger.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	Logger.Debug(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	Logger.Warn(message, fields...)
}

func Error(message interface{}, fields ...zap.Field) {
	switch v := message.(type) {
	case error:
		Logger.Error(v.Error(), fields...)
	case string:
		Logger.Error(v, fields...)
	}
}
func Fatal(message string, fields ...zap.Field) {
	Logger.Fatal(message, fields...)
}

func Infof(format string, args ...interface{}) {
	Logger.Info((fmt.Sprintf(format, args...)))
}
func Warnf(format string, args ...interface{}) {
	Logger.Warn((fmt.Sprintf(format, args...)))
}
func Debugf(format string, args ...interface{}) {
	Logger.Debug((fmt.Sprintf(format, args...)))
}

func Errorf(format string, args ...interface{}) {
	Logger.Error((fmt.Sprintf(format, args...)))
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatal(fmt.Sprintf(format, args...))
}

func LogHttp(c *fiber.Ctx) error {
	Infof("HTTP request - status: %d, method: %s, path: %s, ip: %s",
		c.Response().StatusCode(),
		c.Method(),
		c.Path(),
		c.IP(),
	)
	start := time.Now()

	err := c.Next()

	duration := time.Since(start)

	respStatus := c.Response().StatusCode()
	duration = time.Since(start)

	// 🔥 Capture Request Body for POST/PUT (Great for debugging Frontend mismatches)
	reqBody := ""
	method := c.Method()
	contentType := string(c.Request().Header.ContentType())
	if (method == "POST" || method == "PUT") && !strings.Contains(contentType, "multipart/form-data") {
		reqBody = string(c.Body())
	}


	Infof("HTTP response - status: %d, method: %s, path: %s, body: %s, ip: %s, duration: %s",
		respStatus,
		method,
		c.Path(),
		reqBody,
		c.IP(),
		duration,
	)
	return err
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
