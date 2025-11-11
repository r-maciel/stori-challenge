package handlers

import (
	"net/http"
	"strconv"
	"time"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/infrastructure/http/validators"
	"stori-challenge/internal/ports/services"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	Service services.BalanceService
}

func NewBalanceHandler(svc services.BalanceService) *BalanceHandler {
	return &BalanceHandler{Service: svc}
}

// GetBalance
// @Summary      Get user balance summary within an optional time range
// @Description  Returns balance, total_debits and total_credits for a user. Query params from/to must be RFC3339 with Z.
// @Tags         users
// @Produce      json
// @Param        user_id   path      int     true  "User ID"
// @Param        from      query     string  false "RFC3339 with Z lower bound"
// @Param        to        query     string  false "RFC3339 with Z upper bound"
// @Success      200  {object}  responses.BalanceResponse
// @Failure      400  {object}  responses.ErrorEnvelope
// @Failure      404  {object}  responses.ErrorEnvelope
// @Router       /users/{user_id}/balance [get]
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		CreateErrorResponse(c, shared.NewBadRequest("invalid_user_id", "user_id must be a positive integer", nil), nil)
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	now := time.Now().UTC()

	from, to, verr := validators.ParseAndValidateTimeRange(fromStr, toStr, now)
	if verr != nil {
		CreateErrorResponse(c, verr, nil)
		return
	}

	bal, deb, cred, svcErr := h.Service.GetBalance(c.Request.Context(), userID, from, to)
	if svcErr != nil {
		CreateErrorResponse(c, svcErr, nil)
		return
	}

	// Convert to float64 for response
	balF, _ := bal.Float64()
	debF, _ := deb.Float64()
	credF, _ := cred.Float64()

	c.JSON(http.StatusOK, responses.BalanceResponse{
		Balance:      balF,
		TotalDebits:  debF,
		TotalCredits: credF,
	})
}


