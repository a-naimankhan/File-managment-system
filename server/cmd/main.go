package main

import (
	"File-management-system/server/config"
	"File-management-system/server/internal/delivery"
	"File-management-system/server/internal/repository/postgres"
	"File-management-system/server/internal/repository/postgres/connection"
	"File-management-system/server/internal/service"
	"File-management-system/server/internal/worker"
	"context"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	startServer(cfg)
}

func startServer(cfg *config.Config) {

	db, err := connection.NewPostgresDB(cfg.DB_DSN)
	if err != nil {
		panic("Failed to connect to database")
	}
	userRepo := postgres.NewUserRepo(db)
	fileRepo := postgres.NewFileRepo(db)

	//userRepo := memory.NewUserRepository()
	//fileRepo := memory.NewFileRepository()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wp := worker.NewPool(3)

	wp.Start(ctx)

	userSvc := service.NewUserService(userRepo, cfg.JWTSECRET)
	fileSvc := service.NewFileService(fileRepo, userRepo, cfg.StoragePath, wp)

	h := delivery.NewHandler(userSvc, fileSvc)

	router := h.InitRouter()
	router.Run(":" + cfg.Port)
}
