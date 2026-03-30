package delivery

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) Upload(c *gin.Context) {
	//TODO
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "no file uploaded"})
	}

	userStr := c.PostForm("user_id")
	userID, err := uuid.Parse(userStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	fileContent, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "invalid file content"})
		return
	}

	defer fileContent.Close()

	metadata, err := h.fileService.UploadFile(c.Request.Context(), userID, fileHeader.Filename, fileContent)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	fmt.Println(metadata)
	c.JSON(200, metadata)
}

func (h *Handler) Download(c *gin.Context) {

	fileIDStr := c.PostForm("id")
	fileID, err := uuid.Parse(fileIDStr)

	if err != nil {
		c.JSON(400, gin.H{"error": "invalid file id"})
	}

	fileMeta, err := h.fileService.DownloadFile(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	c.File(fileMeta.Path)
}
