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
	"net/http"
	"os"
	"os/signal"
	"time"
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

	userSvc := service.NewUserService(userRepo, cfg.JWTSECRET)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerPool := worker.NewPool(5) //IN future increase number of workers as tasks get higher
	workerPool.Start(ctx)

	fileSvc := service.NewFileService(fileRepo, userRepo, cfg.StoragePath, workerPool)

	h := delivery.NewHandler(userSvc, fileSvc, cfg.JWTSECRET)

	router := h.InitRouter()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	//running server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Listening on " + cfg.Port)

	//waiting server to turn off
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutdown Server ...")

	//5 sec to end the opertaions
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown: %s", err)
	}

	log.Println("Server exiting")

}
