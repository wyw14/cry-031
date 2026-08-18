package httptransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw/cry031-volunteer/internal/application"
)

func (a *API) createRisk(c *gin.Context) {
	var request application.CreateRiskRequest
	if !a.bind(c, &request) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.CreateRisk(c.Request.Context(), actor, meta, request)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) addFollowUp(c *gin.Context) {
	var body struct {
		Title   string    `json:"title" validate:"required,max=200"`
		OwnerID string    `json:"owner_id" validate:"required"`
		DueAt   time.Time `json:"due_at" validate:"required"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.AddFollowUp(c.Request.Context(), actor, meta, c.Param("riskID"), body.Title, body.OwnerID, body.DueAt)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) resolveRisk(c *gin.Context) {
	var body struct {
		Result string `json:"result" validate:"required,max=1000"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	if err := a.engine.ResolveRisk(c.Request.Context(), actor, meta, c.Param("riskID"), body.Result); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) createHandoff(c *gin.Context) {
	var request application.CreateHandoffRequest
	if !a.bind(c, &request) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.CreateHandoff(c.Request.Context(), actor, meta, request)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) acknowledgeHandoff(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.AcknowledgeHandoff(c.Request.Context(), actor, meta, c.Param("handoffID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) createAnnouncement(c *gin.Context) {
	var body struct {
		Title string `json:"title" validate:"required,max=120"`
		Body  string `json:"body" validate:"required,max=2000"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.CreateAnnouncement(c.Request.Context(), actor, meta, c.Param("teamID"), body.Title, body.Body)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) listNotices(c *gin.Context) {
	actor, _ := actorAndMeta(c)
	result, err := a.engine.ListNotices(c.Request.Context(), actor)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result})
}

func (a *API) dashboard(c *gin.Context) {
	actor, _ := actorAndMeta(c)
	result, err := a.engine.LeaderDashboard(c.Request.Context(), actor, c.Param("teamID"))
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}
