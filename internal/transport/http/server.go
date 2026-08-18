package httptransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/middleware"
	"go.uber.org/zap"
)

type API struct {
	engine      *application.Engine
	attachments application.AttachmentStore
	validate    *validator.Validate
	ready       func() bool
}

func NewRouter(engine *application.Engine, attachments application.AttachmentStore, logger *zap.Logger, timeout time.Duration, ready func() bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.RequestID(), middleware.SecurityHeaders(), middleware.CORS("http://localhost:5173"), middleware.Timeout(timeout), middleware.Actor(), middleware.Logger(logger), middleware.Recovery(logger))
	api := &API{engine: engine, attachments: attachments, validate: validator.New(), ready: ready}
	router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/readyz", api.readiness)
	v1 := router.Group("/api/v1")
	{
		v1.GET("/catalog", api.catalog)
		v1.POST("/attachments", api.uploadAttachment)
		v1.GET("/discovery/teams", api.discoverTeams)
		v1.POST("/teams/:teamID/join", api.joinTeam)
		v1.POST("/memberships/:membershipID/approve", api.approveMembership)
		v1.POST("/memberships/:membershipID/pause", api.pauseMembership)
		v1.POST("/memberships/:membershipID/exit", api.exitMembership)
		v1.POST("/teams/:teamID/transfer-captain", api.transferCaptain)

		v1.GET("/activities", api.listActivities)
		v1.POST("/activities", api.createActivity)
		v1.GET("/activities/:activityID", api.activityDetail)
		v1.POST("/activities/:activityID/publish", api.publishActivity)
		v1.POST("/activities/:activityID/cancel", api.cancelActivity)
		v1.POST("/activities/:activityID/complete", api.completeActivity)
		v1.POST("/activities/:activityID/slots", api.createSlot)
		v1.POST("/slots/:slotID/claims", api.claimSlot)
		v1.POST("/slots/:slotID/promote", api.promoteWaitlist)
		v1.POST("/claims/:claimID/cancel", api.cancelClaim)
		v1.POST("/claims/:claimID/check-in", api.checkIn)

		v1.POST("/claims/:claimID/service/start", api.startService)
		v1.POST("/service-records/:recordID/end", api.endService)
		v1.POST("/service-records/:recordID/confirm", api.confirmService)
		v1.POST("/service-records/:recordID/corrections", api.requestCorrection)
		v1.GET("/profiles/:userID/records", api.listServiceRecords)
		v1.GET("/profiles/:userID/certificate", api.exportCertificate)

		v1.POST("/risks", api.createRisk)
		v1.POST("/risks/:riskID/follow-ups", api.addFollowUp)
		v1.POST("/risks/:riskID/resolve", api.resolveRisk)
		v1.POST("/handoffs", api.createHandoff)
		v1.POST("/handoffs/:handoffID/acknowledge", api.acknowledgeHandoff)
		v1.POST("/teams/:teamID/announcements", api.createAnnouncement)
		v1.GET("/notices", api.listNotices)
		v1.GET("/teams/:teamID/dashboard", api.dashboard)
	}
	return router
}

func (a *API) readiness(c *gin.Context) {
	if a.ready != nil && !a.ready() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (a *API) catalog(c *gin.Context) {
	result, err := a.engine.Catalog(c.Request.Context())
	if err != nil {
		writeError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, result)
}
