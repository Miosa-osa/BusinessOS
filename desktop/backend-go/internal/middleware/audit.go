package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rhl/businessos-backend/internal/models"
)

// A2AAuditServiceI defines the interface for A2A audit logging (hash-chain based)
type A2AAuditServiceI interface {
	LogA2ACall(agent, action, resourceType, resourceID string, snScore float64) (*models.AuditEntry, error)
	QueryAuditTrail(resourceType, resourceID string) ([]*models.AuditEntry, error)
}

// HashChainLogger provides hash-chain audit logging for A2A calls
// with cryptographic chain integrity (PROV-O compliant).
type HashChainLogger struct {
	secret  string
	entries []*models.AuditEntry
}

// NewHashChainLogger creates a new hash-chain logger with HMAC secret
func NewHashChainLogger(secret string) *HashChainLogger {
	return &HashChainLogger{
		secret:  secret,
		entries: make([]*models.AuditEntry, 0),
	}
}

// LogA2ACall creates a new audit entry with hash-chain integrity
func (al *HashChainLogger) LogA2ACall(
	agent, action, resourceType, resourceID string,
	snScore float64,
) (*models.AuditEntry, error) {
	entry := &models.AuditEntry{
		ID:           uuid.New().String(),
		Timestamp:    time.Now().UTC(),
		Agent:        agent,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		SNScore:      snScore,
		GovernanceTier: models.GetGovernanceTier(snScore).Tier,
		Result:       "success",
	}

	// Compute this entry's data hash
	entry.DataHash = al.computeDataHash(entry)

	// Set previous entry's hash if this is not the first entry
	if len(al.entries) > 0 {
		prevEntry := al.entries[len(al.entries)-1]
		entry.PreviousHash = prevEntry.DataHash
	}

	// Sign: HMAC(previous_hash + current_hash)
	entry.Signature = al.computeSignature(entry.PreviousHash, entry.DataHash)

	al.entries = append(al.entries, entry)
	slog.Debug("audit entry created",
		"id", entry.ID,
		"agent", entry.Agent,
		"action", entry.Action,
		"governance_tier", entry.GovernanceTier,
	)

	return entry, nil
}

// QueryAuditTrail retrieves all audit entries for a resource
func (al *HashChainLogger) QueryAuditTrail(resourceType, resourceID string) []*models.AuditEntry {
	var results []*models.AuditEntry
	for _, entry := range al.entries {
		if entry.ResourceType == resourceType && entry.ResourceID == resourceID {
			results = append(results, entry)
		}
	}
	return results
}

// VerifyChainIntegrity verifies the entire hash chain is valid (tamper-detection)
func (al *HashChainLogger) VerifyChainIntegrity() (bool, []string) {
	var issues []string

	for i, entry := range al.entries {
		// Verify data hash
		expectedDataHash := al.computeDataHash(entry)
		if expectedDataHash != entry.DataHash {
			issues = append(issues, "entry "+entry.ID+" data hash mismatch")
		}

		// Verify signature
		expectedSig := al.computeSignature(entry.PreviousHash, entry.DataHash)
		if expectedSig != entry.Signature {
			issues = append(issues, "entry "+entry.ID+" signature invalid")
		}

		// Verify chain link (previous hash reference)
		if i > 0 {
			prevEntry := al.entries[i-1]
			if entry.PreviousHash != prevEntry.DataHash {
				issues = append(issues, "entry "+entry.ID+" chain link broken")
			}
		} else if entry.PreviousHash != "" {
			issues = append(issues, "entry "+entry.ID+" first entry should have empty previous_hash")
		}
	}

	return len(issues) == 0, issues
}

// computeDataHash creates SHA256(agent+action+resourceType+resourceID+timestamp)
func (al *HashChainLogger) computeDataHash(entry *models.AuditEntry) string {
	data := entry.Agent + entry.Action + entry.ResourceType + entry.ResourceID + entry.Timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// computeSignature creates HMAC-SHA256(previousHash + dataHash, secret)
func (al *HashChainLogger) computeSignature(previousHash, dataHash string) string {
	message := previousHash + dataHash
	sig := hmac.New(sha256.New, []byte(al.secret))
	sig.Write([]byte(message))
	return hex.EncodeToString(sig.Sum(nil))
}

// A2AAuditMiddleware is a Gin middleware that logs all A2A requests to audit trail
func A2AAuditMiddleware(logger *HashChainLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract agent ID from headers
		agentID := c.GetHeader("X-Agent-ID")
		if agentID == "" {
			agentID = "unknown-agent"
		}

		// Continue to next handler
		c.Next()

		// Log the request after handler completes
		if c.Request.Method != "GET" {
			action := c.Request.URL.Path
			if c.Request.Method == "POST" {
				action = "create_" + extractResourceType(c.Request.URL.Path)
			} else if c.Request.Method == "PUT" {
				action = "update_" + extractResourceType(c.Request.URL.Path)
			}

			resourceType := extractResourceType(c.Request.URL.Path)
			resourceID := c.Param("id")
			if resourceID == "" {
				resourceID = "unknown"
			}

			snScore := 0.85 // Default score
			if score, exists := c.Get("sn_score"); exists {
				if s, ok := score.(float64); ok {
					snScore = s
				}
			}

			_, _ = logger.LogA2ACall(agentID, action, resourceType, resourceID, snScore)
		}
	}
}

// extractResourceType extracts resource type from URL path
func extractResourceType(path string) string {
	// /api/integrations/a2a/crm/deals -> "deal"
	if len(path) > 0 {
		// Simple extraction: last path segment
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' {
				result := path[i+1:]
				if result != "" && result != "api" && result != "integrations" {
					return result
				}
				break
			}
		}
	}
	return "resource"
}
