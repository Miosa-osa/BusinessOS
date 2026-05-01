package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations/miosa"
)

// BillingHandler serves plan and subscription endpoints.
// Open-source stub — all billing features return community-edition defaults.
type BillingHandler struct {
	cfg         *config.Config
	miosaClient *miosa.ComputeClient
}

// NewBillingHandler constructs a BillingHandler.
func NewBillingHandler(cfg *config.Config, miosaClient *miosa.ComputeClient) *BillingHandler {
	return &BillingHandler{cfg: cfg, miosaClient: miosaClient}
}

// availablePlans is the canonical list of paid tiers.
// Shared with computer.go via planSpecs().
// Open-source edition: plans are listed but cannot be subscribed to.
var availablePlans = []PlanInfo{
	{
		ID:           "pro",
		Name:         "Pro",
		PriceMonthly: 4000,
		RAMGB:        4,
		CPUs:         2,
		StorageGB:    10,
		Credits:      500,
		Features: []string{
			"Cloud computer",
			"Unlimited seats",
			"All integrations",
			"500 agent credits/mo",
		},
	},
	{
		ID:           "growth",
		Name:         "Growth",
		PriceMonthly: 10000,
		RAMGB:        8,
		CPUs:         4,
		StorageGB:    50,
		Credits:      2000,
		Features: []string{
			"Everything in Pro",
			"8GB RAM",
			"4 vCPUs",
			"2,000 agent credits/mo",
			"Priority support",
		},
	},
	{
		ID:           "business",
		Name:         "Business",
		PriceMonthly: 20000,
		RAMGB:        16,
		CPUs:         8,
		StorageGB:    100,
		Credits:      5000,
		Features: []string{
			"Everything in Growth",
			"16GB RAM",
			"8 vCPUs",
			"5,000 agent credits/mo",
			"SSO",
			"Audit logs",
			"Dedicated support",
		},
	},
}

func (h *BillingHandler) GetPlans(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"plans": availablePlans, "message": "billing not available in open-source edition"})
}

func (h *BillingHandler) GetSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"plan": "community", "message": "open-source edition"})
}

func (h *BillingHandler) Subscribe(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not available in open-source edition"})
}

func (h *BillingHandler) PurchaseCredits(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not available in open-source edition"})
}

func (h *BillingHandler) HandleStripeWebhook(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not available in open-source edition"})
}
