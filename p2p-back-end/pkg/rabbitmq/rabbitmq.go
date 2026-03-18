package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/pkg/utils"
)

func NewRabbitMQConnection(cfg *configs.Config) (*amqp.Connection, error) {

	serverUrl, err := utils.UrlBuilder("rabbitmq", cfg)
	if err != nil {
		logs.Error("Failed to Build Rabbitmq url", zap.Error(err))
		return nil, err
	}

	conn, err := amqp.Dial(serverUrl)
	if err != nil {
		logs.Error("Failed to connect to RabbitMQ:", zap.Error(err))
		return nil, err
	}

	logs.Info("RabbitMQ has been connected")
	return conn, nil
}

func NewRabbitMQChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		logs.Error("Failed to Build Channel:", zap.Error(err))
		return nil, err
	}
	return ch, nil
}
