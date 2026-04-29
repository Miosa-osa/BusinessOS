# Signal Theory — Composition Layer
## The Micro-Structure of Signals: Genre Catalogue & Skeleton Library

**Version:** 1.0
**Date:** 2026-03-16
**Author:** Architect Agent (OSA v3.3)

---

## How to Use This Document

Every Signal has five dimensions: **S = (M, G, T, F, W)**

```
M = MODE      — how is it perceived?      (Linguistic | Visual | Code | Mixed)
G = GENRE     — what conventionalized form?
T = TYPE      — what does it DO?          (Direct | Inform | Commit | Decide | Express)
F = FORMAT    — what container?           (Markdown | Doc | Slides | Video | Chat)
W = STRUCTURE — internal skeleton         (required sections, in order)
```

This document defines **W** for every genre — the mandatory internal skeleton that
transforms raw information into a decodable Signal. Without W, output is noise.

**Granularity levels** appear at the end. Every genre can be rendered at L0–L3.

---

## GRANULARITY MODEL (Universal)

```
L0  HEADLINE    ~10 words    One line. What it is and why it matters.
L1  SUMMARY     ~50 words    One paragraph. Enough to decide whether to read on.
L2  DETAIL      ~500 words   All required sections. Enough to act.
L3  COMPLETE    Unlimited    Full artifact. All sub-sections, appendices, evidence.
```

**Progressive disclosure rule:** An agent or human first sees L0. Drills to L1 if
relevant. Only reads L2/L3 when they need to act or decide. Match output granularity
to the receiver's current need — never over-encode or under-encode.

**Bandwidth matching:**
- Executive scanning a feed → L0
- Manager deciding to prioritize → L1
- Operator implementing → L2
- Approver or auditor → L3

---

## PART I — BUSINESS GENRES

---

### B1. Brief

**Compact header:**
`M: Linguistic | T: Direct | F: Doc/Markdown | Audience: Decision-maker who must authorize scope`

**Sub-genres:** Sales Brief, Project Brief, Creative Brief

**Skeleton (W):**
```
1. ONE-LINE OBJECTIVE     What outcome are we authorizing?
2. CONTEXT                Why now? What problem or opportunity?
3. SCOPE                  What is IN. What is explicitly OUT.
4. SUCCESS CRITERIA       How do we know it worked? (measurable)
5. CONSTRAINTS            Budget, timeline, team, non-negotiables
6. STAKEHOLDERS           Who owns, who approves, who is informed
7. OPEN QUESTIONS         What still needs a decision before work starts
```

**Quality checklist:**
- [ ] Objective is one sentence, not a paragraph
- [ ] Scope has an explicit OUT section (prevents scope creep)
- [ ] Success criteria are measurable (numbers, not adjectives)
- [ ] Every stakeholder has exactly one role (RACI-style)
- [ ] Open questions are actionable (assigned, not rhetorical)
- [ ] Zero jargon the receiver does not share

**Example (L0):** "Launch brief for Q2 outreach campaign — $8K budget, 6-week window, 3 ICP segments."

---

### B2. Proposal

**Compact header:**
`M: Mixed (Linguistic + Visual) | T: Commit | F: Doc/Slides | Audience: External buyer or partner evaluating investment`

**Sub-genres:** Client Proposal, Partnership Proposal

**Skeleton (W):**
```
1. EXECUTIVE SUMMARY      The offer in 3–5 sentences (L1 of the whole document)
2. PROBLEM STATEMENT      What pain or opportunity we are addressing for them
3. PROPOSED SOLUTION      What we will do, deliver, or build
4. SCOPE OF WORK          Deliverables, phases, milestones (no ambiguity)
5. INVESTMENT             Pricing, payment schedule, what is included / excluded
6. TIMELINE               Dates, dependencies, go-live
7. WHY US                 Evidence of fit: case studies, credentials, team
8. TERMS & NEXT STEPS     How to proceed, expiry of offer, signatures
```

**Quality checklist:**
- [ ] Executive summary is self-contained (can stand alone)
- [ ] Problem statement uses the buyer's language, not ours
- [ ] Deliverables are concrete nouns, not verb phrases
- [ ] Pricing has no hidden line items
- [ ] Next steps have a single clear call to action
- [ ] Social proof is specific (client name, result, number)

**Example (L0):** "Proposal to automate Acme Corp's onboarding flow — 6 weeks, $24K, 40% reduction in ramp time."

