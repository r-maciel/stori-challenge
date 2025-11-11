package handlers

import (
	"errors"
	"net/http"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
)

// CreateErrorResponse maps an application error to an HTTP response using a generic envelope.
// It uses the AppError code/message when available; otherwise falls back to a generic 500.
func CreateErrorResponse(c *gin.Context, err error, details interface{}) {
	var ae *shared.AppError
	if errors.As(err, &ae) {
		status := http.StatusInternalServerError
		switch ae.Kind {
		case shared.BadRequestKind:
			status = http.StatusBadRequest
		case shared.ConflictKind:
			status = http.StatusConflict
		case shared.NotFoundKind:
			status = http.StatusNotFound
		default:
			status = http.StatusInternalServerError
		}
		c.JSON(status, responses.ErrorEnvelope{
			Error: responses.ErrorBody{
				Code:    ae.Code,
				Message: ae.Msg,
				Details: details,
			},
		})
		return
	}

	// Fallback for non-AppError errors
	c.JSON(http.StatusInternalServerError, responses.ErrorEnvelope{
		Error: responses.ErrorBody{
			Code:    "internal_error",
			Message: "internal error",
			Details: details,
		},
	})
}
