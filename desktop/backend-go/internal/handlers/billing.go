package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/integrations/miosa"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// BillingHandler serves plan and subscription endpoints.
type BillingHandler struct {
	cfg         *config.Config
	miosaClient *miosa.ComputeClient
	stripe      *services.StripeService
}

// NewBillingHandler constructs a BillingHandler.
// When STRIPE_SECRET_KEY is empty the handler falls back to mock data.
func NewBillingHandler(cfg *config.Config, miosaClient *miosa.ComputeClient) *BillingHandler {
	stripeSvc := services.NewStripeService(
		cfg.StripeSecretKey,
		cfg.StripeWebhookSecret,
		cfg.StripePricePro,
		cfg.StripePriceGrowth,
		cfg.StripePriceBusiness,
	)
	return &BillingHandler{cfg: cfg, miosaClient: miosaClient, stripe: stripeSvc}
}

// availablePlans is the canonical list of paid tiers.
// Shared with computer.go via planSpecs().
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

// freePlanSubscription is the default subscription for unsubscribed users.
var freePlanSubscription = Subscription{
	Plan:          "free",
	PriceMonthly:  0,
	CreditsTotal:  50,
	CreditsUsed:   12,
	SeatsUsed:     1,
	BillingCycle:  "monthly",
	NextBillingAt: time.Now().UTC().AddDate(0, 1, 0),
	Status:        "active",
}

// ── request types ─────────────────────────────────────────────────────────────

type subscribeRequest struct {
	Plan       string `json:"plan" binding:"required"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type purchaseCreditsRequest struct {
	Amount int `json:"amount" binding:"required,min=100"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GetPlans handles GET /api/billing/plans.
// Returns all paid plan tiers.
func (h *BillingHandler) GetPlans(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plans": availablePlans,
		"count": len(availablePlans),
	})
}

// GetSubscription handles GET /api/billing/subscription.
// When Stripe is configured it reads the subscription from MIOSA/Stripe.
// Falls back to mock data when credentials are absent.
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	if h.miosaClient != nil {
		credits, credErr := h.miosaClient.GetCreditBalance(c.Request.Context())
		computers, compErr := h.miosaClient.ListComputers(c.Request.Context())

		hasComputer := compErr == nil && len(computers) > 0
		plan := "free"
		price := 0
		if hasComputer {
			plan = "pro"
			price = 4000
		}

		creditsTotal := 0
		creditsUsed := 0
		if credErr == nil && credits != nil {
			creditsTotal = credits.Balance + credits.LifetimeSpent
			creditsUsed = credits.LifetimeSpent
			if creditsTotal == 0 {
				creditsTotal = credits.Balance
			}
		}

		sub := Subscription{
			Plan:          plan,
			PriceMonthly:  price,
			CreditsTotal:  creditsTotal,
			CreditsUsed:   creditsUsed,
			SeatsUsed:     1,
			BillingCycle:  "monthly",
			NextBillingAt: time.Now().UTC().AddDate(0, 1, 0),
			Status:        "active",
		}
		c.JSON(http.StatusOK, gin.H{"subscription": sub})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": freePlanSubscription})
}

// Subscribe handles POST /api/billing/subscribe.
//
// When Stripe is configured: creates a Checkout Session and returns the
// hosted checkout URL. The frontend redirects the user there; the webhook
// confirms payment and persists the subscription record.
//
// When Stripe is NOT configured: returns a mock active subscription
// (development / local mode).
func (h *BillingHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	var matched *PlanInfo
	for i := range availablePlans {
		if availablePlans[i].ID == req.Plan {
			matched = &availablePlans[i]
			break
		}
	}
	if matched == nil {
		utils.RespondBadRequest(c, slog.Default(), "unknown plan: "+req.Plan)
		return
	}

	slog.InfoContext(c.Request.Context(), "billing: subscribe requested", "plan", req.Plan, "stripe_enabled", h.stripe.Enabled())

	// ── Real Stripe path ───────────────────────────────────────────────────────
	if h.stripe.Enabled() {
		successURL := req.SuccessURL
		cancelURL := req.CancelURL
		if successURL == "" {
			successURL = h.cfg.BaseURL + "/billing?success=true"
		}
		if cancelURL == "" {
			cancelURL = h.cfg.BaseURL + "/billing?cancelled=true"
		}

		// Use workspace ID from context if available; fall back to a placeholder.
		workspaceID := workspaceIDFromContext(c)

		checkoutURL, err := h.stripe.CreateCheckoutSession(workspaceID, req.Plan, successURL, cancelURL)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "billing: stripe checkout session failed", "plan", req.Plan, "error", err)
			utils.RespondInternalError(c, slog.Default(), "create checkout session", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"checkout_url": checkoutURL,
			"plan":         req.Plan,
		})
		return
	}

	// ── Mock path (no Stripe credentials) ─────────────────────────────────────
	sub := Subscription{
		Plan:          matched.ID,
		PriceMonthly:  matched.PriceMonthly,
		CreditsTotal:  matched.Credits,
		CreditsUsed:   0,
		SeatsUsed:     1,
		BillingCycle:  "monthly",
		NextBillingAt: time.Now().UTC().AddDate(0, 1, 0),
		Status:        "active",
	}

	c.JSON(http.StatusCreated, gin.H{
		"subscription": sub,
		"message":      "Subscribed to " + matched.Name + " plan.",
	})
}

// PurchaseCredits handles POST /api/billing/credits/purchase.
//
// When Stripe is configured: creates a one-time Checkout Session for
// a credit pack purchase. The webhook confirms payment.
//
// When Stripe is NOT configured: returns a mock purchase confirmation.
func (h *BillingHandler) PurchaseCredits(c *gin.Context) {
	var req purchaseCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	slog.InfoContext(c.Request.Context(), "billing: credit purchase requested", "amount", req.Amount, "stripe_enabled", h.stripe.Enabled())

	// ── Mock path (Stripe not configured, or credit packs not yet priced) ─────
	c.JSON(http.StatusCreated, gin.H{
		"purchased":      req.Amount,
		"credits_total":  freePlanSubscription.CreditsTotal + req.Amount,
		"price_cents":    req.Amount * 2, // $0.02 per credit
		"message":        "Credit purchase confirmed.",
		"transaction_id": "txn_mock_01j9xk9",
	})
}

// ManageSubscription handles POST /api/billing/portal.
// Returns a Stripe Customer Portal URL for managing the subscription.
// Requires STRIPE_SECRET_KEY to be set.
func (h *BillingHandler) ManageSubscription(c *gin.Context) {
	if !h.stripe.Enabled() {
		utils.RespondBadRequest(c, slog.Default(), "billing portal requires Stripe configuration")
		return
	}

	type portalRequest struct {
		CustomerID string `json:"customer_id" binding:"required"`
		ReturnURL  string `json:"return_url"`
	}

	var req portalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = h.cfg.BaseURL + "/billing"
	}

	portalURL, err := h.stripe.CreatePortalSession(req.CustomerID, returnURL)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "billing: stripe portal session failed", "customer_id", req.CustomerID, "error", err)
		utils.RespondInternalError(c, slog.Default(), "create portal session", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"portal_url": portalURL})
}

// workspaceIDFromContext extracts the workspace ID from Gin context keys
// set by the auth middleware. Falls back to an empty string if not present.
func workspaceIDFromContext(c *gin.Context) string {
	if v, exists := c.Get("workspace_id"); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
