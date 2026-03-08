package handlers

import "github.com/gin-gonic/gin"

// registerCRMRoutes wires up CRM and business-data routes:
// /api/clients, /api/crm, /api/tables.
func (h *Handlers) registerCRMRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	// Client routes - /api/clients
	RegisterClientRoutes(api, NewClientHandler(h.pool), auth)

	// CRM routes - /api/crm
	RegisterCRMRoutes(api, NewCRMHandler(h.pool), auth)

	// Table routes - /api/tables
	RegisterTableRoutes(api, NewTableHandler(h.pool), auth)
}