---

### B3. Report

**Compact header:**
`M: Mixed (Linguistic + Visual) | T: Inform | F: Doc/Dashboard | Audience: Manager or executive making ongoing decisions`

**Sub-genres:** Status Report, Financial Report, Analytics Report

**Skeleton (W):**
```
1. REPORT HEADER          Period covered, author, distribution list, date
2. EXECUTIVE SUMMARY      Key findings in 3–5 bullets (L1 of the document)
3. KEY METRICS            Scorecards, KPIs, delta from prior period
4. WHAT HAPPENED          Factual narrative of the period (events, activities)
5. ANALYSIS               Why the numbers are what they are (root causes)
6. RISKS & ISSUES         What is off-track, what is at risk
7. DECISIONS NEEDED       Explicit asks from the reader
8. NEXT PERIOD OUTLOOK    What to expect, what changes
9. APPENDIX               Raw data, methodology, source tables
```

**Quality checklist:**
- [ ] Every metric has a baseline or target for comparison
- [ ] Analysis section explains cause, not just correlation
- [ ] Risks are rated (probability × impact)
- [ ] Decisions Needed section is explicitly labeled (not buried in narrative)
- [ ] Charts have titles, axis labels, and data source
- [ ] No metric without context (what is good? what is bad?)

**Example (L0):** "March analytics report — 23% revenue growth, CAC up 15%, two pipeline risks flagged."

---

### B4. Spec

**Compact header:**
`M: Linguistic + Code | T: Direct | F: Markdown/Doc | Audience: Implementer who must build to this contract`

**Sub-genres:** Technical Spec, Product Spec, Feature Spec

**Skeleton (W):**
```
1. OVERVIEW               What this spec covers and why it exists
2. GOALS                  What success looks like (measurable outcomes)
3. NON-GOALS              What this spec explicitly does NOT address
4. BACKGROUND             Context the implementer needs to understand the problem
5. REQUIREMENTS           Functional requirements (must/should/could language — MoSCoW)
6. NON-FUNCTIONAL REQ.    Performance, security, reliability, scalability targets
7. DESIGN / ARCHITECTURE  How it will be built (diagrams, data models, API contracts)
8. EDGE CASES             Known failure modes, boundary conditions, error states
9. OPEN QUESTIONS         Unresolved decisions (with owner and deadline)
10. TESTING CRITERIA      How QA will verify this is done
11. DEPENDENCIES          Upstream/downstream systems, teams, data
12. ROLLOUT PLAN          Phasing, feature flags, migration steps
```

