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
		c.JSON(400, gin.H{"error": "" +
			"no file uploaded"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	userIDStr, ok := val.(string)
	if !ok {
		c.JSON(400, gin.H{"error": "internal error : user id format"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "internal error : user id format"})
		return
	}
	fileContent, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "couldn't open the file"})
		return
	}
	defer fileContent.Close()

	metadata, err := h.fileService.UploadFile(c.Request.Context(), userID, fileHeader.Filename, fileContent)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(metadata)
	c.JSON(200, metadata)
}

func (h *Handler) Download(c *gin.Context) {

	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)

	if err != nil {
		c.JSON(400, gin.H{"error": "invalid file id"})
		return
	}

	fileMeta, err := h.fileService.DownloadFile(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	c.File(fileMeta.Path)
}

func (h *Handler) Execute(c *gin.Context) {
	fileIDstr := c.Param("id")

	fileId, err := uuid.Parse(fileIDstr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid file id"})
		return
	}

	if err := h.fileService.StartImageToPDF(c.Request.Context(), fileId); err != nil {
		c.JSON(500, gin.H{"error": "couldn't start the image"})
		return
	}

}
