package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) uploadAttachment(c *gin.Context) {
	if a.attachments == nil {
		writeError(c, errors.New("attachment adapter is unavailable"), nil)
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		writeError(c, err, map[string]string{"file": "请选择要上传的文件"})
		return
	}
	defer file.Close()
	actor, _ := actorAndMeta(c)
	token, err := a.attachments.Save(c.Request.Context(), actor.UserID, header.Filename, file, header.Size)
	if err != nil {
		writeError(c, err, map[string]string{"file": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"attachment_token": token, "filename": header.Filename, "size": header.Size})
}
