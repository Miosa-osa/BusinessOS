package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
)

// HandleStripeWebhook handles POST /api/v1/billing/webhook.
// Stripe signs every event with a HMAC signature; we verify it before processing.
// The route must NOT have auth middleware or body-parsing middleware applied —
// the raw body is required for signature verification.
func (h *BillingHandler) HandleStripeWebhook(c *gin.Context) {
	if !h.stripe.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stripe not configured"})
		return
	}

	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20)) // 1 MB limit
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "stripe webhook: failed to read body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		slog.WarnContext(c.Request.Context(), "stripe webhook: missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Stripe-Signature header"})
		return
	}

	event, err := h.stripe.HandleWebhook(payload, signature)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "stripe webhook: signature verification failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"})
		return
	}

	slog.InfoContext(c.Request.Context(), "stripe webhook: received event", "type", event.Type, "id", event.ID)

	switch event.Type {

	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			slog.ErrorContext(c.Request.Context(), "stripe webhook: failed to parse checkout session", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse event data"})
			return
		}
		h.handleCheckoutCompleted(c, &sess)

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			slog.ErrorContext(c.Request.Context(), "stripe webhook: failed to parse subscription", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse event data"})
			return
		}
		h.handleSubscriptionUpdated(c, &sub)

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			slog.ErrorContext(c.Request.Context(), "stripe webhook: failed to parse subscription", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse event data"})
			return
		}
		h.handleSubscriptionDeleted(c, &sub)

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			slog.ErrorContext(c.Request.Context(), "stripe webhook: failed to parse invoice", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse event data"})
			return
		}
		h.handlePaymentFailed(c, &inv)

	default:
		// Unhandled event type — acknowledge receipt so Stripe does not retry.
		slog.InfoContext(c.Request.Context(), "stripe webhook: unhandled event type", "type", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// handleCheckoutCompleted is called when a Stripe Checkout Session completes
// successfully. The workspace_id and plan are read from session metadata.
// TODO: persist a platform_subscriptions record once the DB schema is wired.
func (h *BillingHandler) handleCheckoutCompleted(c *gin.Context, sess *stripe.CheckoutSession) {
	workspaceID := sess.Metadata["workspace_id"]
	plan := sess.Metadata["plan"]
	customerID := ""
	if sess.Customer != nil {
		customerID = sess.Customer.ID
	}
	subscriptionID := ""
	if sess.Subscription != nil {
		subscriptionID = sess.Subscription.ID
	}

	slog.InfoContext(c.Request.Context(), "stripe webhook: checkout completed",
		"workspace_id", workspaceID,
		"plan", plan,
		"customer_id", customerID,
		"subscription_id", subscriptionID,
	)

	// TODO: upsert platform_subscriptions record:
	//   INSERT INTO platform_subscriptions
	//     (workspace_id, stripe_customer_id, stripe_subscription_id, plan, status)
	//   VALUES ($1, $2, $3, $4, 'active')
	//   ON CONFLICT (workspace_id) DO UPDATE SET ...
}

// handleSubscriptionUpdated is called when a subscription changes plan or status.
// TODO: update platform_subscriptions record.
func (h *BillingHandler) handleSubscriptionUpdated(c *gin.Context, sub *stripe.Subscription) {
	customerID := ""
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	slog.InfoContext(c.Request.Context(), "stripe webhook: subscription updated",
		"subscription_id", sub.ID,
		"customer_id", customerID,
		"status", sub.Status,
		"cancel_at_period_end", sub.CancelAtPeriodEnd,
	)

	// TODO: UPDATE platform_subscriptions SET status=$1, cancel_at_period_end=$2
	// WHERE stripe_subscription_id=$3
}

// handleSubscriptionDeleted is called when a subscription is fully cancelled.
// TODO: mark platform_subscriptions record as cancelled.
func (h *BillingHandler) handleSubscriptionDeleted(c *gin.Context, sub *stripe.Subscription) {
	customerID := ""
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	slog.InfoContext(c.Request.Context(), "stripe webhook: subscription deleted",
		"subscription_id", sub.ID,
		"customer_id", customerID,
	)

	// TODO: UPDATE platform_subscriptions SET status='canceled'
	// WHERE stripe_subscription_id=$1
}

// handlePaymentFailed is called when an invoice payment fails.
// TODO: mark subscription as past_due and trigger dunning notification.
func (h *BillingHandler) handlePaymentFailed(c *gin.Context, inv *stripe.Invoice) {
	customerID := ""
	if inv.Customer != nil {
		customerID = inv.Customer.ID
	}

	// In stripe-go v82 the subscription reference lives under Parent.
	subscriptionID := ""
	if inv.Parent != nil &&
		inv.Parent.SubscriptionDetails != nil &&
		inv.Parent.SubscriptionDetails.Subscription != nil {
		subscriptionID = inv.Parent.SubscriptionDetails.Subscription.ID
	}

	slog.WarnContext(c.Request.Context(), "stripe webhook: payment failed",
		"invoice_id", inv.ID,
		"customer_id", customerID,
		"subscription_id", subscriptionID,
		"amount_due", inv.AmountDue,
	)

	// TODO: UPDATE platform_subscriptions SET status='past_due'
	// WHERE stripe_subscription_id=$1
	// TODO: enqueue dunning email notification
}
