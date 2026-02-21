package controller

import (
	//"errors"
	//"fmt"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/errs"
)

func responseWithError(c *fiber.Ctx, err error) error {
	switch e := err.(type) {
	case errs.AppError:
		return c.Status(e.Code).JSON(
			models.ResponseError{
				Message:    e.Message,
				Status:     e.Status,
				StatusCode: e.Code,
			},
		)
	case error:
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.ResponseError{
				Message:    e.Error(),
				Status:     fiber.ErrInternalServerError.Message,
				StatusCode: fiber.ErrInternalServerError.Code,
			},
		)
	}
	return nil

}

func badReqErrResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(
		models.ResponseError{
			Message:    message,
			Status:     fiber.ErrBadRequest.Message,
			StatusCode: fiber.ErrBadRequest.Code,
		},
	)
}

func forbiddenErrResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(
		models.ResponseError{
			Message:    message,
			Status:     fiber.ErrForbidden.Message,
			StatusCode: fiber.ErrForbidden.Code,
		},
	)
}

func responseSuccess(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(
		models.ResponseData{
			Message:    "Succeed",
			Status:     "OK",
			StatusCode: fiber.StatusOK,
			Data:       data,
		},
	)
}