**Quality checklist:**
- [ ] Requirements use MoSCoW (must/should/could/won't) — no ambiguous "should"
- [ ] Every non-functional requirement has a number attached
- [ ] Open questions have an owner and a deadline
- [ ] Edge cases section exists and is non-trivial
- [ ] API contracts are explicit (request/response schemas)
- [ ] Non-goals section prevents future scope disputes

**Example (L0):** "Feature spec for bulk CSV import — 10K rows, sub-5s processing, full error report on failure."

---

### B5. Decision Record

**Compact header:**
`M: Linguistic | T: Decide | F: Markdown | Audience: Future team members who need to understand why a choice was made`

**Sub-genres:** Architecture Decision Record (ADR), Business Decision Record, Strategy Pivot Record

**Skeleton (W):**
```
1. TITLE & ID             ADR-NNN: [Decision Title]
2. STATUS                 Proposed | Accepted | Deprecated | Superseded by [ID]
3. DATE                   When this decision was made
4. CONTEXT                What situation forced a decision? (the forcing function)
5. DECISION               What was chosen and the single core reason
6. CONSEQUENCES
   - Positive             Benefits we gain
   - Negative             Trade-offs we accept
   - Neutral              Implications to manage
7. ALTERNATIVES CONSIDERED
   - Option A             Description — Rejected because: [reason]
   - Option B             Description — Rejected because: [reason]
8. REFERENCES             Related decisions, documentation, research
```

**Quality checklist:**
- [ ] Decision statement is one sentence — no hedging
- [ ] Context explains the forcing function (why decision was unavoidable)
- [ ] At least two alternatives considered (not straw men)
- [ ] Negative consequences explicitly stated (not minimized)
- [ ] Status field is current and accurate
- [ ] Can be read standalone without context from the author

**Example (L0):** "ADR-042: Adopt event sourcing for order history — audit trail requirement forces immutable log."

---

### B6. Plan

**Compact header:**
`M: Linguistic + Visual | T: Commit | F: Doc/Table/Board | Audience: Executor and their manager who must coordinate work`

**Sub-genres:** Week Plan, Project Plan, Launch Plan

**Skeleton (W):**
```
1. OBJECTIVE              What done looks like at the end of this plan's horizon
2. SCOPE & ASSUMPTIONS    What is in, what is assumed to be true
3. MILESTONES             Key checkpoints with dates and owners
4. TASK BREAKDOWN         Tasks → subtasks → assignees → due dates → status
5. DEPENDENCIES           What must happen first (internal + external)
6. RISKS                  What could derail this plan (and mitigation)
7. RESOURCES              People, budget, tools, access needed
8. COMMUNICATIONS PLAN    Who gets updates, how often, in what format
9. SUCCESS CRITERIA       How we know the plan was executed successfully
```

**Quality checklist:**
- [ ] Objective is time-bound and measurable
- [ ] Every task has exactly one owner (not "team")
- [ ] Dependencies are explicit (not implied)
- [ ] Risks section is not empty (a plan with no risks is not honest)
- [ ] Milestones are events, not date ranges
- [ ] Plan is living — has a last-updated date

**Example (L0):** "Q2 launch plan — 8 milestones, April 14 go-live, 3 dependencies on platform team."

---

### B7. Review

**Compact header:**
`M: Linguistic | T: Decide | F: Doc/Markdown | Audience: Reviewee and their manager who must act on findings`

**Sub-genres:** Code Review, Performance Review, Retrospective

**Skeleton (W) — Synchronous Review (Code/Doc):**
```
1. OVERALL VERDICT        Approved | Needs Changes | Blocked
2. CRITICAL ISSUES        Must fix before merge/approval (blocking)
3. MAJOR ISSUES           Should fix — significant quality or risk concern
4. MINOR ISSUES           Nice to have — style, clarity, small improvements
5. POSITIVE OBSERVATIONS  What was done well (required — not optional fluff)
6. SUMMARY OF ASKS        Numbered list of every action required from reviewee
```

**Skeleton (W) — Performance / Retrospective:**
```
1. PERIOD & CONTEXT       What timeframe and environment is being evaluated
2. WHAT WENT WELL         Specific behaviors or outcomes (with evidence)
3. WHAT TO IMPROVE        Specific gaps (with evidence, not opinion)
4. RATING / ASSESSMENT    Score or tier with rationale
5. GOALS FOR NEXT PERIOD  2–3 specific, measurable commitments
6. SUPPORT NEEDED         What the manager/team must provide
```

**Quality checklist:**
- [ ] Every issue has a file/line reference (code) or specific example (performance)
- [ ] Feedback is behavioral, not personal
- [ ] Positive observations are specific, not generic praise
- [ ] Action items are numbered and unambiguous
- [ ] Verdict is stated at the top — not buried at the end
- [ ] Severity labels are used consistently

**Example (L0):** "Code review: approved with 2 minor fixes — logic correct, missing null check on line 47."

---

### B8. Guide

**Compact header:**
`M: Linguistic + Visual | T: Direct | F: Doc/Wiki | Audience: Operator who must follow a process correctly`

**Sub-genres:** Onboarding Guide, Process Guide, User Guide, Runbook

**Skeleton (W):**
```
1. WHAT THIS GUIDE COVERS One sentence: topic + target audience
2. PREREQUISITES          What the reader must have/know before starting
3. OVERVIEW               The big picture (what they will achieve)
4. STEP-BY-STEP PROCEDURE Numbered steps — one action per step
   - Step 1: [Action verb] [Object] — [why, if non-obvious]
   - Step N: ...
5. VERIFICATION           How to confirm each major step succeeded
6. COMMON ERRORS          What goes wrong + how to fix it
7. GLOSSARY               Terms the reader may not know
8. RELATED RESOURCES      Links to related guides, specs, contacts
```

**Quality checklist:**
- [ ] Every step starts with an action verb (Click, Navigate, Enter, Run)
- [ ] Screenshots or code blocks at every non-obvious step
- [ ] Prerequisites listed — reader is not surprised by missing access
- [ ] Common errors section is populated from real failures
- [ ] Guide has been tested by someone who did not write it
- [ ] Version/last-updated date is visible

**Example (L0):** "Onboarding guide for new engineers — 12 steps from access provisioning to first deploy."

---

### B9. Script

**Compact header:**
`M: Linguistic | T: Direct + Express | F: Doc | Audience: Performer (salesperson, presenter, host) who must deliver verbally`

**Sub-genres:** Sales Script, Video Script, Call Script, Demo Script

**Skeleton (W):**
```
1. SCRIPT HEADER          Purpose, audience, estimated duration, version
2. OPENING                Hook — first 10 seconds must earn continued attention
3. CONTEXT SETTING        Why this conversation / content matters right now
4. CORE CONTENT BLOCKS    [Label each block]
   - Block A: [Topic] — [talking points] — [transitions]
   - Block N: ...
5. OBJECTION HANDLERS     Common objections + exact response language
6. CALL TO ACTION         Single, specific next step (not multiple options)
7. CLOSING                How to end cleanly
8. NOTES / STAGE DIRECTIONS [Tone cues, pacing notes, visual cues if applicable]
```

**Quality checklist:**
- [ ] Opening does not start with "Hi my name is..."
- [ ] Each block has a clear transition to the next
- [ ] Objection handlers use the exact words the buyer uses (not paraphrases)
- [ ] Single CTA — not "you can either call, email, or book a meeting"
- [ ] Reading time matches declared duration
- [ ] Language is spoken-word natural, not written prose

**Example (L0):** "Discovery call script — 22 minutes, 4 blocks, 6 objection handlers, books a next meeting."

---

### B10. Template

**Compact header:**
`M: Linguistic | T: Inform or Direct | F: Doc/Text | Audience: Operator who must generate many similar signals quickly`

**Sub-genres:** Email Template, Message Template, Outreach Template, Contract Template

**Skeleton (W):**
```
1. TEMPLATE HEADER        Name, use case, version, last-updated
2. VARIABLE MANIFEST      All {{placeholders}} listed with description and example
3. SUBJECT / HEADLINE     For email/message — pre-written with variable slots
4. BODY                   Full content with {{variables}} embedded
5. CALL TO ACTION         Pre-written CTA block
6. USAGE NOTES            When to use, when NOT to use, customization guidance
7. PERFORMANCE DATA       (Optional) Open rate, reply rate, conversion if tracked
```

**Quality checklist:**
- [ ] Every variable is documented in the manifest — no mystery placeholders
- [ ] Template reads naturally when all variables are filled in
- [ ] Usage notes include a "do not use when" clause
- [ ] CTA is single and specific
- [ ] Subject line is ≤50 characters for email templates
- [ ] Template has been A/B tested or validated in production

**Example (L0):** "Cold outreach email template — 3 variables, 87-word body, 24% reply rate on record."

---

## PART II — COMMUNICATION GENRES

---

### C1. Status Update

**Compact header:**
`M: Linguistic | T: Inform | F: Chat/Doc | Audience: Manager or team tracking progress asynchronously`

**Sub-genres:** Weekly Signal, Standup, Progress Note

**Skeleton (W):**
```
1. PERIOD                 What timeframe this covers
2. DONE                   What was completed (bullet list, concrete nouns)
3. IN PROGRESS            What is actively being worked on + % or state
4. BLOCKED                What cannot proceed and why (plus who needs to act)
5. NEXT                   What starts next period
6. FLAG                   (Optional) One key thing the reader must know or decide
```

**Quality checklist:**
- [ ] Done items are completed deliverables, not activities
- [ ] Blocked items identify the specific blocker and the specific unblocking action
- [ ] Flag section exists only when there is something genuinely important
- [ ] Total length: under 150 words for standup, under 300 for weekly
- [ ] Past tense for done, present/future for in-progress and next

**Example (L0):** "Week 11 signal — 3 features shipped, 1 blocked on legal review, Q2 kickoff flagged for approval."

---

### C2. Meeting Notes

**Compact header:**
`M: Linguistic | T: Commit | F: Doc/Chat | Audience: All attendees and absent stakeholders who must act`

**Skeleton (W):**
```
1. MEETING HEADER         Date, attendees, facilitator, purpose, duration
2. DECISIONS MADE         [D] Decision — owner — rationale (one line per decision)
3. ACTION ITEMS           [A] Action — owner — due date (one line per item)
4. KEY DISCUSSION POINTS  Brief narrative of what was discussed (not transcript)
5. PARKING LOT            Topics raised but deferred — with follow-up owner
6. NEXT MEETING           Date, purpose, required prep
```

**Quality checklist:**
- [ ] Decisions and action items are in their own labeled sections (not buried in narrative)
- [ ] Every action item has exactly one owner and a specific due date
- [ ] Notes sent within 2 hours of meeting end
- [ ] Discussion section is a summary, not a transcript
- [ ] Parking lot items have owners — they are not dropped
- [ ] Next meeting has a declared purpose

**Example (L0):** "Strategy sync 03/16 — 2 decisions, 5 actions, 1 parking lot item, next meeting 03/23."

---

### C3. Chat Message

**Compact header:**
`M: Linguistic | T: Direct or Inform | F: Chat | Audience: Colleague who must parse fast in a high-velocity channel`

**Sub-genres:** Quick Question, Async Update, FYI, Request

**Skeleton (W) — Request:**
```
1. CONTEXT        One sentence: why you are messaging
2. REQUEST        What you specifically need (action verb + object)
3. DEADLINE       When you need it by
4. BACKGROUND     (Optional) Link or 1-2 sentences if context is non-obvious
```

**Skeleton (W) — FYI:**
```
1. SIGNAL WORD    "FYI:" or "Heads up:" at the start
2. WHAT           One sentence of what happened or changed
3. SO WHAT        Why it matters to the receiver
4. ACTION         "No action needed" OR specific ask
```

**Quality checklist:**
- [ ] Message starts with context, not pleasantries
- [ ] Exactly one request per message (not a list of asks)
- [ ] Deadline is explicit if one exists
- [ ] FYI messages explicitly state whether action is needed
- [ ] No walls of text — hard limit 5 lines before using a thread

**Example (L0):** "FYI: staging deploy failed — no action needed, infra team has it, resolved by EOD."

---

### C4. Email

**Compact header:**
`M: Linguistic | T: Direct or Inform or Commit | F: Email | Audience: External or internal receiver evaluating whether to act`

**Sub-genres:** Outreach Email, Follow-Up Email, Introduction Email, Escalation Email

**Skeleton (W) — Outreach / Business:**
```
1. SUBJECT LINE   ≤50 chars, specific, no clickbait
2. OPENER         One line of genuine context (not "hope you're well")
3. REASON         Why you are emailing — stated in sentence 2
4. VALUE / OFFER  What is in it for the receiver
5. EVIDENCE       One specific proof point (result, name, number)
6. SINGLE CTA     One clear next step — not multiple options
7. SIGNATURE      Name, role, contact — no 5-line legal disclaimer in internal email
```

**Quality checklist:**
- [ ] Subject line is specific enough to be searchable 6 months later
- [ ] Opener references something real about the receiver (not generic)
- [ ] Single CTA — not "feel free to call, email, or book a slot"
- [ ] Body is ≤150 words for cold outreach
- [ ] No passive voice in the CTA
- [ ] Attachments referenced explicitly in the body

**Example (L0):** "Outreach email to Acme CTO — 87 words, one ask: 20-minute call, sent Monday 8am."

---

### C5. Announcement

**Compact header:**
`M: Linguistic | T: Inform | F: Email/Chat/Doc | Audience: Team or public who must update their mental model`

**Sub-genres:** Team Announcement, Product Launch, Policy Change, Personnel Change

**Skeleton (W):**
```
1. HEADLINE           What is changing / happening (one sentence)
2. EFFECTIVE DATE      When this takes effect
3. WHAT IS CHANGING    Specific description of the change
4. WHY                 Reason — context earns buy-in
5. WHAT THIS MEANS FOR YOU   Impact on the specific receiver (most important section)
6. WHAT YOU NEED TO DO       Action required — or "no action needed" if none
7. QUESTIONS / CONTACT       Who to contact and how
```

**Quality checklist:**
- [ ] Headline states the change — not "exciting news" or "update"
- [ ] Effective date in sentence 1 or sentence 2 (never buried)
- [ ] "What this means for you" is personalized to the receiver group
- [ ] Action required is explicit (including "no action needed")
- [ ] Single contact point for questions — not a list
- [ ] Positive framing without hiding real impact on receiver

**Example (L0):** "Policy change: PTO accrual moves to unlimited effective April 1 — no action needed, details below."

---

## PART III — CONTENT GENRES

---

### K1. Article

**Compact header:**
`M: Linguistic | T: Inform or Express | F: Markdown/HTML | Audience: Reader choosing to invest attention for insight or skill`

**Sub-genres:** Blog Post, Thought Piece, Tutorial, Case Study

**Skeleton (W) — Informational / Thought Piece:**
```
1. HEADLINE           Promise: what insight or transformation will reader gain
2. LEDE               First 2–3 sentences — hook that earns the read
3. THESIS             Core argument or insight stated plainly
4. BODY SECTIONS      3–5 sections, each with its own sub-claim
   - Section header (the sub-claim as a statement, not a label)
   - Supporting evidence / argument
   - Transition to next section
5. SYNTHESIS          What the sections add up to — the insight restated with depth
6. IMPLICATION        So what? What should the reader do or think differently?
7. CLOSING            Memorable final line that earns recall
```

**Skeleton (W) — Tutorial:**
```
1. HEADLINE           What skill or outcome the reader will have
2. PREREQUISITES      What the reader must already know
3. OVERVIEW           The path we will take
4. STEP-BY-STEP       Numbered steps with code/screenshots
5. EXPLANATION        Why each step works (not just what to do)
6. COMMON MISTAKES    What goes wrong and why
7. WHAT'S NEXT        Where to go from here
```

**Quality checklist:**
- [ ] Headline makes a specific promise — not a topic label
- [ ] Lede does not start with a rhetorical question or generic setup
- [ ] Each section heading is a statement (claim), not a category label
- [ ] Thesis is stated — not implied
- [ ] Closing line earns recall — not "thanks for reading"
- [ ] Tutorial steps were tested by a human before publishing

**Example (L0):** "Article: 'Why status updates fail: 3 structural errors killing async communication' — 1,200 words."

---

### K2. Video Script

**Compact header:**
`M: Linguistic (spoken) | T: Inform or Express | F: Script Doc | Audience: Video editor + presenter + eventual viewer`

**Sub-genres:** YouTube Video, Course Video, Demo, Webinar Recording

**Skeleton (W):**
```
1. SCRIPT HEADER      Title, duration, target audience, tone, call to action
2. HOOK (0–30s)       One hook: surprising claim, question, or demonstration
3. PAYOFF PROMISE     "By the end of this video, you will [specific outcome]"
4. CONTENT BLOCKS     Each block has:
   - [VISUAL CUE]     What is on screen
   - [NARRATION]      Exactly what is said
   - [TRANSITION]     Bridge to next block
5. RECAP              Restate 3 key takeaways
6. CTA                Single specific next action for the viewer
7. OUTRO              Channel/brand close (pre-formatted, consistent)
8. B-ROLL NOTES       Visual requirements for editor
```

**Quality checklist:**
- [ ] Hook lands within 10 seconds — not 30
- [ ] Payoff promise is specific and immediately followable
- [ ] Narration is written for spoken delivery (contractions, short sentences)
- [ ] Visual cues match the narration at every point
- [ ] Single CTA — not "like, subscribe, comment, and visit our website"
- [ ] Script duration matches declared video length when read aloud

**Example (L0):** "YouTube script: '5 ADR mistakes killing your engineering team' — 8 minutes, subscribe CTA."

---

### K3. Social Post

**Compact header:**
`M: Linguistic | T: Express or Inform | F: Platform text | Audience: Scroll-stopping stranger who invests 0 attention by default`

**Sub-genres:** LinkedIn Post, Twitter/X Thread, Instagram Caption

**Skeleton (W) — LinkedIn / Long-form social:**
```
1. HOOK LINE         First line must stop the scroll — one sentence, no context
2. WHITE SPACE        Empty line after hook (forces "see more" click)
3. SETUP              2–3 lines of context or story
4. CORE INSIGHT       The one thing worth remembering — stated plainly
5. ELABORATION        3–5 bullets or short paragraphs expanding the insight
6. LANDING            Final line — memorable, emotionally resonant
7. CTA / ENGAGEMENT   Question, link, or call to share (optional, one only)
```

**Skeleton (W) — Twitter/X Thread:**
```
Tweet 1:  Hook + setup — must work as standalone
Tweet 2-N: One insight per tweet — each must work standalone
Last Tweet: Synthesis + CTA to follow or save thread
```

**Quality checklist:**
- [ ] First line works without context (tested by removing all other lines)
- [ ] Core insight is stated in plain language — no jargon
- [ ] Post does not start with "I" or "We" (LinkedIn algorithm penalty + weak signal)
- [ ] White space used generously (one idea per visual chunk)
- [ ] No more than one CTA
- [ ] Hashtags ≤3 for LinkedIn, ≤2 for X

**Example (L0):** "LinkedIn post: 'Most status updates are noise' — hook + 5-bullet framework + 1 question."

---

### K4. Course Module

**Compact header:**
`M: Mixed | T: Direct + Inform | F: Video + Doc | Audience: Learner who must be able to apply the skill after completion`

**Sub-genres:** Lesson, Exercise, Quiz, Workshop Module

**Skeleton (W) — Lesson:**
```
1. MODULE HEADER      Number, title, estimated time, prerequisites
2. LEARNING OBJECTIVE [Action verb] [skill] [to what standard] (Bloom's taxonomy)
3. CONCEPT EXPLANATION Core idea — plain language, no jargon first pass
4. WORKED EXAMPLE     Show it, don't just tell it
5. GUIDED PRACTICE    Learner does it with scaffolding
6. INDEPENDENT PRACTICE Learner does it without scaffolding
7. KNOWLEDGE CHECK    2–3 questions verifying the objective
8. SUMMARY            Restate the learning objective as achieved
9. BRIDGE             One sentence connecting this to the next module
```

**Quality checklist:**
- [ ] Learning objective uses measurable action verbs (build, diagnose, write — not "understand")
- [ ] Worked example is realistic, not toy-problem trivial
- [ ] Guided and independent practice are distinct (scaffolding explicitly removed)
- [ ] Knowledge check tests application, not memorization
- [ ] Module time estimate is accurate within ±20%
- [ ] Bridge sentence makes the module feel necessary, not standalone

**Example (L0):** "Module 3: Writing ADRs — 18 minutes, objective: author a complete ADR for a real decision."

---

### K5. Presentation

**Compact header:**
`M: Visual + Linguistic | T: Direct or Decide or Inform | F: Slides | Audience: Room or call audience who must leave with a changed mind or a decision`

**Sub-genres:** Pitch Deck, Workshop Slides, Webinar, Executive Briefing

**Skeleton (W) — Pitch / Decision presentation:**
```
1. TITLE SLIDE        Problem statement as title — not company name
2. THE PROBLEM        One slide, one number or story that makes the problem real
3. THE INSIGHT        Why existing solutions fail or miss the point
4. THE SOLUTION       What you propose — one slide, clear visual
5. EVIDENCE           Results, case study, proof (specifics only)
6. THE ASK            Exactly what you need from this room, today
7. THE MODEL          How it works (process, economics, mechanics)
8. TEAM / CREDIBILITY Why you/this team can execute this
9. APPENDIX           Deep-dive slides for Q&A
```

**Skeleton (W) — Workshop / Teaching:**
```
1. AGENDA             What we will cover + time per section
2. CONTEXT SETTER     Why this matters (5 minutes max)
3. CONTENT SECTIONS   One concept per slide — never more
4. EXERCISE SLIDES    Clear instructions, time box, output format
5. DEBRIEF SLIDES     What participants should have discovered
6. SUMMARY            3 key takeaways
7. NEXT STEPS         What to do after this workshop
```

**Quality checklist:**
- [ ] Each slide has one idea (one headline, one visual)
- [ ] Slide headline is a statement (claim), not a label
- [ ] No slide is a wall of bullets (max 5 words per bullet, max 4 bullets per slide)
- [ ] Ask/decision is explicit — not implied
- [ ] Appendix exists and is organized for Q&A navigation
- [ ] Deck works as a standalone document (can be read without the presenter)

**Example (L0):** "Pitch deck: Series A ask, 14 slides, $2M at $10M pre, Q2 close target."

---

## PART IV — GRANULARITY IN PRACTICE

### The Progressive Disclosure Protocol

Every signal exists at four levels simultaneously. Agents and operators choose the level
matched to the receiver's current need.

```
L0  HEADLINE    The signal in ~10 words.
                Purpose: navigation, filtering, scanning
                Format: one sentence, no structure

L1  SUMMARY     The signal in ~50 words. One paragraph.
                Purpose: decide whether to read further
                Format: context + finding + implication

L2  DETAIL      The signal in ~500 words. All required sections.
                Purpose: enough to act or implement
                Format: full skeleton with abbreviated sections

L3  COMPLETE    The signal at full fidelity. Unlimited.
                Purpose: approve, audit, implement precisely
                Format: full skeleton with complete sections + appendix
```

### Example: A Report at All Four Levels

**L0 (10 words):**
"March analytics: 23% revenue growth, CAC up 15%, two pipeline risks."

**L1 (50 words):**
"March closed at $340K — 23% above February and 8% above target. Customer acquisition cost
rose 15% due to paid channel saturation. Pipeline for Q2 shows $1.2M with two at-risk deals
flagged. Recommendation: reduce paid spend, accelerate referral program before Q2 mid-point."

**L2 (500 words):**
Full report with: Header, Executive Summary (L1 text), Key Metrics table, What Happened
narrative, Analysis section (CAC root cause), Risks table (2 items, P×I scored), Decisions
Needed (1 ask), Next Period Outlook.

**L3 (Complete):**
Full L2 plus: Appendix with raw data tables, cohort analysis, channel breakdown, methodology
notes, comparison to 6-month trend, full pipeline list with deal-level notes.

---

### Signal Routing Decision Tree

```
Does the receiver need to act right now?
├─ No, scanning → L0
├─ Yes, deciding whether to act → L1
├─ Yes, implementing → L2
└─ Yes, approving or auditing → L3

Is this a push (I am sending unsolicited)?
├─ Yes → Start at L0, offer L1 inline, link to L2/L3
└─ No, receiver requested it → Deliver at the level they requested

Is this asynchronous or synchronous?
├─ Async (email, doc, chat) → Start at L0/L1, layer downward
└─ Sync (meeting, call) → L1 verbal, L2 in follow-up doc
```

---

## PART V — NOISE TAXONOMY

Common failure modes by genre. Use as a pre-send checklist.

| Noise Type             | Symptom                                          | Fix                                      |
|------------------------|--------------------------------------------------|------------------------------------------|
| Genre mismatch         | Wrote prose when receiver needed a table          | Re-encode in correct genre skeleton      |
| Bandwidth overload     | 2,000-word email when 150 words suffice           | Reduce to L1, link to L3                 |
| Missing skeleton       | No sections, wall of text                         | Apply W dimension for the genre          |
| Buried verdict         | Conclusion appears at paragraph 7                 | Move verdict to line 1                   |
| Ambiguous ask          | "Let me know your thoughts"                       | Replace with specific action + deadline  |
| No success criteria    | Plan with no definition of done                   | Add measurable criteria to every goal    |
| Variables undeclared   | {{first_name}} in a sent email                    | Complete variable manifest before send   |
| Wrong granularity      | L3 delivered when L0 was needed                   | Match granularity to receiver state      |
| Ownership gap          | Task assigned to "team"                           | One owner per item — no exceptions       |
| Tense inconsistency    | Status update mixes past/present/future randomly  | Done=past, active=present, next=future   |

---

## QUICK-REFERENCE INDEX

| Genre              | Type    | Format       | Key Section to Get Right     |
|--------------------|---------|--------------|-------------------------------|
| B1 Brief           | Direct  | Doc          | Scope (in AND out)            |
| B2 Proposal        | Commit  | Doc/Slides   | Executive Summary             |
| B3 Report          | Inform  | Doc/Dashboard| Decisions Needed              |
| B4 Spec            | Direct  | Markdown     | Non-goals + Edge Cases        |
| B5 Decision Record | Decide  | Markdown     | Alternatives Considered       |
| B6 Plan            | Commit  | Doc/Table    | Owner per task (never "team") |
| B7 Review          | Decide  | Doc          | Verdict at top                |
| B8 Guide           | Direct  | Doc/Wiki     | Verification steps            |
| B9 Script          | Direct  | Doc          | Objection Handlers            |
| B10 Template       | Inform  | Doc/Text     | Variable Manifest             |
| C1 Status Update   | Inform  | Chat/Doc     | Blocked (specific unblocking) |
| C2 Meeting Notes   | Commit  | Doc/Chat     | Actions with owners + dates   |
| C3 Chat Message    | Direct  | Chat         | Single request per message    |
| C4 Email           | Direct  | Email        | Single CTA                    |
| C5 Announcement    | Inform  | Email/Chat   | What This Means For You       |
| K1 Article         | Inform  | Markdown     | Thesis stated explicitly      |
| K2 Video Script    | Inform  | Script Doc   | Hook in first 10 seconds      |
| K3 Social Post     | Express | Platform     | First line works standalone   |
| K4 Course Module   | Direct  | Video+Doc    | Measurable learning objective |
| K5 Presentation    | Decide  | Slides       | The Ask (one slide)           |
