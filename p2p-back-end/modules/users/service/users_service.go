package service

import (
	"context"
	"fmt"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/errs"
	"p2p-back-end/pkg/utils"
)

type usersService struct {
	userRepo       models.UserRepository
	sourceUserRepo models.SourceUserRepository
	producerSrv    models.ProducerService
	masterRepo     models.MasterRepository
	deptSrv        models.DepartmentService
}

func NewUsersService(
	userRepo models.UserRepository,
	sourceUserRepo models.SourceUserRepository,
	producerSrv models.ProducerService,
	masterRepo models.MasterRepository,
	deptSrv models.DepartmentService,
) models.UsersService {
	return &usersService{
		userRepo,
		sourceUserRepo,
		producerSrv,
		masterRepo,
		deptSrv,
	}
}

func (s *usersService) SyncAllUsersData(ctx context.Context) error {
	logs.Info("⏳ Start Syncing Master Users Data....")

	// Removed Repair Department call as we now resolve UUIDs directly

	err := utils.BatchSync[models.CentralUser, uint](
		ctx,
		0,
		1000,
		func(ctx context.Context, lastID uint, limit int) ([]models.CentralUser, error) {
			return s.sourceUserRepo.GetUsers(ctx, lastID, limit)
		},
		s.saveUsers,
		func(item models.CentralUser) uint {
			return item.UserID
		},
	)

	if err != nil {
		logs.Warn(err.Error())
		return errs.NewDataSyncError("Failed to synchronize users data")
	}

	return nil
}

func (s *usersService) saveUsers(ctx context.Context, data []models.CentralUser) error {
	var users []models.UserEntity

	for _, u := range data {
		entity := utils.SourceUserToUser(&u)
		
		// Resolve Master UUID associations
		if u.CompanyID > 0 {
			entity.CompanyID, _ = s.masterRepo.FindCompanyUUID(ctx, u.CompanyID)
		}
		if u.DepartmentID > 0 {
			entity.DepartmentID, _ = s.masterRepo.FindDeptUUID(ctx, u.DepartmentID)
		}
		if u.SectionID > 0 {
			entity.SectionID, _ = s.masterRepo.FindSectionUUID(ctx, u.SectionID)
		}
		if u.PositionID > 0 {
			entity.PositionID, _ = s.masterRepo.FindPositionUUID(ctx, u.PositionID)
		}

		users = append(users, *entity)
	}

	changedRows, err := s.userRepo.SyncUsers(ctx, users)
	if err != nil {
		logs.Warn(err.Error())
		return err
	}

	if len(changedRows) > 0 {
		logs.Infof("Detected %d changed users in batch, sending to RabbitMQ...", len(changedRows))

		var usersEvent []events.UserEvent
		for _, row := range changedRows {
			event := utils.UserToUserChangeEvent(&row)
			if event != nil {
				usersEvent = append(usersEvent, *event)
			}
		}

		if len(usersEvent) > 0 {
			syncEvent := &events.MessageUserEvent{
				Users: usersEvent,
			}

			if err := s.producerSrv.UserChange(syncEvent); err != nil {
				logs.Warnf("Producer Error: %v", err)
			}
		}
	}

	return nil
}

func (s *usersService) SyncUserByUserName(ctx context.Context, username string) (*models.UserResponse, error) {

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return utils.UsersToUsersResponse(user), nil
	}

	sourceUser, err := s.sourceUserRepo.FindByUsername(ctx, username)
	if err != nil {
		logs.Warn(err.Error())
		return nil, errs.NewDataSyncError("user not found in source")
	}

	converted := utils.SourceUserToUser(sourceUser)
	
	if sourceUser.CompanyID > 0 {
		converted.CompanyID, _ = s.masterRepo.FindCompanyUUID(ctx, sourceUser.CompanyID)
	}
	if sourceUser.DepartmentID > 0 {
		converted.DepartmentID, _ = s.masterRepo.FindDeptUUID(ctx, sourceUser.DepartmentID)
	}
	if sourceUser.SectionID > 0 {
		converted.SectionID, _ = s.masterRepo.FindSectionUUID(ctx, sourceUser.SectionID)
	}
	if sourceUser.PositionID > 0 {
		converted.PositionID, _ = s.masterRepo.FindPositionUUID(ctx, sourceUser.PositionID)
	}
	syncUsers, err := s.userRepo.SyncUsers(ctx, []models.UserEntity{*converted})
	if err != nil || len(syncUsers) == 0 {
		return nil, errs.NewDataSyncError("failed to sync user to local db")
	}

	finalUser := &syncUsers[0]

	go s.sendUserChangeEvent(*finalUser)

	// Removed Repair Department call

	return utils.UsersToUsersResponse(finalUser), nil
}

// buildDeptMapping removed - using DB individual lookups now

func (s *usersService) sendUserChangeEvent(u models.UserEntity) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error(fmt.Sprintf("Producer Panic: %v", r))
		}
	}()

	reqSync := utils.UserToUserChangeEvent(&u)
	event := &events.MessageUserEvent{
		Users: []events.UserEvent{*reqSync},
	}

	if err := s.producerSrv.UserChange(event); err != nil {
		logs.Warn(fmt.Sprintf("Async Producer Error: %v", err))
	}
}

// แบบนี้เป็น best และรีดประสิทธิภาพดีกว่า
func (s *usersService) BroadcastAllLocalUsers(ctx context.Context) error {
	return utils.BatchSync[models.UserEntity, uint](
		ctx,
		0,
		1000,
		func(ctx context.Context, lastID uint, limit int) ([]models.UserEntity, error) {
			return s.userRepo.GetUsers(ctx, lastID, limit)
		},
		func(ctx context.Context, data []models.UserEntity) error {
			if len(data) == 0 {
				return nil
			}
			userEvents := make([]events.UserEvent, len(data))
			for i := range data {
				event := utils.UserToUserChangeEvent(&data[i])
				if event != nil {
					userEvents[i] = *event
				}
			}
			syncEvent := &events.MessageUserEvent{
				Users: userEvents,
			}
			return s.producerSrv.UserChange(syncEvent)
		},
		func(item models.UserEntity) uint {
			return item.CentralID
		},
	)
}

func (s *usersService) SyncUsersFromEvent(ctx context.Context, usersEvent []events.UserEvent) error {
	var users []models.UserEntity
	for _, ue := range usersEvent {
		eEvent := ue
		entity := utils.EventUserToUsers(&eEvent)

		if ue.CompanyID > 0 {
			entity.CompanyID, _ = s.masterRepo.FindCompanyUUID(ctx, ue.CompanyID)
		}
		if ue.DepartmentID > 0 {
			entity.DepartmentID, _ = s.masterRepo.FindDeptUUID(ctx, ue.DepartmentID)
		}
		if ue.SectionID > 0 {
			entity.SectionID, _ = s.masterRepo.FindSectionUUID(ctx, ue.SectionID)
		}
		if ue.PositionID > 0 {
			entity.PositionID, _ = s.masterRepo.FindPositionUUID(ctx, ue.PositionID)
		}
		
		users = append(users, *entity)
	}

	_, err := s.userRepo.SyncUsers(ctx, users)
	return err
}
