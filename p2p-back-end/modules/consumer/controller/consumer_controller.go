package controller

import (
	"fmt"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

type MessageProcessor func(body []byte) error

type consumerController struct {
	consumerService models.ConsumerService
	processors      map[string]MessageProcessor
}

func NewConsumerController(consumerService models.ConsumerService) models.ConsumerController {
	h := &consumerController{
		consumerService: consumerService,
	}

	h.processors = map[string]MessageProcessor{
		"autocorp.company.change":    h.consumerService.ProcessCompanyChange,
		"autocorp.department.change": h.consumerService.ProcessDepartmentChange,
		"autocorp.section.change":    h.consumerService.ProcessSectionChange,
		"autocorp.position.change":   h.consumerService.ProcessPositionChange,
		"autocorp.user.change":       h.consumerService.ProcessUserChange,
	}

	return h
}

func (h *consumerController) HandleMessage(d amqp.Delivery) {
	processor, ok := h.processors[d.RoutingKey]
	if !ok {
		logs.Warn(fmt.Sprintf("No processor found for key: %s", d.RoutingKey))
		if err := d.Ack(false); err != nil {
			logs.Error(fmt.Sprintf("Failed to acknowledge message: %s", err.Error()))
		}
		return
	}

	if err := processor(d.Body); err != nil {
		if isFatalError(err) {
			logs.Warn(fmt.Sprintf("Fatal error (Discarding): %v", err))
			if err := d.Ack(false); err != nil {
				logs.Error(fmt.Sprintf("Failed to acknowledge message: %s", err.Error()))
			}
		} else {
			logs.Warn(fmt.Sprintf("Transient error (Retrying): %v", err))
			if errNack := d.Nack(false, true); errNack != nil {
				logs.Error(fmt.Sprintf("Failed to Nack message: %v", errNack))
			}
		}
		return
	}

	if err := d.Ack(false); err != nil {
		logs.Error(fmt.Sprintf("Failed to acknowledge message: %s", err.Error()))
	}
}

func isFatalError(err error) bool {
	if strings.Contains(err.Error(), "unmarshal") || strings.Contains(err.Error(), "invalid character") {
		return true
	}
	return false
}
