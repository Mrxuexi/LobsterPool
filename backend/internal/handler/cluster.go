package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lobsterpool/lobsterpool/internal/provider"
)

type ClusterHandler struct {
	provider provider.Provider
}

func NewClusterHandler(p provider.Provider) *ClusterHandler {
	return &ClusterHandler{provider: p}
}

func (h *ClusterHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.provider.ListClusters())
}
