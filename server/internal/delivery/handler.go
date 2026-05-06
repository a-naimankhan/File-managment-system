package delivery

import (
	"File-management-system/server/internal/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService   domain.UserService
	fileService   domain.FileService
	folderService domain.FolderService
	jwtSecret     string
}

func NewHandler(uS domain.UserService, fS domain.FileService, folderS domain.FolderService, jwtSecret string) *Handler {
	return &Handler{userService: uS, fileService: fS, folderService: folderS, jwtSecret: jwtSecret}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		// Публичные руты
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		// Защищенные руты (нужен токен)
		// Добавляем миддлвеер на всю группу
		protected := api.Group("/", h.userIdentify)
		{
			files := protected.Group("/files")
			{
				files.POST("/upload", h.Upload)
				files.GET("/:id", h.Download)
				files.GET("/", h.ListFiles)
				files.DELETE("/:id", h.DeleteFile)
			}

			folders := protected.Group("/folders")
			{
				folders.POST("/", h.CreateFolder)
				folders.DELETE("/:id", h.DeleteFolder)
				folders.PATCH("/:id/rename", h.RenameFolder)
			}

			protected.GET("/tree", h.ListContents)

			converts := protected.Group("/convert")
			{
				converts.POST("/img-pdf/:id", h.Execute)
			}
		}

		tests := api.Group("/test")
		{
			tests.GET("/ping", h.Ping)
		}

	}
	return r
}
