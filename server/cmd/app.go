package main

import (
	"File-management-system/server/internal/delivery"
	"File-management-system/server/internal/repository/memory"
	"File-management-system/server/internal/service"
)

func main() {
	startServer()
	select {}
}

func startServer() {

	userRepo := memory.NewUserRepository()
	fileRepo := memory.NewFileRepository()

	userSvc := service.NewUserService(userRepo)
	fileSvc := service.NewFileService(fileRepo, userRepo, "./upload")

	h := delivery.NewHandler(userSvc, fileSvc)

	router := h.InitRouter()
	router.Run(":8080")
}
