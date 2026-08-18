package httptransport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/domain"
)

func (a *API) listActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := a.engine.ListActivities(c.Request.Context(), application.ActivityFilter{TeamID: c.Query("team_id"), Status: domain.ActivityStatus(c.Query("status")), Query: c.Query("query"), Sort: c.Query("sort"), Page: page, Size: size})
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) createActivity(c *gin.Context) {
	var request application.CreateActivityRequest
	if !a.bind(c, &request) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.CreateActivity(c.Request.Context(), actor, meta, request)
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) activityDetail(c *gin.Context) {
	result, err := a.engine.ActivityDetail(c.Request.Context(), c.Param("activityID"))
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) publishActivity(c *gin.Context)  { a.activityAction(c, a.engine.PublishActivity) }
func (a *API) cancelActivity(c *gin.Context)   { a.activityAction(c, a.engine.CancelActivity) }
func (a *API) completeActivity(c *gin.Context) { a.activityAction(c, a.engine.CompleteActivity) }

func (a *API) activityAction(c *gin.Context, action func(context.Context, application.Actor, application.RequestMeta, string) error) {
	actor, meta := actorAndMeta(c)
	if err := action(c.Request.Context(), actor, meta, c.Param("activityID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) createSlot(c *gin.Context) {
	var body struct {
		Name        string `json:"name" validate:"required,max=80"`
		Description string `json:"description" validate:"max=500"`
		Capacity    int    `json:"capacity" validate:"gte=1,lte=500"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	result, err := a.engine.CreateSlot(c.Request.Context(), actor, meta, application.CreateSlotRequest{ActivityID: c.Param("activityID"), Name: body.Name, Description: body.Description, Capacity: body.Capacity})
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) claimSlot(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	result, err := a.engine.ClaimSlot(c.Request.Context(), actor, meta, c.Param("slotID"))
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (a *API) promoteWaitlist(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.PromoteWaitlist(c.Request.Context(), actor, meta, c.Param("slotID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) cancelClaim(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.CancelClaim(c.Request.Context(), actor, meta, c.Param("claimID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) checkIn(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.CheckIn(c.Request.Context(), actor, meta, c.Param("claimID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}
