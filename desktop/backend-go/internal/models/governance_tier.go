package models

// GovernanceTier defines the approval requirements based on Signal-to-Noise confidence score.
// This implements a three-level governance model: auto (high confidence), human (medium),
// and board (low confidence), following Signal Theory S=(M,G,T,F,W).
type GovernanceTier struct {
	// SNScore is the lower bound of the S/N score for this tier
	SNScore float64 `json:"sn_score"`

	// Tier is the tier name: "auto", "human", or "board"
	Tier string `json:"tier"`

	// RequiresApproval indicates if human approval is mandatory
	RequiresApproval bool `json:"requires_approval"`

	// ApprovalRole is the minimum role required for approval
	// - "auto": no approval needed
	// - "human": manager, director, or above
	// - "board": ceo, cfo, or board member
	ApprovalRole string `json:"approval_role"`

	// Description explains the tier's purpose
	Description string `json:"description"`
}

// GovernanceTierConfig provides the complete tier matrix.
// Used to route A2A decisions based on agent confidence (S/N score).
var GovernanceTierConfig = []GovernanceTier{
	{
		SNScore:          0.8,
		Tier:             "auto",
		RequiresApproval: false,
		ApprovalRole:     "none",
		Description:      "High confidence: no approval needed, auto-execute",
	},
	{
		SNScore:          0.7,
		Tier:             "human",
		RequiresApproval: true,
		ApprovalRole:     "manager",
		Description:      "Medium confidence: manager approval required",
	},
	{
		SNScore:          0.0,
		Tier:             "board",
		RequiresApproval: true,
		ApprovalRole:     "ceo",
		Description:      "Low confidence: C-level / board approval required",
	},
}

// GetGovernanceTier returns the appropriate tier for a given S/N score.
func GetGovernanceTier(snScore float64) GovernanceTier {
	for _, tier := range GovernanceTierConfig {
		if snScore >= tier.SNScore {
			return tier
		}
	}
	return GovernanceTierConfig[len(GovernanceTierConfig)-1]
}

// ApprovalRequired checks if a governance tier requires approval.
func ApprovalRequired(snScore float64) bool {
	return snScore < 0.8
}

// RequiredApprovalRole returns the minimum role needed to approve at this S/N score.
func RequiredApprovalRole(snScore float64) string {
	tier := GetGovernanceTier(snScore)
	return tier.ApprovalRole
}
