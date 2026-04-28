package delivery

import (
	"File-management-system/server/internal/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService domain.UserService
	fileService domain.FileService
	jwtSecret   string
}

func NewHandler(uS domain.UserService, fS domain.FileService, jwtSecret string) *Handler {
	return &Handler{userService: uS, fileService: fS, jwtSecret: jwtSecret}
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
		// Добавляем миддлварь на всю группу
		protected := api.Group("/", h.userIdentify)
		{
			files := protected.Group("/files")
			{
				files.POST("/upload", h.Upload)
				files.GET("/:id", h.Download)
				files.GET(, h.ListFiles)
				files.DELETE("/:id", h.DeleteByID)
			}

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
