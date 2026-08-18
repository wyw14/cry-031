package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"github.com/wyw/cry031-volunteer/internal/middleware"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id"`
}

func (a *API) bind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		writeError(c, err, map[string]string{"body": "请求体格式不正确"})
		return false
	}
	if err := a.validate.Struct(target); err != nil {
		fields := map[string]string{}
		if validation, ok := err.(validator.ValidationErrors); ok {
			for _, item := range validation {
				fields[item.Field()] = "字段校验失败: " + item.Tag()
			}
		}
		writeError(c, err, fields)
		return false
	}
	return true
}

func writeError(c *gin.Context, err error, fields map[string]string) {
	status, code, message := http.StatusInternalServerError, "internal_error", "服务器内部错误"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "资源不存在"
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = http.StatusForbidden, "forbidden", "当前用户没有权限执行该操作"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrScheduleConflict), errors.Is(err, domain.ErrCapacityExceeded):
		status, code, message = http.StatusConflict, "business_conflict", err.Error()
	case errors.Is(err, domain.ErrInvalidState), errors.Is(err, domain.ErrInvalidDuration), errors.Is(err, domain.ErrInvalidCorrection), errors.Is(err, domain.ErrActivityClosed), errors.Is(err, domain.ErrUnacknowledgedRisk):
		status, code, message = http.StatusUnprocessableEntity, "invalid_business_state", err.Error()
	case errors.Is(err, domain.ErrInvalidFilter):
		status, code, message = http.StatusBadRequest, "invalid_filter", err.Error()
	default:
		if len(fields) > 0 {
			status, code, message = http.StatusBadRequest, "validation_failed", "请求字段校验失败"
		}
	}
	c.JSON(status, errorEnvelope{Error: apiError{Code: code, Message: message, Fields: fields, RequestID: middleware.GetRequestID(c)}})
}

func actorAndMeta(c *gin.Context) (application.Actor, application.RequestMeta) {
	return middleware.GetActor(c), application.RequestMeta{RequestID: middleware.GetRequestID(c), IdempotencyKey: c.GetHeader("Idempotency-Key")}
}
