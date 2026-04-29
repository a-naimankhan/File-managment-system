package delivery

import (
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

	c.JSON(200, metadata)
}

func (h *Handler) Download(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	userId, err := uuid.Parse(val.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "internal error : user id format"})
		return
	}

	fileIdStr := c.Param("file")
	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "internal error : file id format"})
		return
	}

	fileMeta, err := h.fileService.DownloadFile(c.Request.Context(), userId, fileId)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(403, gin.H{"error": "access denied"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.File(fileMeta.Path)

}

func (h *Handler) Execute(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	userId, err := uuid.Parse(val.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "internal error : user id format"})
		return
	}

	fileIdStr := c.Param("file")
	fileId, err := uuid.Parse(fileIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "internal error : file id format"})
		return
	}

	if err := h.fileService.StartImageToPDF(c.Request.Context(), userId, fileId); err != nil {
		if err.Error() == "access denied" {
			c.JSON(403, gin.H{"error": "access denied"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "conversion started"})

}

func (h *Handler) ListFiles(c *gin.Context) {
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
		c.JSON(400, gin.H{"error": "invalid user id format"})
		return
	}

	files, err := h.fileService.ListFiles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, files)

}

func (h *Handler) DeleteFile(c *gin.Context) {
	val, exist := c.Get("userID")
	if !exist {
		c.JSON(401, gin.H{"error": "user not found in context"})
		return
	}

	userId, err := uuid.Parse(val.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id format"})
		return
	}

	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid file id"})
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), userId, fileID); err != nil {
		if err.Error() == "access denied" {
			c.JSON(403, gin.H{"error": "access denied"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "file deleted"})

}
