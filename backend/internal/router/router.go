package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/handler"
	"github.com/lobsterpool/lobsterpool/internal/middleware"
	"github.com/lobsterpool/lobsterpool/internal/models"
)

type Handlers struct {
	Health   *handler.HealthHandler
	Auth     *handler.AuthHandler
	Admin    *handler.AdminHandler
	Template *handler.TemplateHandler
	Instance *handler.InstanceHandler
	Cluster  *handler.ClusterHandler
}

func registerAuthRoutes(api *gin.RouterGroup, h *Handlers, jwtSecret string, userStore *models.UserStore) {
	auth := api.Group("/auth")
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)
	auth.POST("/logout", h.Auth.Logout)
	auth.GET("/me", middleware.Auth(jwtSecret, userStore), h.Auth.Me)
	auth.POST("/change-password", middleware.Auth(jwtSecret, userStore), h.Auth.ChangePassword)
}

func registerTemplateRoutes(api *gin.RouterGroup, h *Handlers, authRequired gin.HandlerFunc) {
	templates := api.Group("/templates")
	templates.Use(authRequired)
	templates.GET("", h.Template.List)
	templates.GET("/:id", h.Template.Get)
}

func registerClusterRoutes(api *gin.RouterGroup, h *Handlers, authRequired gin.HandlerFunc) {
	clusters := api.Group("/clusters")
	clusters.Use(authRequired)
	clusters.GET("", h.Cluster.List)
}

func registerInstanceRoutes(api *gin.RouterGroup, h *Handlers, authRequired gin.HandlerFunc) {
	instances := api.Group("/instances")
	instances.Use(authRequired)
	instances.GET("", h.Instance.List)
	instances.POST("", h.Instance.Create)
	instances.GET("/:id", h.Instance.Get)
	instances.DELETE("/:id", h.Instance.Delete)
}

func registerAdminRoutes(api *gin.RouterGroup, h *Handlers, authRequired gin.HandlerFunc) {
	admin := api.Group("/admin")
	admin.Use(authRequired, middleware.AdminOnly())
	admin.GET("/overview", h.Admin.Overview)
	admin.GET("/users", h.Admin.ListUsers)
	admin.PATCH("/users/:id/max-instances", h.Admin.UpdateUserMaxInstances)
	admin.GET("/instances", h.Admin.ListInstances)
	admin.POST("/templates", h.Admin.CreateTemplate)
}

func Setup(h *Handlers, jwtSecret string, userStore *models.UserStore) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	api.GET("/health", h.Health.Check)
	registerAuthRoutes(api, h, jwtSecret, userStore)

	authRequired := middleware.Auth(jwtSecret, userStore)
	registerTemplateRoutes(api, h, authRequired)
	registerClusterRoutes(api, h, authRequired)
	registerInstanceRoutes(api, h, authRequired)
	registerAdminRoutes(api, h, authRequired)

	return r
}
