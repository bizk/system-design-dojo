package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bizk/system-design-dojo/server/internal/config"
	"github.com/bizk/system-design-dojo/server/internal/database"
	"github.com/bizk/system-design-dojo/server/internal/migrations"
	"github.com/bizk/system-design-dojo/server/internal/sessions"
	"github.com/bizk/system-design-dojo/server/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrations.Run(db); err != nil {
		log.Fatal(err)
	}

	storageClient, err := storage.Connect(cfg.StorageEndpoint, cfg.StorageAccessKey, cfg.StorageSecretKey, cfg.StorageUseSSL)
	if err != nil {
		log.Fatal(err)
	}
	if err := storage.EnsureBucket(context.Background(), storageClient, cfg.StorageBucket); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(cors())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	sessions.RegisterRoutes(router.Group("/api"), sessions.NewHandler(sessions.NewStore(db), storageClient, cfg.StorageBucket, cfg.MaxMediaSize))

	log.Fatal(router.Run(fmt.Sprintf(":%s", cfg.ServerPort)))
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
