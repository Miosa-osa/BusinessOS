package services

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LearningService handles self-learning, feedback processing, and personalization
type LearningService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewLearningService creates a new learning service
func NewLearningService(pool *pgxpool.Pool) *LearningService {
	return &LearningService{
		pool:   pool,
		logger: slog.Default().With("service", "learning"),
	}
}

// ============================================================================
// Types
// ============================================================================

// LearningEvent represents something the system learned
type LearningEvent struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                string     `json:"user_id"`
	LearningType          string     `json:"learning_type"`
	LearningContent       string     `json:"learning_content"`
	LearningSummary       string     `json:"learning_summary,omitempty"`
	SourceType            string     `json:"source_type"`
	SourceID              *uuid.UUID `json:"source_id,omitempty"`
	SourceContext         string     `json:"source_context,omitempty"`
	ConfidenceScore       float64    `json:"confidence_score"`
	TimesApplied          int        `json:"times_applied"`
	LastAppliedAt         *time.Time `json:"last_applied_at,omitempty"`
	SuccessfulApplications int       `json:"successful_applications"`
	CreatedMemoryID       *uuid.UUID `json:"created_memory_id,omitempty"`
	CreatedFactKey        string     `json:"created_fact_key,omitempty"`
	Category              string     `json:"category,omitempty"`
	Tags                  []string   `json:"tags,omitempty"`
	WasValidated          bool       `json:"was_validated"`
	ValidatedAt           *time.Time `json:"validated_at,omitempty"`
	ValidationResult      string     `json:"validation_result,omitempty"`
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// BehaviorPattern represents an observed user behavior pattern
type BehaviorPattern struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   string     `json:"user_id"`
	PatternType              string     `json:"pattern_type"`
	PatternKey               string     `json:"pattern_key"`
	PatternValue             string     `json:"pattern_value"`
	PatternDescription       string     `json:"pattern_description,omitempty"`
	ObservationCount         int        `json:"observation_count"`
	FirstObservedAt          time.Time  `json:"first_observed_at"`
	LastObservedAt           time.Time  `json:"last_observed_at"`
	ConfidenceScore          float64    `json:"confidence_score"`
	MinObservationsForConfidence int    `json:"min_observations_for_confidence"`
	IsApplied                bool       `json:"is_applied"`
	AppliedInPrompt          bool       `json:"applied_in_prompt"`
	IsActive                 bool       `json:"is_active"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// FeedbackEntry represents user feedback on AI output
type FeedbackEntry struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             string     `json:"user_id"`
	TargetType         string     `json:"target_type"`
	TargetID           uuid.UUID  `json:"target_id"`
	FeedbackType       string     `json:"feedback_type"`
	FeedbackValue      string     `json:"feedback_value,omitempty"`
	Rating             *int       `json:"rating,omitempty"`
	ConversationID     *uuid.UUID `json:"conversation_id,omitempty"`
	AgentType          string     `json:"agent_type,omitempty"`
	FocusMode          string     `json:"focus_mode,omitempty"`
	OriginalContent    string     `json:"original_content,omitempty"`
	ExpectedContent    string     `json:"expected_content,omitempty"`
	WasProcessed       bool       `json:"was_processed"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty"`
	ResultingLearningID *uuid.UUID `json:"resulting_learning_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// PersonalizationProfile contains aggregated user preferences
type PersonalizationProfile struct {
	ID                    uuid.UUID       `json:"id"`
	UserID                string          `json:"user_id"`
	PreferredTone         string          `json:"preferred_tone"`
	PreferredVerbosity    string          `json:"preferred_verbosity"`
	PreferredFormat       string          `json:"preferred_format"`
	PrefersExamples       bool            `json:"prefers_examples"`
	PrefersAnalogies      bool            `json:"prefers_analogies"`
	PrefersCodeSamples    bool            `json:"prefers_code_samples"`
	PrefersVisualAids     bool            `json:"prefers_visual_aids"`
	ExpertiseAreas        []string        `json:"expertise_areas,omitempty"`
	LearningAreas         []string        `json:"learning_areas,omitempty"`
	CommonTopics          []string        `json:"common_topics,omitempty"`
	Timezone              string          `json:"timezone,omitempty"`
	PreferredWorkingHours map[string]any  `json:"preferred_working_hours,omitempty"`
	MostActiveHours       []int           `json:"most_active_hours,omitempty"`
	TotalConversations    int             `json:"total_conversations"`
	TotalFeedbackGiven    int             `json:"total_feedback_given"`
	PositiveFeedbackRatio float64         `json:"positive_feedback_ratio"`
	ProfileCompleteness   float64         `json:"profile_completeness"`
	LastProfileUpdate     time.Time       `json:"last_profile_update"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// FeedbackInput represents input for recording feedback
