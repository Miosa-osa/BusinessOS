package optimal

import "strings"

// GetSkeleton returns the genre template for a classified signal. Templates are
// the canonical genre skeletons from the OptimalOS Signal Theory spec. They give
// the agent the structural scaffold to fill in — not finished documents.
//
// Supported genres: transcript, brief, spec, plan, note, decision-log, standup,
// review, report, pitch. Unknown genres return the note skeleton.
func GetSkeleton(genre string) string {
	switch strings.ToLower(strings.TrimSpace(genre)) {
	case "transcript":
		return skeletonTranscript
	case "brief":
		return skeletonBrief
	case "spec":
		return skeletonSpec
	case "plan":
		return skeletonPlan
	case "note":
		return skeletonNote
	case "decision-log", "decision_log", "decisionlog":
		return skeletonDecisionLog
	case "standup":
		return skeletonStandup
	case "review":
		return skeletonReview
	case "report":
		return skeletonReport
	case "pitch":
		return skeletonPitch
	default:
		return skeletonNote
	}
}

// ── genre skeletons ───────────────────────────────────────────────────────────
// Each skeleton uses {{PLACEHOLDER}} tokens for agent fill-in.
// Sections match the canonical definitions in CLAUDE.md genre catalogue.

const skeletonTranscript = `---
genre: transcript
date: {{DATE}}
participants: {{PARTICIPANTS}}
---

## Participants
{{PARTICIPANTS}}

## Key Points
- {{POINT_1}}
- {{POINT_2}}

## Decisions Made
- {{DECISION}} — decided by {{DECIDER}}

## Action Items
- [ ] {{PERSON}}: {{TASK}} (due {{DATE}})

## Open Questions
- {{QUESTION}}
`

const skeletonBrief = `---
genre: brief
date: {{DATE}}
receiver: {{RECEIVER}}
---

## Objective
{{ONE_SENTENCE_OUTCOME}}

## Key Messages
- {{MESSAGE_1}}
- {{MESSAGE_2}}
- {{MESSAGE_3}}

## Call to Action
{{SINGLE_UNAMBIGUOUS_ASK}} — by {{DEADLINE}}

## Supporting Materials
{{LINKS_OR_ATTACHMENTS}}
`

const skeletonSpec = `---
genre: spec
date: {{DATE}}
owner: {{OWNER}}
status: draft
---

## Goal
{{WHAT_AND_WHY}}

## Requirements
1. {{REQUIREMENT_1}}
2. {{REQUIREMENT_2}}

## Constraints
- {{CONSTRAINT_1}}

## Architecture
{{HOW_IT_FITS}}

## Acceptance Criteria
- [ ] {{CRITERION_1}}
- [ ] {{CRITERION_2}}
`

const skeletonPlan = `---
genre: plan
date: {{DATE}}
week: {{WEEK_OF}}
---

## Objective
{{WHAT_THIS_PLAN_ACHIEVES}}

## Non-Negotiables
1. {{NON_NEG_1}}
2. {{NON_NEG_2}}
3. {{NON_NEG_3}}

## Time Blocks
| Day | Block | Focus |
|-----|-------|-------|
| {{DAY}} | {{TIME}} | {{TASK}} |

## Dependencies
- {{DEPENDENCY}} — waiting on {{PERSON}}

## Success Criteria
- {{METRIC_1}}
- {{METRIC_2}}
`

const skeletonNote = `---
genre: note
date: {{DATE}}
---

## Context
{{ONE_LINE_TRIGGER}}

## Content
{{INFORMATION}}

## Route
{{TARGET_NODE}}
`

const skeletonDecisionLog = `---
genre: decision-log
date: {{DATE}}
decided_by: {{DECIDER}}
---

## Decision
{{WHAT_WAS_DECIDED}}

## Rationale
{{WHY}}

## Decided By
{{DECIDER}} — {{DATE}}

## Impact
{{DOWNSTREAM_EFFECTS}}

## Alternatives Considered
- {{ALTERNATIVE_1}}: rejected because {{REASON}}
`

const skeletonStandup = `---
genre: standup
date: {{DATE}}
---

## Done
- {{COMPLETED_ITEM}}

## Doing
- {{IN_PROGRESS_ITEM}}

## Blocked
- {{BLOCKER}} — need {{WHAT_YOU_NEED}}
`

const skeletonReview = `---
genre: review
date: {{DATE}}
---

## Wins
- {{WIN_1}}

## Issues
- [{{SEVERITY}}] {{ISSUE}}

## Patterns
- {{PATTERN_OBSERVED}}

## Next
- {{NEXT_ACTION}}
`

const skeletonReport = `---
genre: report
date: {{DATE}}
period: {{PERIOD}}
---

## Summary
{{ONE_PARAGRAPH_SUMMARY}}

## Data
{{KEY_METRICS}}

## Analysis
{{INTERPRETATION}}

## Recommendations
1. {{RECOMMENDATION_1}}

## Next Steps
- [ ] {{ACTION}}
`

const skeletonPitch = `---
genre: pitch
date: {{DATE}}
audience: {{AUDIENCE}}
---

## Problem
{{PROBLEM_STATEMENT}}

## Solution
{{WHAT_WE_DO}}

## Evidence
- {{DATA_POINT_1}}
- {{DATA_POINT_2}}

## Ask
{{WHAT_WE_NEED}} — by {{DEADLINE}}
`
