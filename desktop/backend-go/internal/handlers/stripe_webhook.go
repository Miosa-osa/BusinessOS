package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleStripeWebhook handles POST /api/v1/billing/webhook.
// Stripe is not yet configured — returns 503 until wired up.
func (h *BillingHandler) HandleStripeWebhook(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stripe not configured"})
}