type FeedbackInput struct {
	UserID          string
	TargetType      string // 'message', 'artifact', 'memory', 'suggestion', 'agent_response'
	TargetID        uuid.UUID
	FeedbackType    string // 'thumbs_up', 'thumbs_down', 'correction', 'comment', 'rating'
	FeedbackValue   string
	Rating          *int
	ConversationID  *uuid.UUID
	AgentType       string
	FocusMode       string
	OriginalContent string
	ExpectedContent string
}

// ============================================================================
// Feedback Processing
// ============================================================================

// RecordFeedback records user feedback and triggers learning
func (s *LearningService) RecordFeedback(ctx context.Context, input FeedbackInput) (*FeedbackEntry, error) {
	feedback := &FeedbackEntry{
		ID:              uuid.New(),
		UserID:          input.UserID,
		TargetType:      input.TargetType,
		TargetID:        input.TargetID,
		FeedbackType:    input.FeedbackType,
		FeedbackValue:   input.FeedbackValue,
		Rating:          input.Rating,
		ConversationID:  input.ConversationID,
		AgentType:       input.AgentType,
		FocusMode:       input.FocusMode,
		OriginalContent: input.OriginalContent,
		ExpectedContent: input.ExpectedContent,
		WasProcessed:    false,
		CreatedAt:       time.Now(),
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO feedback_log (
			id, user_id, target_type, target_id, feedback_type, feedback_value, rating,
			conversation_id, agent_type, focus_mode, original_content, expected_content,
			was_processed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, feedback.ID, feedback.UserID, feedback.TargetType, feedback.TargetID,
		feedback.FeedbackType, feedback.FeedbackValue, feedback.Rating,
		feedback.ConversationID, feedback.AgentType, feedback.FocusMode,
		feedback.OriginalContent, feedback.ExpectedContent, feedback.WasProcessed, feedback.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to record feedback: %w", err)
	}

	// Process feedback asynchronously
	go s.processFeedback(feedback)

	// Update personalization stats
	go s.updateFeedbackStats(input.UserID, input.FeedbackType)

	return feedback, nil
}

// processFeedback processes feedback to extract learnings
func (s *LearningService) processFeedback(feedback *FeedbackEntry) {
	ctx := context.Background()

	var learningID *uuid.UUID

	switch feedback.FeedbackType {
	case "thumbs_down", "correction":
		// This indicates the AI did something wrong - learn from it
		learning, err := s.createLearningFromCorrection(ctx, feedback)
		if err != nil {
			s.logger.Error("failed to create learning from correction", "error", err)
		} else if learning != nil {
			learningID = &learning.ID
		}

	case "thumbs_up":
		// Positive feedback - reinforce the pattern
		s.reinforcePattern(ctx, feedback)

	case "rating":
		// Rating feedback - update preferences based on rating
		if feedback.Rating != nil && *feedback.Rating >= 4 {
			s.reinforcePattern(ctx, feedback)
		} else if feedback.Rating != nil && *feedback.Rating <= 2 {
			learning, _ := s.createLearningFromCorrection(ctx, feedback)
			if learning != nil {
				learningID = &learning.ID
			}
		}
	}

	// Mark feedback as processed
	now := time.Now()
	s.pool.Exec(ctx, `
		UPDATE feedback_log SET was_processed = true, processed_at = $1, resulting_learning_id = $2 WHERE id = $3
	`, now, learningID, feedback.ID)
}

// createLearningFromCorrection creates a learning event from negative feedback
func (s *LearningService) createLearningFromCorrection(ctx context.Context, feedback *FeedbackEntry) (*LearningEvent, error) {
	// Determine learning type
	learningType := "correction"
	if feedback.FeedbackType == "thumbs_down" {
		learningType = "feedback"
	}

	// Generate summary
	summary := fmt.Sprintf("User indicated dissatisfaction with %s response", feedback.AgentType)
	if feedback.ExpectedContent != "" {
		summary = fmt.Sprintf("User expected: %s", truncate(feedback.ExpectedContent, 200))
	}

	learning := &LearningEvent{
		ID:              uuid.New(),
		UserID:          feedback.UserID,
		LearningType:    learningType,
		LearningContent: feedback.FeedbackValue,
		LearningSummary: summary,
		SourceType:      "explicit_feedback",
		SourceID:        &feedback.ID,
		SourceContext:   fmt.Sprintf("Agent: %s, Focus: %s", feedback.AgentType, feedback.FocusMode),
		ConfidenceScore: 0.8, // High confidence for explicit feedback
		Category:        feedback.AgentType,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO learning_events (
			id, user_id, learning_type, learning_content, learning_summary,
			source_type, source_id, source_context, confidence_score, category,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, learning.ID, learning.UserID, learning.LearningType, learning.LearningContent,
		learning.LearningSummary, learning.SourceType, learning.SourceID, learning.SourceContext,
		learning.ConfidenceScore, learning.Category, learning.IsActive, learning.CreatedAt, learning.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return learning, nil
}

// reinforcePattern reinforces positive patterns
func (s *LearningService) reinforcePattern(ctx context.Context, feedback *FeedbackEntry) {
	// Record as positive learning
	learning := &LearningEvent{
		ID:              uuid.New(),
		UserID:          feedback.UserID,
		LearningType:    "pattern",
		LearningContent: "Positive feedback received",
		LearningSummary: fmt.Sprintf("User approved %s response style", feedback.AgentType),
		SourceType:      "explicit_feedback",
		SourceID:        &feedback.ID,
		SourceContext:   fmt.Sprintf("Agent: %s, Focus: %s", feedback.AgentType, feedback.FocusMode),
		ConfidenceScore: 0.7,
		Category:        feedback.AgentType,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	s.pool.Exec(ctx, `
		INSERT INTO learning_events (
			id, user_id, learning_type, learning_content, learning_summary,
			source_type, source_id, source_context, confidence_score, category,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, learning.ID, learning.UserID, learning.LearningType, learning.LearningContent,
		learning.LearningSummary, learning.SourceType, learning.SourceID, learning.SourceContext,
		learning.ConfidenceScore, learning.Category, learning.IsActive, learning.CreatedAt, learning.UpdatedAt)
}

// ============================================================================
// Behavior Pattern Detection
// ============================================================================

// ObserveBehavior records a behavior observation
func (s *LearningService) ObserveBehavior(ctx context.Context, userID, patternType, patternKey, patternValue string) error {
	// Try to update existing pattern
	result, err := s.pool.Exec(ctx, `
		UPDATE user_behavior_patterns
		SET pattern_value = $4,
		    pattern_description = COALESCE(pattern_description, $5),
		    observation_count = observation_count + 1,
		    last_observed_at = NOW(),
		    confidence_score = LEAST(1.0, (observation_count + 1)::float / min_observations_for_confidence),
		    updated_at = NOW()
		WHERE user_id = $1 AND pattern_type = $2 AND pattern_key = $3
	`, userID, patternType, patternKey, patternValue, fmt.Sprintf("Observed %s: %s", patternType, patternKey))

	if err != nil {
		return err
	}

	// If no existing pattern, create new one
	if result.RowsAffected() == 0 {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO user_behavior_patterns (
				id, user_id, pattern_type, pattern_key, pattern_value, pattern_description,
				observation_count, first_observed_at, last_observed_at, confidence_score,
				is_active, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, 1, NOW(), NOW(), 0.33, true, NOW(), NOW())
		`, uuid.New(), userID, patternType, patternKey, patternValue,
			fmt.Sprintf("Observed %s: %s", patternType, patternKey))
	}

	return nil
}

func normalizeUserFactKey(raw string) string {
	key := strings.TrimSpace(raw)
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "\t", "_")
	key = strings.ReplaceAll(key, "\n", "_")
	key = strings.ReplaceAll(key, ":", "-")
	key = strings.ReplaceAll(key, "/", "-")
	key = strings.ReplaceAll(key, "\\", "-")
	if len(key) <= 255 {
		return key
	}
	h := sha1.Sum([]byte(key))
	shortHash := fmt.Sprintf("%x", h)[:10]
	maxPrefix := 255 - 1 - len(shortHash)
	if maxPrefix < 0 {
		maxPrefix = 0
	}
	return key[:maxPrefix] + "-" + shortHash
}

func (s *LearningService) upsertPatternAsUserFact(ctx context.Context, userID string, p BehaviorPattern) error {
	// Store as an inactive-by-default user fact so users can Confirm/Reject via the existing UI.
	factKey := normalizeUserFactKey(fmt.Sprintf("pattern:%s:%s", p.PatternType, p.PatternKey))

	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_facts (user_id, fact_key, fact_value, fact_type, confidence_score, is_active)
		VALUES ($1, $2, $3, 'pattern', $4, false)
		ON CONFLICT (user_id, fact_key)
		DO UPDATE SET
			fact_value = EXCLUDED.fact_value,
			fact_type = EXCLUDED.fact_type,
			confidence_score = EXCLUDED.confidence_score,
			updated_at = NOW()
	`, userID, factKey, p.PatternValue, p.ConfidenceScore)
	return err
}

// DetectPatternsToUserFacts runs pattern detection and persists high-confidence patterns into user_facts.
// Patterns are saved as inactive by default so they can be explicitly confirmed by the user.
func (s *LearningService) DetectPatternsToUserFacts(ctx context.Context, userID string) ([]BehaviorPattern, error) {
	patterns, err := s.DetectPatterns(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, p := range patterns {
		_ = s.upsertPatternAsUserFact(ctx, userID, p) // best-effort
	}

	return patterns, nil
}

// BackfillRecentUsersBehaviorPatterns finds users with recent activity and persists detected patterns into user_facts.
// Returns (usersProcessed, factsUpserted).
func (s *LearningService) BackfillRecentUsersBehaviorPatterns(ctx context.Context, userLimit int) (int, int, error) {
	if userLimit <= 0 {
		userLimit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT user_id
		FROM conversations
		WHERE created_at > NOW() - INTERVAL '30 days'
		ORDER BY user_id
		LIMIT $1
	`, userLimit)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	usersProcessed := 0
	factsUpserted := 0
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			continue
		}
		usersProcessed++

		patterns, err := s.DetectPatterns(ctx, uid)
		if err != nil {
			continue
		}
		for _, p := range patterns {
			if err := s.upsertPatternAsUserFact(ctx, uid, p); err == nil {
				factsUpserted++
			}
		}
	}

	return usersProcessed, factsUpserted, nil
}

