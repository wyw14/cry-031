package httptransport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/domain"
)

func (a *API) startService(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	result, err := a.engine.StartService(c.Request.Context(), actor, meta, c.Param("claimID"))
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) endService(c *gin.Context) {
	var body struct {
		PhotoNote string `json:"photo_note" validate:"max=1000"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.EndService(c.Request.Context(), actor, meta, c.Param("recordID"), body.PhotoNote)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) confirmService(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.ConfirmService(c.Request.Context(), actor, meta, c.Param("recordID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) requestCorrection(c *gin.Context) {
	var body struct {
		Note      string    `json:"note" validate:"required,max=1000"`
		StartedAt time.Time `json:"started_at" validate:"required"`
		EndedAt   time.Time `json:"ended_at" validate:"required"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.RequestCorrection(c.Request.Context(), actor, meta, c.Param("recordID"), body.Note, body.StartedAt, body.EndedAt)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) listServiceRecords(c *gin.Context) {
	actor, _ := actorAndMeta(c)
	userID := c.Param("userID")
	allowed, err := a.engine.CanAccessProfile(c.Request.Context(), actor, userID)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	if !allowed {
		writeError(c, domain.ErrForbidden, nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := a.engine.ListServiceRecords(c.Request.Context(), application.ServiceRecordFilter{UserID: userID, TeamID: c.Query("team_id"), Status: domain.ServiceStatus(c.Query("status")), Sort: c.Query("sort"), Page: page, Size: size})
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) exportCertificate(c *gin.Context) {
	actor, _ := actorAndMeta(c)
	userID := c.Param("userID")
	allowed, err := a.engine.CanAccessProfile(c.Request.Context(), actor, userID)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	if !allowed {
		writeError(c, domain.ErrForbidden, nil)
		return
	}
	data, err := a.engine.ExportCertificate(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="service-certificate.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
