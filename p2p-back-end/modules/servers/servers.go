package servers

// @title P2P Project API
// @version 1.0
// @description This is the API server for the P2P project.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

import (
	"fmt"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/pkg/utils"

	"github.com/robfig/cron/v3"
)

type server struct {
	App       *fiber.App
	Db        *gorm.DB
	Db2       *gorm.DB
	Cfg       *configs.Config
	Redis     *redis.Client
	Keycloak  *gocloak.GoCloak
	Cron      *cron.Cron
	MqConn    *amqp.Connection
	MqChannel *amqp.Channel

	Shd *SharedDeps
}

func NewServer(
	cfg *configs.Config,
	db *gorm.DB,
	db2 *gorm.DB,
	redis *redis.Client,
	keycloak *gocloak.GoCloak,
	mqConn *amqp.Connection,
	mqCh *amqp.Channel,
) *server {
	s := &server{
		App: fiber.New(fiber.Config{
			ReadBufferSize: 32768,
		}),
		Db:        db,
		Db2:       db2,
		Cfg:       cfg,
		Redis:     redis,
		Keycloak:  keycloak,
		MqConn:    mqConn,
		MqChannel: mqCh,
		Cron: cron.New(cron.WithChain(
			cron.Recover(cron.DefaultLogger),
			cron.DelayIfStillRunning(cron.DefaultLogger),
		)),
	}
	s.Shd = initSharedDeps(s)
	return s
}

func (s *server) Start() {
	// --- Phase 3: Graceful Shutdown ---
	defer func() {
		logs.Info("♻️ Graceful Shutdown: Closing resources...")
		if s.MqChannel != nil {
			if err := s.MqChannel.Close(); err != nil {
				logs.Error("Failed to close RabbitMQ channel", zap.Error(err))
			}
		}
		if s.MqConn != nil {
			if err := s.MqConn.Close(); err != nil {
				logs.Error("Failed to close RabbitMQ connection", zap.Error(err))
			}
		}
	}()

	if err := s.Handlers(); err != nil {
		logs.Fatal("Failed to setup handlers", zap.Error(err))
	}

	// --- Phase 2: RabbitMQ Resilience ---
	if s.MqChannel != nil {
		s.startRabbitMQ()
	}

	fiberConnURL, err := utils.UrlBuilder("fiber", s.Cfg)
	if err != nil {
		logs.Fatal("Failed to build fiber URL", zap.Error(err))
	}

	port := s.Cfg.App.Port
	mode := s.Cfg.App.Mode

	logs.Info("Starting Fiber server",
		zap.String("mode", mode),
		zap.String("port", port),
		zap.String("url", fiberConnURL),
	)

	if err := s.App.Listen(fiberConnURL); err != nil {
		logs.Fatal("Fiber server failed to listen", zap.Error(err))
	}
}

func (s *server) startRabbitMQ() {
	go func() {
		for {
			logs.Info("🐰 RabbitMQ: Attempting to setup consumer...")

			// 1. ตั้งค่า QoS
			if err := s.MqChannel.Qos(10, 0, false); err != nil {
				logs.Fatal("Failed to set QoS", zap.Error(err))
			}

			// 2. Declare Queue และ Bind
			q, err := s.MqChannel.QueueDeclare("p2p_service_sync_queue", true, false, false, false, nil)
			if err != nil {
				logs.Error("Failed to declare queue", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			keys := []string{
				"autocorp.company.change",
				"autocorp.department.change",
				"autocorp.section.change",
				"autocorp.position.change",
				"autocorp.user.change",
			}
			for _, key := range keys {
				if err := s.MqChannel.QueueBind(q.Name, key, "authen_event_topic", false, nil); err != nil {
					logs.Fatal(fmt.Sprintf("Failed to bind queue %s to key %s", q.Name, key), zap.Error(err))
				}
			}

			// 3. เริ่มฟังข้อความ
			msgs, err := s.MqChannel.Consume(q.Name, "", false, false, false, false, nil)
			if err != nil {
				logs.Error("Failed to start consuming", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			// ดักจับตอนหลุด
			notifyClose := s.MqChannel.NotifyClose(make(chan *amqp.Error))

			// 4. รัน Consumer loop
			go func() {
				for d := range msgs {
					s.Shd.ConsumerController.HandleMessage(d)
				}
			}()

			// Block รอจนกว่า Channel จะล่ม
			errClose := <-notifyClose
			logs.Warn("❌ RabbitMQ Channel closed. Retrying connection...", zap.Error(errClose))

			s.reconnectChannel()
		}
	}()
}

func (s *server) reconnectChannel() {
	for {
		if s.MqConn == nil || s.MqConn.IsClosed() {
			logs.Info("🔄 RabbitMQ: Connection is closed, redialing...")

			newConn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
				s.Cfg.RabbitMQ.Username, s.Cfg.RabbitMQ.Password,
				s.Cfg.RabbitMQ.Host, s.Cfg.RabbitMQ.Port,
				s.Cfg.RabbitMQ.VHost))

			if err != nil {
				logs.Error("❌ Failed to reconnect RabbitMQ, retrying in 5s...", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
			s.MqConn = newConn
		}

		ch, err := s.MqConn.Channel()
		if err == nil {
			s.MqChannel = ch
			logs.Info("✅ RabbitMQ: Channel reopened successfully")
			return
		}

		logs.Error("❌ Failed to reopen channel, retrying in 5s...", zap.Error(err))
		time.Sleep(5 * time.Second)
	}
}
