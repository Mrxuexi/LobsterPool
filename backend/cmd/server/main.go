package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/config"
	"github.com/lobsterpool/lobsterpool/internal/database"
	"github.com/lobsterpool/lobsterpool/internal/handler"
	"github.com/lobsterpool/lobsterpool/internal/models"
	"github.com/lobsterpool/lobsterpool/internal/provider"
	"github.com/lobsterpool/lobsterpool/internal/router"
)

func shouldUseMockProvider(cfg *config.Config) bool {
	return cfg.DevMode || os.Getenv("LP_MOCK_PROVIDER") == "true"
}

func buildProvider(cfg *config.Config) (provider.Provider, error) {
	if shouldUseMockProvider(cfg) {
		log.Println("Using mock provider")
		return provider.NewMockProvider(), nil
	}

	log.Println("Using Kubernetes provider")
	return provider.NewKubernetesProvider(cfg.Kubeconfig, cfg.Namespace)
}

func setupStaticRoutes(r *gin.Engine, staticDir string) {
	if _, err := os.Stat(staticDir); err != nil {
		return
	}

	r.Static("/assets", staticDir+"/assets")
	r.StaticFile("/", staticDir+"/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File(staticDir + "/index.html")
	})
}

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	templateStore := models.NewTemplateStore(db)
	instanceStore := models.NewInstanceStore(db)
	userStore := models.NewUserStore(db)

	p, err := buildProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	handlers := &router.Handlers{
		Health:   handler.NewHealthHandler(),
		Auth:     handler.NewAuthHandler(userStore, cfg.JWTSecret),
		Admin:    handler.NewAdminHandler(userStore, templateStore, instanceStore),
		Template: handler.NewTemplateHandler(templateStore),
		Instance: handler.NewInstanceHandler(instanceStore, templateStore, userStore, p, cfg.Namespace),
	}

	r := router.Setup(handlers, cfg.JWTSecret, userStore)
	setupStaticRoutes(r, cfg.StaticDir)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("LobsterPool starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
