package delivery

import (
	"File-management-system/server/internal/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService domain.UserService
	fileService domain.FileService
}

func NewHandler(uS domain.UserService, fS domain.FileService) *Handler {
	return &Handler{userService: uS, fileService: fS}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		files := api.Group("/files", h.userIdentify)
		{
			files.POST("/upload", h.Upload)
			files.GET("/:id", h.Download)
		}

		tests := api.Group("/test")
		{
			tests.GET("/ping", h.Ping)
		}
		//TODO write handlers for those endpoints
		//
		converts := api.Group("/convert")
		{
			//converts.GET("/" , toFill)

			converts.POST("/img-pdf/:id", h.ConvertImageToPDF)
			//converts.POST("pdf-img" , toFill)
		}
	}

	return r
}
