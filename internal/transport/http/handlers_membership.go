package httptransport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/domain"
)

func (a *API) discoverTeams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := a.engine.DiscoverTeams(c.Request.Context(), application.TeamFilter{CommunityID: c.Query("community_id"), Status: domain.TeamStatus(c.Query("status")), Query: c.Query("query"), Sort: c.Query("sort"), Page: page, PageSize: size})
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *API) joinTeam(c *gin.Context) {
	var body struct {
		InviteCode string `json:"invite_code" validate:"max=64"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	if err := a.engine.JoinTeam(c.Request.Context(), actor, meta, application.JoinRequest{TeamID: c.Param("teamID"), InviteCode: body.InviteCode}); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusCreated)
}

func (a *API) approveMembership(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.ApproveMembership(c.Request.Context(), actor, meta, c.Param("membershipID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) pauseMembership(c *gin.Context) {
	var body struct {
		Pause bool `json:"pause"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	if err := a.engine.PauseMembership(c.Request.Context(), actor, meta, c.Param("membershipID"), body.Pause); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) exitMembership(c *gin.Context) {
	actor, meta := actorAndMeta(c)
	if err := a.engine.ExitTeam(c.Request.Context(), actor, meta, c.Param("membershipID")); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) transferCaptain(c *gin.Context) {
	var body struct {
		TargetUserID string `json:"target_user_id" validate:"required"`
		Reason       string `json:"reason" validate:"required,max=500"`
	}
	if !a.bind(c, &body) {
		return
	}
	actor, meta := actorAndMeta(c)
	if err := a.engine.TransferCaptain(c.Request.Context(), actor, meta, c.Param("teamID"), body.TargetUserID, body.Reason); err != nil {
		writeError(c, err, nil)
		return
	}
	c.Status(http.StatusNoContent)
}
