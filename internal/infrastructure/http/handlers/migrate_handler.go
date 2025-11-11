package handlers

import (
	"net/http"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/infrastructure/http/validators"
	"stori-challenge/internal/ports/services"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
)

type MigrateHandler struct {
	Service services.MigrationService
}

func NewMigrateHandler(svc services.MigrationService) *MigrateHandler {
	return &MigrateHandler{Service: svc}
}

// PostMigrate
// @Summary      Migrate transactions via CSV upload
// @Description  Accepts a CSV file with columns id,user_id,amount,datetime and migrates transactions
// @Tags         migrate
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "CSV file"
// @Success      201  {object}  shared.SuccessResponse
// @Failure      400  {object}  shared.ErrorResponse
// @Failure      409  {object}  shared.ErrorResponse
// @Router       /migrate [post]
func (h *MigrateHandler) PostMigrate(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		CreateErrorResponse(c,
			shared.NewBadRequest("missing_file", "file is required", nil),
			[]responses.MigrateRowError{{
				Row: 0, Field: "file", Value: "", Message: "file is required",
			}},
		)
		return
	}
	if appErr := validators.ValidateFileMeta(fileHeader.Filename, fileHeader.Size); appErr != nil {
		CreateErrorResponse(c,
			appErr,
			[]responses.MigrateRowError{{
				Row: 0, Field: "file", Value: fileHeader.Filename, Message: appErr.Msg,
			}},
		)
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		CreateErrorResponse(c,
			shared.NewBadRequest("file_open_failed", "unable to open uploaded file", nil),
			[]responses.MigrateRowError{{
				Row: 0, Field: "file", Value: fileHeader.Filename, Message: "unable to open uploaded file",
			}},
		)
		return
	}
	defer f.Close()

	inserted, errItems, svcErr := h.Service.Process(c.Request.Context(), f)
	if svcErr != nil {
		// Map []services.RowError -> []responses.MigrateRowError for details
		out := make([]responses.MigrateRowError, 0, len(errItems))
		for _, it := range errItems {
			out = append(out, responses.MigrateRowError{
				Row:     it.Row,
				Field:   it.Field,
				Value:   it.Value,
				Message: it.Message,
			})
		}
		CreateErrorResponse(c, svcErr, out)
		return
	}

	c.JSON(http.StatusCreated, responses.MigrateSuccessResponse{Inserted: inserted})
}
