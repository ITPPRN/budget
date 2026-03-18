package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

type MessageProcessor func(body []byte) error

type consumerController struct {
	userService   models.UsersService
	masterService models.MasterService
	processors    map[string]MessageProcessor
}

func NewConsumerController(userService models.UsersService, mu models.MasterService) models.ConsumerController {
	h := &consumerController{
		userService:   userService,
		masterService: mu,
	}

	// ลงทะเบียน Handlers ไว้ที่นี่ที่เดียว
	h.processors = map[string]MessageProcessor{
		"autocorp.company.change":    h.processSyncCompany,
		"autocorp.department.change": h.processSyncDepartment,
		"autocorp.section.change":    h.processSyncSection,
		"autocorp.position.change":   h.processSyncPosition,
		"autocorp.user.change":       h.processSyncUsers,
	}

	return h
}

func (h *consumerController) HandleMessage(d amqp.Delivery) {

	processor, ok := h.processors[d.RoutingKey]
	if !ok {
		logs.Warn("No processor found for routing key", zap.String("routing_key", d.RoutingKey))
		if err := d.Ack(false); err != nil {
			logs.Error("Failed to acknowledge message", zap.Error(err))
		}
		return
	}

	if err := processor(d.Body); err != nil {
		if isFatalError(err) {
			logs.Warn("Fatal processing error (Discarding)", zap.Error(err), zap.String("routing_key", d.RoutingKey))
			if err := d.Ack(false); err != nil {
				logs.Error("Failed to acknowledge message", zap.Error(err))
			}
		} else {
			logs.Warn("Transient processing error (Retrying)", zap.Error(err), zap.String("routing_key", d.RoutingKey))
			if errNack := d.Nack(false, true); errNack != nil {
				logs.Error("Failed to Nack message", zap.Error(errNack))
			}
		}
		return
	}
	if err := d.Ack(false); err != nil {
		logs.Error("Failed to acknowledge message", zap.Error(err))
	}
}

func isFatalError(err error) bool {
	if strings.Contains(err.Error(), "unmarshal") || strings.Contains(err.Error(), "invalid character") {
		return true
	}
	return false
}

func (h *consumerController) processSyncCompany(body []byte) error {
	var payload events.MessageCompaniesEvent

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal company: %w", err)
	}
	count := len(payload.Companies)

	if err := h.masterService.SyncCompaniesFromEvent(context.Background(), payload.Companies); err != nil {
		return fmt.Errorf("failed to sync companies: %w", err)
	}

	logs.Info("Sync completed",
		zap.String("module", "company"),
		zap.Int("count", count),
	)

	return nil
}

func (h *consumerController) processSyncDepartment(body []byte) error {
	var payload events.MessageDepartmentEvent

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal department: %w", err)
	}
	count := len(payload.Departments)

	if err := h.masterService.SyncDepartmentsFromEvent(context.Background(), payload.Departments); err != nil {
		return fmt.Errorf("failed to sync departments: %w", err)
	}

	logs.Info("Sync completed",
		zap.String("module", "department"),
		zap.Int("count", count),
	)

	return nil
}

func (h *consumerController) processSyncSection(body []byte) error {
	var payload events.MessageSectionEvent

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal section: %w", err)
	}
	count := len(payload.Sections)

	if err := h.masterService.SyncSectionsFromEvent(context.Background(), payload.Sections); err != nil {
		return fmt.Errorf("failed to sync sections: %w", err)
	}

	logs.Info("Sync completed",
		zap.String("module", "section"),
		zap.Int("count", count),
	)

	return nil
}

func (h *consumerController) processSyncPosition(body []byte) error {
	var payload events.MessagePositionEvent

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal position: %w", err)
	}
	count := len(payload.Positions)

	if err := h.masterService.SyncPositionsFromEvent(context.Background(), payload.Positions); err != nil {
		return fmt.Errorf("failed to sync positions: %w", err)
	}

	logs.Info("Sync completed",
		zap.String("module", "position"),
		zap.Int("count", count),
	)

	return nil
}

func (h *consumerController) processSyncUsers(body []byte) error {
	var payload events.MessageUserEvent

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal users: %w", err)
	}
	count := len(payload.Users)

	if err := h.userService.SyncUsersFromEvent(context.Background(), payload.Users); err != nil {
		return fmt.Errorf("failed to sync users: %w", err)
	}

	logs.Info("Sync completed",
		zap.String("module", "users"),
		zap.Int("count", count),
	)

	return nil
}