// DetectPatterns analyzes user behavior to detect patterns
func (s *LearningService) DetectPatterns(ctx context.Context, userID string) ([]BehaviorPattern, error) {
	// Detect time preferences
	s.detectTimePatterns(ctx, userID)

	// Detect topic interests
	s.detectTopicPatterns(ctx, userID)

	// Detect communication preferences
	s.detectCommunicationPatterns(ctx, userID)

	// Return high-confidence patterns
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, pattern_type, pattern_key, pattern_value, pattern_description,
		       observation_count, first_observed_at, last_observed_at, confidence_score,
		       is_applied, applied_in_prompt, is_active, created_at, updated_at
		FROM user_behavior_patterns
		WHERE user_id = $1 AND is_active = true AND confidence_score >= 0.6
		ORDER BY confidence_score DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []BehaviorPattern
	for rows.Next() {
		var p BehaviorPattern
		err := rows.Scan(&p.ID, &p.UserID, &p.PatternType, &p.PatternKey, &p.PatternValue,
			&p.PatternDescription, &p.ObservationCount, &p.FirstObservedAt, &p.LastObservedAt,
			&p.ConfidenceScore, &p.IsApplied, &p.AppliedInPrompt, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}

// detectTimePatterns detects when user is most active
func (s *LearningService) detectTimePatterns(ctx context.Context, userID string) {
	// Analyze conversation start times
	rows, err := s.pool.Query(ctx, `
		SELECT EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as count
		FROM conversations
		WHERE user_id = $1 AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY hour
		ORDER BY count DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	var activeHours []int
	for rows.Next() {
		var hour, count int
		if rows.Scan(&hour, &count) == nil && count >= 3 {
			activeHours = append(activeHours, hour)
		}
	}

	if len(activeHours) > 0 {
		s.ObserveBehavior(ctx, userID, "time_preference", "active_hours", fmt.Sprintf("%v", activeHours))
	}
}

// detectTopicPatterns detects common topics
func (s *LearningService) detectTopicPatterns(ctx context.Context, userID string) {
	// This would analyze conversation content for common themes
	// Simplified version using project names
	rows, err := s.pool.Query(ctx, `
		SELECT name, COUNT(*) as count
		FROM projects
		WHERE user_id = $1
		GROUP BY name
		ORDER BY count DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var count int
		if rows.Scan(&name, &count) == nil {
			s.ObserveBehavior(ctx, userID, "topic_interest", name, "project")
		}
	}
}

// detectCommunicationPatterns detects communication style preferences
func (s *LearningService) detectCommunicationPatterns(ctx context.Context, userID string) {
	// Analyze message lengths to determine verbosity preference
	var avgLength float64
	err := s.pool.QueryRow(ctx, `
		SELECT AVG(LENGTH(content)) FROM messages
		WHERE conversation_id IN (SELECT id FROM conversations WHERE user_id = $1)
		AND role = 'user' AND created_at > NOW() - INTERVAL '30 days'
	`, userID).Scan(&avgLength)

	if err == nil && avgLength > 0 {
		var preference string
		if avgLength < 100 {
			preference = "concise"
		} else if avgLength < 300 {
			preference = "balanced"
		} else {
			preference = "detailed"
		}
		s.ObserveBehavior(ctx, userID, "communication_style", "verbosity", preference)
	}
}

// ============================================================================
// Personalization Profile
// ============================================================================

// GetPersonalizationProfile retrieves or creates a user's personalization profile
func (s *LearningService) GetPersonalizationProfile(ctx context.Context, userID string) (*PersonalizationProfile, error) {
	var profile PersonalizationProfile
	var workingHoursJSON []byte

	// Use nullable wrappers for array types to handle NULL values
	var expertiseAreas, learningAreas, commonTopics []string
	var mostActiveHours []int

	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, preferred_tone, preferred_verbosity, preferred_format,
		       prefers_examples, prefers_analogies, prefers_code_samples, prefers_visual_aids,
		       COALESCE(expertise_areas, '{}'), COALESCE(learning_areas, '{}'), COALESCE(common_topics, '{}'),
		       COALESCE(timezone, ''), preferred_working_hours,
		       COALESCE(most_active_hours, '{}'), total_conversations, total_feedback_given, positive_feedback_ratio,
		       profile_completeness, last_profile_update, created_at, updated_at
		FROM personalization_profiles
		WHERE user_id = $1
	`, userID).Scan(
		&profile.ID, &profile.UserID, &profile.PreferredTone, &profile.PreferredVerbosity,
		&profile.PreferredFormat, &profile.PrefersExamples, &profile.PrefersAnalogies,
		&profile.PrefersCodeSamples, &profile.PrefersVisualAids, &expertiseAreas,
		&learningAreas, &commonTopics, &profile.Timezone, &workingHoursJSON,
		&mostActiveHours, &profile.TotalConversations, &profile.TotalFeedbackGiven,
		&profile.PositiveFeedbackRatio, &profile.ProfileCompleteness, &profile.LastProfileUpdate,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	// Assign arrays after scan
	profile.ExpertiseAreas = expertiseAreas
	profile.LearningAreas = learningAreas
	profile.CommonTopics = commonTopics
	profile.MostActiveHours = mostActiveHours

	if err == pgx.ErrNoRows {
		// Create default profile
		profile = PersonalizationProfile{
			ID:                  uuid.New(),
			UserID:              userID,
			PreferredTone:       "professional",
			PreferredVerbosity:  "balanced",
			PreferredFormat:     "structured",
			PrefersExamples:     true,
			PrefersAnalogies:    false,
			PrefersCodeSamples:  false,
			PrefersVisualAids:   false,
			ProfileCompleteness: 0.1,
			LastProfileUpdate:   time.Now(),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		_, err = s.pool.Exec(ctx, `
			INSERT INTO personalization_profiles (
				id, user_id, preferred_tone, preferred_verbosity, preferred_format,
				prefers_examples, prefers_analogies, prefers_code_samples, prefers_visual_aids,
				profile_completeness, last_profile_update, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, profile.ID, profile.UserID, profile.PreferredTone, profile.PreferredVerbosity,
			profile.PreferredFormat, profile.PrefersExamples, profile.PrefersAnalogies,
			profile.PrefersCodeSamples, profile.PrefersVisualAids, profile.ProfileCompleteness,
			profile.LastProfileUpdate, profile.CreatedAt, profile.UpdatedAt)

		if err != nil {
			return nil, err
		}

		return &profile, nil
	}

	if err != nil {
		return nil, err
	}

	if workingHoursJSON != nil {
		json.Unmarshal(workingHoursJSON, &profile.PreferredWorkingHours)
	}

	return &profile, nil
}

// UpdatePersonalizationProfile updates or creates a user's profile (UPSERT)
func (s *LearningService) UpdatePersonalizationProfile(ctx context.Context, profile *PersonalizationProfile) error {
	workingHoursJSON, _ := json.Marshal(profile.PreferredWorkingHours)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO personalization_profiles (
			user_id, preferred_tone, preferred_verbosity, preferred_format,
			prefers_examples, prefers_analogies, prefers_code_samples, prefers_visual_aids,
			expertise_areas, learning_areas, common_topics, timezone,
			preferred_working_hours, most_active_hours, profile_completeness,
			last_profile_update, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW(), NOW()
		)
		ON CONFLICT (user_id) DO UPDATE SET
			preferred_tone = EXCLUDED.preferred_tone,
			preferred_verbosity = EXCLUDED.preferred_verbosity,
			preferred_format = EXCLUDED.preferred_format,
			prefers_examples = EXCLUDED.prefers_examples,
			prefers_analogies = EXCLUDED.prefers_analogies,
			prefers_code_samples = EXCLUDED.prefers_code_samples,
			prefers_visual_aids = EXCLUDED.prefers_visual_aids,
			expertise_areas = EXCLUDED.expertise_areas,
			learning_areas = EXCLUDED.learning_areas,
			common_topics = EXCLUDED.common_topics,
			timezone = EXCLUDED.timezone,
			preferred_working_hours = EXCLUDED.preferred_working_hours,
			most_active_hours = EXCLUDED.most_active_hours,
			profile_completeness = EXCLUDED.profile_completeness,
			last_profile_update = NOW(),
			updated_at = NOW()
	`, profile.UserID, profile.PreferredTone, profile.PreferredVerbosity, profile.PreferredFormat,
		profile.PrefersExamples, profile.PrefersAnalogies, profile.PrefersCodeSamples,
		profile.PrefersVisualAids, profile.ExpertiseAreas, profile.LearningAreas,
		profile.CommonTopics, profile.Timezone, workingHoursJSON, profile.MostActiveHours,
		profile.ProfileCompleteness)

	return err
}

// RefreshProfileFromPatterns updates profile based on detected patterns
func (s *LearningService) RefreshProfileFromPatterns(ctx context.Context, userID string) error {
	profile, err := s.GetPersonalizationProfile(ctx, userID)
	if err != nil {
		return err
	}

	// Get high-confidence patterns
	patterns, err := s.DetectPatterns(ctx, userID)
	if err != nil {
		return err
	}

	// Update profile based on patterns
	for _, p := range patterns {
		switch p.PatternType {
		case "communication_style":
			if p.PatternKey == "verbosity" {
				profile.PreferredVerbosity = p.PatternValue
			}
		case "time_preference":
			if p.PatternKey == "active_hours" {
				// Parse hours
				var hours []int
				json.Unmarshal([]byte(p.PatternValue), &hours)
				profile.MostActiveHours = hours
			}
		case "topic_interest":
			if !contains(profile.CommonTopics, p.PatternKey) {
				profile.CommonTopics = append(profile.CommonTopics, p.PatternKey)
			}
		}
	}

	// Calculate profile completeness
	profile.ProfileCompleteness = s.calculateCompleteness(profile)

	return s.UpdatePersonalizationProfile(ctx, profile)
}

// calculateCompleteness calculates how complete a profile is
func (s *LearningService) calculateCompleteness(profile *PersonalizationProfile) float64 {
	var score float64
	total := 10.0

	if profile.PreferredTone != "" {
		score++
	}
	if profile.PreferredVerbosity != "" {
		score++
	}
	if profile.PreferredFormat != "" {
		score++
	}
	if len(profile.ExpertiseAreas) > 0 {
		score++
	}
	if len(profile.CommonTopics) > 0 {
		score++
	}
	if profile.Timezone != "" {
		score++
	}
	if len(profile.MostActiveHours) > 0 {
		score++
	}
	if profile.TotalConversations > 10 {
		score++
	}
	if profile.TotalFeedbackGiven > 5 {
		score++
	}
	if profile.PositiveFeedbackRatio > 0 {
		score++
	}

	return score / total
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *LearningService) updateFeedbackStats(userID, feedbackType string) {
	ctx := context.Background()

	// Update total feedback count
	s.pool.Exec(ctx, `
		UPDATE personalization_profiles
		SET total_feedback_given = total_feedback_given + 1,
		    updated_at = NOW()
		WHERE user_id = $1
	`, userID)

	// Update positive ratio if thumbs_up
	if feedbackType == "thumbs_up" {
		s.pool.Exec(ctx, `
			UPDATE personalization_profiles
			SET positive_feedback_ratio = (
				SELECT COALESCE(
					COUNT(*) FILTER (WHERE feedback_type = 'thumbs_up')::float /
					NULLIF(COUNT(*), 0), 0
				)
				FROM feedback_log WHERE user_id = $1
			),
			updated_at = NOW()
			WHERE user_id = $1
		`, userID)
	}
}

// GetLearningsForContext retrieves relevant learnings for a context
func (s *LearningService) GetLearningsForContext(ctx context.Context, userID, agentType string, limit int) ([]LearningEvent, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, learning_type, learning_content, learning_summary,
		       source_type, source_id, source_context, confidence_score,
		       times_applied, last_applied_at, successful_applications,
		       category, tags, is_active, created_at, updated_at
		FROM learning_events
		WHERE user_id = $1 AND is_active = true
		AND (category = $2 OR category IS NULL)
		ORDER BY confidence_score DESC, created_at DESC
		LIMIT $3
	`, userID, agentType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var learnings []LearningEvent
	for rows.Next() {
		var l LearningEvent
		err := rows.Scan(&l.ID, &l.UserID, &l.LearningType, &l.LearningContent, &l.LearningSummary,
			&l.SourceType, &l.SourceID, &l.SourceContext, &l.ConfidenceScore,
			&l.TimesApplied, &l.LastAppliedAt, &l.SuccessfulApplications,
			&l.Category, &l.Tags, &l.IsActive, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			continue
		}
		learnings = append(learnings, l)
	}

	return learnings, nil
}

// ApplyLearning marks a learning as applied
func (s *LearningService) ApplyLearning(ctx context.Context, learningID uuid.UUID, successful bool) error {
	successIncrement := 0
	if successful {
		successIncrement = 1
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE learning_events
		SET times_applied = times_applied + 1,
		    successful_applications = successful_applications + $1,
		    last_applied_at = NOW(),
		    confidence_score = LEAST(1.0, confidence_score + CASE WHEN $2 THEN 0.05 ELSE -0.05 END),
		    updated_at = NOW()
		WHERE id = $3
	`, successIncrement, successful, learningID)

	return err
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
