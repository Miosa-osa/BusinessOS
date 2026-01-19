# Pedro Tasks: Learning & Personalization

> **Priority:** P0 - Critical (Pedro's Work)
> **Backend Status:** Complete (8 endpoints)
> **Frontend Status:** Not Started
> **Owner:** Pedro
> **Estimated Effort:** 1-2 sprints

---

## Overview

The Learning module enables AI personalization based on user feedback and behavior patterns. The system learns user preferences and adapts AI responses accordingly.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/learning/feedback` | Record user feedback on AI |
| POST | `/api/learning/behavior` | Log user behavior patterns |
| GET | `/api/learning/profile` | Get personalization profile |
| PUT | `/api/learning/profile` | Update profile preferences |
| POST | `/api/learning/profile/refresh` | Refresh from detected patterns |
| GET | `/api/learning/patterns` | Get detected behavior patterns |
| GET | `/api/learning/learnings` | Get learnings for agent/context |
| POST | `/api/learning/learnings/:id/apply` | Mark learning as applied |

---

## Data Models

### Personalization Profile
```typescript
interface PersonalizationProfile {
  user_id: string;

  // Communication Style
  preferred_tone: 'professional' | 'casual' | 'technical' | 'friendly';
  preferred_verbosity: 'brief' | 'balanced' | 'detailed';
  preferred_format: 'structured' | 'narrative' | 'visual';

  // Learning Preferences
  learning_preferences: {
    examples: boolean;
    analogies: boolean;
    code_samples: boolean;
    visual_aids: boolean;
    step_by_step: boolean;
  };

  // Expertise
  expertise_areas: string[];
  learning_areas: string[];

  // Activity
  most_active_hours: number[];  // 0-23
  total_conversations: number;
  feedback_given: number;
  positive_feedback_ratio: number;

  // Meta
  profile_completeness: number;  // 0-100
  last_updated: string;
}
```

### Feedback
```typescript
interface Feedback {
  id: string;
  user_id: string;

  // Target
  feedback_type: 'thumbs_up' | 'thumbs_down' | 'correction' | 'comment' | 'rating';
  target_type: 'message' | 'artifact' | 'memory' | 'suggestion' | 'agent_response';
  target_id: string;

  // Content
  rating?: number;  // 1-5 for rating type
  comment?: string;
  correction?: string;

  // Context
  conversation_id?: string;
  agent_id?: string;

  created_at: string;
}
```

### Learning
```typescript
interface Learning {
  id: string;
  user_id: string;

  // Content
  content: string;
  learning_type: 'preference' | 'behavior' | 'correction' | 'pattern';
  source: 'feedback' | 'behavior' | 'explicit';

  // Context
  agent_id?: string;
  context_id?: string;

  // Application
  is_applied: boolean;
  applied_count: number;
  last_applied_at?: string;

  // Confidence
  confidence: number;  // 0-1

  created_at: string;
}
```

---

## Frontend Implementation Tasks

### Phase 1: Feedback UI in Chat

#### 1.1 Message Feedback Buttons
**File:** `src/lib/components/chat/MessageFeedback.svelte`

```svelte
<div class="message-feedback">
  <Button
    variant="ghost"
    size="sm"
    on:click={() => submitFeedback('thumbs_up')}
    class:active={currentFeedback === 'thumbs_up'}
  >
    <ThumbsUpIcon />
  </Button>
  <Button
    variant="ghost"
    size="sm"
    on:click={() => submitFeedback('thumbs_down')}
    class:active={currentFeedback === 'thumbs_down'}
  >
    <ThumbsDownIcon />
  </Button>
  <Button
    variant="ghost"
    size="sm"
    on:click={() => openFeedbackModal()}
  >
    <MessageIcon /> Feedback
  </Button>
</div>
```

#### 1.2 Detailed Feedback Modal
**File:** `src/lib/components/chat/FeedbackModal.svelte`

- [ ] Rating stars (1-5)
- [ ] Comment textarea
- [ ] Correction input (for wrong responses)
- [ ] Category tags (too long, too short, inaccurate, etc.)

#### 1.3 Inline Correction
- [ ] "This is wrong" button → inline edit
- [ ] Submit correction
- [ ] AI acknowledges and learns

### Phase 2: Personalization Profile

#### 2.1 Profile Settings Page
**File:** `src/routes/(app)/settings/personalization/+page.svelte`

- [ ] **Communication Style Section**
  - Tone selector (professional, casual, technical, friendly)
  - Verbosity selector (brief, balanced, detailed)
  - Format selector (structured, narrative, visual)

- [ ] **Learning Preferences Section**
  - Toggle: Include examples
  - Toggle: Use analogies
  - Toggle: Show code samples
  - Toggle: Visual aids
  - Toggle: Step-by-step explanations

- [ ] **Expertise Section**
  - Expertise areas (multi-select or tags)
  - Learning areas (topics to explain more)

- [ ] **Profile Stats**
  - Total conversations
  - Feedback given
  - Positive ratio
  - Profile completeness bar

#### 2.2 Profile Preview
- [ ] Sample AI response with current settings
- [ ] "Try different settings" feature

### Phase 3: Detected Patterns

#### 3.1 Patterns Dashboard
**File:** `src/routes/(app)/settings/personalization/patterns/+page.svelte`

- [ ] List of detected behavior patterns
- [ ] Pattern details (what was observed)
- [ ] Confirm/reject pattern
- [ ] Convert pattern to preference

#### 3.2 Pattern Card
**File:** `src/lib/components/learning/PatternCard.svelte`

```svelte
<div class="pattern-card">
  <h4>{pattern.description}</h4>
  <p>Detected from {pattern.observation_count} interactions</p>
  <Badge>{pattern.confidence * 100}% confident</Badge>

  <div class="actions">
    <Button on:click={() => applyPattern(pattern)}>
      Apply to Profile
    </Button>
    <Button variant="outline" on:click={() => dismissPattern(pattern)}>
      Dismiss
    </Button>
  </div>
</div>
```

### Phase 4: Learnings Management

#### 4.1 Learnings List
**File:** `src/routes/(app)/settings/personalization/learnings/+page.svelte`

- [ ] List all learnings
- [ ] Filter by: type, source, agent, applied status
- [ ] Show application count
- [ ] Delete learning

#### 4.2 Learning Card
- [ ] Learning content
- [ ] Source information
- [ ] Confidence level
- [ ] Applied count
- [ ] Actions: Apply, Delete

### Phase 5: API Client

#### 5.1 Learning API
**File:** `src/lib/api/learning/learning.ts`

```typescript
// Feedback
export async function submitFeedback(data: FeedbackInput): Promise<Feedback>

// Profile
export async function getPersonalizationProfile(): Promise<PersonalizationProfile>
export async function updatePersonalizationProfile(
  data: Partial<PersonalizationProfile>
): Promise<PersonalizationProfile>
export async function refreshProfile(): Promise<PersonalizationProfile>

// Patterns
export async function getDetectedPatterns(): Promise<Pattern[]>
export async function applyPattern(patternId: string): Promise<void>
export async function dismissPattern(patternId: string): Promise<void>

// Learnings
export async function getLearnings(filters?: LearningFilters): Promise<Learning[]>
export async function applyLearning(id: string): Promise<void>
export async function deleteLearning(id: string): Promise<void>

// Behavior
export async function logBehavior(data: BehaviorInput): Promise<void>
```

#### 5.2 Learning Store
**File:** `src/lib/stores/learning.ts`

```typescript
interface LearningStore {
  profile: PersonalizationProfile | null;
  patterns: Pattern[];
  learnings: Learning[];
  isLoading: boolean;

  loadProfile(): Promise<void>;
  updateProfile(data: Partial<PersonalizationProfile>): Promise<void>;
  submitFeedback(data: FeedbackInput): Promise<void>;
  loadPatterns(): Promise<void>;
  applyPattern(id: string): Promise<void>;
  loadLearnings(): Promise<void>;
}
```

---

## UI/UX Requirements

### Feedback UX
- Non-intrusive feedback buttons
- Quick thumbs up/down (one click)
- Optional detailed feedback
- Confirmation toast

### Profile UX
- Clear explanation of each setting
- Preview of how settings affect AI
- "Reset to defaults" option

### Pattern UX
- Explain why pattern was detected
- Show evidence (observations)
- Easy confirm/reject

---

## Integration Points

### Chat Integration
- Feedback buttons on every AI message
- Profile settings affect all AI responses
- Learnings applied automatically

### System-Wide
- All AI interactions use personalization profile
- Feedback improves all agents
- Learnings persist across sessions

---

## Testing Requirements

- [ ] Unit tests for learning store
- [ ] Component tests for feedback UI
- [ ] E2E: Submit feedback flow
- [ ] E2E: Update profile settings
- [ ] E2E: Apply pattern to profile

---

## Linear Issues to Create

1. **[LEARN-001]** Add feedback buttons to chat messages
2. **[LEARN-002]** Create detailed feedback modal
3. **[LEARN-003]** Build personalization profile page
4. **[LEARN-004]** Implement communication style settings
5. **[LEARN-005]** Add learning preferences toggles
6. **[LEARN-006]** Create detected patterns dashboard
7. **[LEARN-007]** Build pattern card with confirm/reject
8. **[LEARN-008]** Add learnings management page
9. **[LEARN-009]** API client and store
10. **[LEARN-010]** E2E tests

---

## Dependencies

- Ties into Memories/User Facts (03-MEMORIES-USER-FACTS.md)

## Blockers

- None identified

---

## Notes

- This is core to making AI feel "personalized"
- Start with explicit preferences, add auto-learning later
- Consider A/B testing different personalization approaches
- Privacy: Users should be able to delete all learning data
