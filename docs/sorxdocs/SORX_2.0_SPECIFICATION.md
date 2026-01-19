# Sorx 2.0 - System of Reasoning

## Universal Skill-Based Integration Framework

---

## What is Sorx 2.0?

**Sorx 2.0** (System of Reasoning) is a next-generation integration framework where AI agents **learn skills** to connect with any system - modern APIs, legacy systems, databases, desktop applications, hardware, or anything with an interface.

Unlike traditional integration platforms that provide pre-built tools, Sorx 2.0 agents **acquire skills through experience** and **improve over time** - just like a human learning their job.

### The USB-C Paradigm

Just as USB-C is a universal connector that works with any device, Sorx 2.0 is a universal interface that works with any system:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SORX 2.0                                       │
│                         Universal Connector                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Modern Systems          Legacy Systems          Physical World            │
│   ──────────────          ──────────────          ──────────────            │
│   REST APIs               SOAP/XML-RPC            IoT Devices               │
│   GraphQL                 FTP/SFTP                Hardware APIs             │
│   WebSocket               EDI                     Serial/USB                │
│   gRPC                    Mainframe               Bluetooth                 │
│                                                                             │
│   Databases               Desktop Apps            File Systems              │
│   ──────────────          ──────────────          ──────────────            │
│   PostgreSQL              AppleScript             Local Files               │
│   MySQL                   PowerShell              Cloud Storage             │
│   MongoDB                 Windows COM             Network Shares            │
│   Redis                   X11 Automation          FTP/SFTP                  │
│                                                                             │
│   Enterprise              Communication           Custom                    │
│   ──────────────          ──────────────          ──────────────            │
│   SAP                     Email (SMTP/IMAP)       Any Protocol              │
│   Oracle                  SMS Gateways            Any Interface             │
│   Salesforce              Voice APIs              Any System                │
│   AS/400                  Fax Services            Anything                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Foundational Architecture: Open/Closed Model

### The Two-Layer System

Sorx 2.0 operates as a bridge between two distinct architectural layers:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    OPEN/CLOSED ARCHITECTURE MODEL                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │                      CLOSED ARCHITECTURE                              │  │
│   │                        (BusinessOS Core)                              │  │
│   │                                                                       │  │
│   │   ┌───────────────────────────────────────────────────────────────┐ │  │
│   │   │                    OBJECTIVE DATABASE                          │ │  │
│   │   │                                                                │ │  │
│   │   │   • Data points that define task completion                   │ │  │
│   │   │   • Controlled by BusinessOS                                  │ │  │
│   │   │   • Source of truth for workflows                             │ │  │
│   │   │   • Never exposed directly to external systems                │ │  │
│   │   │                                                                │ │  │
│   │   └───────────────────────────────────────────────────────────────┘ │  │
│   │                                                                       │  │
│   │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐              │  │
│   │   │ DEPARTMENTS │   │    ROLES    │   │    NODES    │              │  │
│   │   │             │   │             │   │  (Workers)  │              │  │
│   │   │  Sales      │   │  @sales     │   │  Node-001   │              │  │
│   │   │  Support    │   │  @support   │   │  Node-002   │              │  │
│   │   │  Finance    │   │  @finance   │   │  Node-003   │              │  │
│   │   │  Marketing  │   │  @marketing │   │  ...        │              │  │
│   │   │  Operations │   │  @ops       │   │             │              │  │
│   │   └─────────────┘   └─────────────┘   └─────────────┘              │  │
│   │                                                                       │  │
│   │   WE CONTROL:                                                        │  │
│   │   ✓ What "done" means (objective data points)                       │  │
│   │   ✓ Who can do what (role permissions)                              │  │
│   │   ✓ How work flows (workflow logic)                                 │  │
│   │   ✓ What success looks like (completion criteria)                   │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                     │                                       │
│                                     │ Skills bridge the gap                 │
│                                     ▼                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │                       OPEN ARCHITECTURE                               │  │
│   │                     (Sorx 2.0 Skill Layer)                            │  │
│   │                                                                       │  │
│   │   External world: APIs, services, systems, integrations              │  │
│   │                                                                       │  │
│   │   ┌─────────────────────────────────────────────────────────────┐   │  │
│   │   │                    ROLE-BASED SKILLS                         │   │  │
│   │   │               (Not RPA - These are like PEOPLE)              │   │  │
│   │   │                                                              │   │  │
│   │   │   @sales-role                    @support-role               │   │  │
│   │   │   ├─ hubspot_qualify_lead        ├─ zendesk_triage_ticket   │   │  │
│   │   │   ├─ gmail_outreach_sequence     ├─ slack_respond_customer  │   │  │
│   │   │   ├─ calendly_book_meeting       ├─ notion_create_kb_article│   │  │
│   │   │   └─ hubspot_close_deal          └─ jira_escalate_issue     │   │  │
│   │   │                                                              │   │  │
│   │   │   @finance-role                  @ops-role                   │   │  │
│   │   │   ├─ quickbooks_create_invoice   ├─ clickup_create_project  │   │  │
│   │   │   ├─ stripe_process_payment      ├─ asana_assign_tasks      │   │  │
│   │   │   ├─ xero_reconcile_accounts     ├─ monday_update_status    │   │  │
│   │   │   └─ excel_generate_report       └─ zapier_trigger_workflow │   │  │
│   │   │                                                              │   │  │
│   │   └─────────────────────────────────────────────────────────────┘   │  │
│   │                                                                       │  │
│   │   SKILLS CONNECT TO:                                                 │  │
│   │   → HubSpot, Salesforce, Pipedrive (CRM)                            │  │
│   │   → Gmail, Outlook, Slack, Teams (Communication)                    │  │
│   │   → QuickBooks, Xero, Stripe (Finance)                              │  │
│   │   → ClickUp, Asana, Monday, Notion (Productivity)                   │  │
│   │   → Any external system with an interface                           │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Objective Database: The Source of Truth

The **Objective Database** is the closed-architecture core that defines:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         OBJECTIVE DATABASE                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   DATA POINTS (Define what "done" means)                                    │
│   ═════════════════════════════════════                                     │
│                                                                             │
│   Task: "Close the deal with Acme Corp"                                     │
│                                                                             │
│   Objective Data Points:                                                    │
│   □ deal.status = "closed_won"           ← Must be true                    │
│   □ deal.contract_signed = true          ← Must be true                    │
│   □ deal.payment_terms_agreed = true     ← Must be true                    │
│   □ client.onboarding_scheduled = true   ← Must be true                    │
│   □ invoice.created = true               ← Must be true                    │
│                                                                             │
│   The task is NOT complete until ALL data points are satisfied.            │
│   Skills execute to achieve each data point.                               │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   WORKFLOW LOGIC (How we get there)                                         │
│   ═════════════════════════════════                                         │
│                                                                             │
│   The system knows:                                                         │
│   1. Current state of all data points                                      │
│   2. Which skills can change which data points                             │
│   3. Dependencies (can't invoice before contract signed)                   │
│   4. Optimal path to completion (minimize movement)                        │
│                                                                             │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐            │
│   │ Current │────▶│ Gap     │────▶│ Skills  │────▶│ Execute │            │
│   │ State   │     │ Analysis│     │ Needed  │     │ Optimal │            │
│   └─────────┘     └─────────┘     └─────────┘     └─────────┘            │
│                                                                             │
│   MINIMIZE MOVEMENT = Find the shortest path through data points           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Roles: Not RPA, These are Like PEOPLE

Skills don't exist in isolation - they belong to **Roles**, just like skills belong to people in a real company:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ROLE-BASED ORGANIZATION                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   THIS IS NOT RPA                                                           │
│   ═══════════════                                                           │
│                                                                             │
│   RPA = "Click button, fill form, repeat"  ❌ Mechanical                   │
│   Sorx = "Role with skills, learns, adapts" ✓ Human-like                   │
│                                                                             │
│   Think of it like hiring for a company:                                    │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                         DEPARTMENT: SALES                            │  │
│   │                                                                       │  │
│   │   @sales-role (The "person" doing sales work)                        │  │
│   │   │                                                                   │  │
│   │   ├─ SKILLS THIS ROLE HAS:                                           │  │
│   │   │   ├─ hubspot_qualify_lead      (learned)                        │  │
│   │   │   ├─ hubspot_update_deal       (learned)                        │  │
│   │   │   ├─ gmail_send_proposal       (learned)                        │  │
│   │   │   ├─ calendly_schedule_call    (learned)                        │  │
│   │   │   ├─ slack_notify_team         (learned)                        │  │
│   │   │   └─ [can learn more as needed]                                 │  │
│   │   │                                                                   │  │
│   │   ├─ PERMISSIONS:                                                    │  │
│   │   │   ├─ ✓ Can access CRM data                                      │  │
│   │   │   ├─ ✓ Can send emails on behalf of sales                       │  │
│   │   │   ├─ ✓ Can schedule meetings                                    │  │
│   │   │   ├─ ✗ Cannot access financial data                             │  │
│   │   │   └─ ✗ Cannot modify contracts                                  │  │
│   │   │                                                                   │  │
│   │   └─ NODES ASSIGNED: Node-001, Node-002, Node-003                   │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                        DEPARTMENT: FINANCE                           │  │
│   │                                                                       │  │
│   │   @finance-role                                                      │  │
│   │   │                                                                   │  │
│   │   ├─ SKILLS:                                                         │  │
│   │   │   ├─ quickbooks_create_invoice  (learned)                       │  │
│   │   │   ├─ stripe_process_refund      (learned)                       │  │
│   │   │   ├─ xero_reconcile             (learned)                       │  │
│   │   │   └─ excel_financial_report     (learned)                       │  │
│   │   │                                                                   │  │
│   │   ├─ PERMISSIONS:                                                    │  │
│   │   │   ├─ ✓ Can access financial data                                │  │
│   │   │   ├─ ✓ Can create/modify invoices                               │  │
│   │   │   ├─ ✓ Can process payments                                     │  │
│   │   │   └─ ✗ Cannot access sales pipeline                             │  │
│   │   │                                                                   │  │
│   │   └─ NODES ASSIGNED: Node-010, Node-011                             │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Skill Reliability Tiers

### Two Types of Skills: Deterministic vs AI-Driven

Not all skills are created equal. Some are **100% reliable** (hardcoded), others are **AI-driven** (can fail):

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SKILL RELIABILITY TIERS                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TIER 1: DETERMINISTIC (100% Uptime)                                       │
│   ════════════════════════════════════                                      │
│                                                                             │
│   These skills are HARDCODED. Like RPA, but smarter.                        │
│   Zero AI involvement at execution time.                                    │
│   If API is up, skill WILL succeed.                                         │
│                                                                             │
│   Examples:                                                                 │
│   ├─ hubspot_create_contact    → API call, fixed parameters               │
│   ├─ slack_send_message        → API call, fixed format                   │
│   ├─ gmail_send_email          → SMTP, well-defined                       │
│   ├─ quickbooks_create_invoice → API call, structured data                │
│   └─ clickup_create_task       → API call, known schema                   │
│                                                                             │
│   Characteristics:                                                          │
│   ✓ 100% success rate (when external API is up)                            │
│   ✓ No model needed for execution                                          │
│   ✓ Instant execution                                                      │
│   ✓ Predictable, testable, auditable                                       │
│                                                                             │
│   Model: NONE (pure code execution)                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TIER 2: STRUCTURED AI (95-99% Uptime)                                     │
│   ══════════════════════════════════════                                    │
│                                                                             │
│   AI involvement for parameter extraction or simple decisions.              │
│   Core execution is still deterministic.                                    │
│                                                                             │
│   Examples:                                                                 │
│   ├─ email_smart_reply        → AI drafts, but SMTP is deterministic      │
│   ├─ ticket_auto_categorize   → AI classifies, but API call is fixed      │
│   ├─ lead_score_update        → AI scores, but HubSpot update is fixed    │
│   └─ meeting_smart_schedule   → AI picks time, but Calendly is fixed      │
│                                                                             │
│   Characteristics:                                                          │
│   ✓ High success rate                                                      │
│   ✓ Small model sufficient (Haiku)                                         │
│   ✓ Fast execution                                                         │
│   ✓ Fallback to human if AI unsure                                         │
│                                                                             │
│   Model: HAIKU (fast, cheap, good enough)                                   │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TIER 3: REASONING AI (80-95% Uptime)                                      │
│   ════════════════════════════════════                                      │
│                                                                             │
│   Complex reasoning required. Multiple steps. Decision trees.               │
│   May need human-in-the-loop for edge cases.                               │
│                                                                             │
│   Examples:                                                                 │
│   ├─ proposal_generation      → AI writes, formats, customizes            │
│   ├─ contract_negotiation     → AI suggests terms, handles objections     │
│   ├─ support_complex_issue    → AI diagnoses, proposes solutions          │
│   └─ financial_analysis       → AI interprets data, makes recommendations │
│                                                                             │
│   Characteristics:                                                          │
│   ~ Variable success rate                                                   │
│   ~ Larger model needed (Sonnet/Opus)                                       │
│   ~ Slower execution                                                        │
│   ~ May require approval                                                    │
│                                                                             │
│   Model: SONNET (balanced) or OPUS (complex reasoning)                     │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TIER 4: GENERATIVE AI (Variable Uptime)                                   │
│   ════════════════════════════════════════                                  │
│                                                                             │
│   New skill generation. Novel situations. First-time execution.             │
│   Higher risk, higher reward.                                               │
│                                                                             │
│   Examples:                                                                 │
│   ├─ skill_creator            → Generates new skills from description     │
│   ├─ workflow_optimizer       → Redesigns processes                       │
│   ├─ integration_builder      → Creates new API connections               │
│   └─ automation_designer      → Builds complex multi-step workflows       │
│                                                                             │
│   Characteristics:                                                          │
│   ~ Unpredictable success rate until validated                             │
│   ~ Largest model needed (Opus)                                            │
│   ~ Requires testing before production use                                 │
│   ~ Human review recommended                                               │
│                                                                             │
│   Model: OPUS (maximum capability)                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Model Selection by Skill Complexity

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     MODEL SELECTION MATRIX                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   SKILL COMPLEXITY           MODEL          COST      LATENCY    SUCCESS   │
│   ════════════════           ═════          ════      ═══════    ═══════   │
│                                                                             │
│   Deterministic (Tier 1)     NONE           $0        <100ms     100%     │
│   Structured AI (Tier 2)     HAIKU          $         <500ms     95-99%   │
│   Reasoning AI (Tier 3)      SONNET         $$        <2s        80-95%   │
│   Complex Reasoning          OPUS           $$$       <5s        70-90%   │
│   Generative (Tier 4)        OPUS           $$$$      <10s       Variable │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   AUTOMATIC MODEL SELECTION                                                 │
│   ═════════════════════════                                                 │
│                                                                             │
│   When a skill is created or executed:                                      │
│                                                                             │
│   1. Analyze skill complexity                                               │
│      └─ Parameter count, decision points, output type                      │
│                                                                             │
│   2. Check historical performance                                           │
│      └─ What model succeeded before? What failed?                          │
│                                                                             │
│   3. Select minimum viable model                                            │
│      └─ Don't use Opus if Haiku will work                                  │
│                                                                             │
│   4. Upgrade if failures occur                                              │
│      └─ If Haiku fails 3x, try Sonnet                                      │
│      └─ If Sonnet fails 3x, try Opus                                       │
│                                                                             │
│   5. Downgrade if overqualified                                             │
│      └─ If Opus has 100% success on simple task, try Sonnet               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Skill Auto-Splitting: Self-Healing Skills

### When Skills Fail, The System Learns

When a skill starts accumulating errors, the system doesn't just retry - it **learns** and **splits**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SKILL AUTO-SPLITTING                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   THE PROBLEM                                                               │
│   ═══════════                                                               │
│                                                                             │
│   Skill: email_send_followup (v1)                                          │
│   Success Rate: Dropping from 95% → 75% → 60%                              │
│                                                                             │
│   Why? The skill handles too many different situations:                     │
│   • Cold leads (different tone)                                            │
│   • Warm leads (different content)                                         │
│   • Post-meeting (different context)                                       │
│   • Post-proposal (different urgency)                                      │
│                                                                             │
│   One skill trying to do too much = errors                                 │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   THE SOLUTION: AUTO-SPLIT                                                  │
│   ════════════════════════════                                              │
│                                                                             │
│   When error rate exceeds threshold (e.g., 20%):                           │
│                                                                             │
│   1. ANALYZE FAILURES                                                       │
│      └─ What patterns do failed executions have in common?                 │
│      └─ What was different about successful ones?                          │
│                                                                             │
│   2. IDENTIFY SPLIT POINTS                                                  │
│      └─ "Failures mostly occur when lead_status = 'cold'"                 │
│      └─ "Failures mostly occur when days_since_meeting > 7"               │
│                                                                             │
│   3. CREATE SKILL TREE                                                      │
│      └─ Split one skill into multiple specialized skills                   │
│      └─ Add decision node at the root                                      │
│                                                                             │
│   BEFORE:                                                                   │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │                    email_send_followup (v1)                          │  │
│   │                         60% success                                   │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   AFTER (Auto-Split):                                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │                    email_followup_router (decision node)             │  │
│   │                              │                                        │  │
│   │          ┌──────────────────┼──────────────────┐                    │  │
│   │          │                  │                  │                    │  │
│   │          ▼                  ▼                  ▼                    │  │
│   │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐              │  │
│   │   │ cold_lead   │   │ warm_lead   │   │ post_meeting│              │  │
│   │   │ _followup   │   │ _followup   │   │ _followup   │              │  │
│   │   │   95%       │   │   98%       │   │   97%       │              │  │
│   │   └─────────────┘   └─────────────┘   └─────────────┘              │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   OVERALL SUCCESS: 97% (up from 60%)                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Tree Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SKILL TREE STRUCTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   class SkillTree:                                                          │
│       id: str                     # "email_followup_tree"                  │
│       root: SkillNode             # Decision node                          │
│       created_from: str           # "email_send_followup_v1" (parent)      │
│       split_reason: str           # "High error rate in cold lead cases"  │
│       overall_success_rate: float # Weighted average of leaves            │
│                                                                             │
│   class SkillNode:                                                          │
│       type: "decision" | "skill"                                           │
│       condition: str              # "lead.status == 'cold'"                │
│       skill_id: str               # For leaf nodes                         │
│       children: list[SkillNode]   # For decision nodes                     │
│                                                                             │
│   EXAMPLE TREE:                                                             │
│                                                                             │
│   {                                                                         │
│     "id": "email_followup_tree",                                           │
│     "root": {                                                              │
│       "type": "decision",                                                  │
│       "conditions": [                                                      │
│         {                                                                  │
│           "if": "lead.status == 'cold'",                                  │
│           "then": {"type": "skill", "skill_id": "cold_lead_followup_v1"} │
│         },                                                                 │
│         {                                                                  │
│           "if": "lead.status == 'warm'",                                  │
│           "then": {"type": "skill", "skill_id": "warm_lead_followup_v1"} │
│         },                                                                 │
│         {                                                                  │
│           "if": "lead.last_meeting != null",                              │
│           "then": {"type": "skill", "skill_id": "post_meeting_followup"}  │
│         },                                                                 │
│         {                                                                  │
│           "default": true,                                                │
│           "then": {"type": "skill", "skill_id": "generic_followup_v1"}   │
│         }                                                                  │
│       ]                                                                    │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Error Threshold and Split Logic

```python
# skill_health_monitor.py

class SkillHealthMonitor:
    """Monitors skill performance and triggers auto-splitting."""

    def __init__(self):
        self.error_threshold = 0.20        # 20% error rate triggers analysis
        self.min_executions = 50           # Need enough data before splitting
        self.analysis_window = timedelta(days=7)

    def check_skill_health(self, skill_id: str) -> HealthStatus:
        """Check if a skill needs intervention."""

        metrics = get_skill_metrics(
            skill_id,
            window=self.analysis_window
        )

        if metrics.execution_count < self.min_executions:
            return HealthStatus.INSUFFICIENT_DATA

        error_rate = metrics.failures / metrics.execution_count

        if error_rate < 0.05:  # < 5% errors
            return HealthStatus.HEALTHY

        if error_rate < self.error_threshold:  # 5-20%
            return HealthStatus.DEGRADED

        # > 20% errors - needs splitting
        return HealthStatus.NEEDS_SPLIT

    def analyze_failures(self, skill_id: str) -> SplitAnalysis:
        """Analyze failure patterns to determine split points."""

        failures = get_failed_executions(skill_id)
        successes = get_successful_executions(skill_id)

        # Find distinguishing features
        failure_patterns = extract_common_patterns(failures)
        success_patterns = extract_common_patterns(successes)

        # Identify split conditions
        split_points = []
        for pattern in failure_patterns:
            if pattern not in success_patterns:
                split_points.append({
                    "condition": pattern.as_condition(),
                    "failure_rate_when_true": pattern.failure_rate,
                    "sample_size": pattern.count
                })

        return SplitAnalysis(
            skill_id=skill_id,
            recommended_splits=split_points,
            confidence=calculate_confidence(split_points)
        )

    def execute_split(self, skill_id: str, analysis: SplitAnalysis) -> SkillTree:
        """Create a skill tree from split analysis."""

        # Create specialized skills for each split point
        new_skills = []
        for split in analysis.recommended_splits:
            new_skill = generate_specialized_skill(
                parent_skill_id=skill_id,
                specialization=split.condition
            )
            new_skills.append(new_skill)

        # Create decision tree
        tree = SkillTree(
            id=f"{skill_id}_tree",
            created_from=skill_id,
            root=build_decision_tree(new_skills)
        )

        # Retire original skill
        deprecate_skill(skill_id, replacement=tree.id)

        return tree
```

---

## Workflow Completion: Objective Data Points

### Task is Done When ALL Data Points Are Satisfied

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      WORKFLOW COMPLETION LOGIC                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   A task is NOT complete until ALL objective data points are TRUE.          │
│                                                                             │
│   EXAMPLE: "Onboard new client Acme Corp"                                   │
│   ════════════════════════════════════════                                  │
│                                                                             │
│   Objective Data Points:                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ □ contract.signed = true                           [PENDING]        │  │
│   │ □ payment.first_invoice_paid = true                [PENDING]        │  │
│   │ □ project.created = true                           [PENDING]        │  │
│   │ □ project.kickoff_scheduled = true                 [PENDING]        │  │
│   │ □ team.assigned = true                             [PENDING]        │  │
│   │ □ client.welcome_email_sent = true                 [PENDING]        │  │
│   │ □ client.access_provisioned = true                 [PENDING]        │  │
│   │ □ documentation.client_wiki_created = true         [PENDING]        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   TASK STATUS: 0/8 complete → NOT DONE                                     │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   EXECUTION FLOW                                                            │
│   ══════════════                                                            │
│                                                                             │
│   1. System analyzes current state (all data points = false)               │
│                                                                             │
│   2. System identifies skills needed:                                       │
│      ├─ docusign_send_contract     → contract.signed                       │
│      ├─ quickbooks_create_invoice  → payment.first_invoice_paid            │
│      ├─ clickup_create_project     → project.created                       │
│      ├─ calendly_schedule_meeting  → project.kickoff_scheduled             │
│      ├─ clickup_assign_team        → team.assigned                         │
│      ├─ gmail_send_welcome         → client.welcome_email_sent             │
│      ├─ okta_provision_access      → client.access_provisioned             │
│      └─ notion_create_wiki         → documentation.client_wiki_created     │
│                                                                             │
│   3. System finds optimal execution order:                                  │
│      (Respecting dependencies, parallelizing where possible)               │
│                                                                             │
│      PARALLEL:                                                              │
│      ├─ docusign_send_contract                                             │
│      ├─ clickup_create_project                                             │
│      └─ notion_create_wiki                                                  │
│                                                                             │
│      AFTER contract.signed:                                                 │
│      ├─ quickbooks_create_invoice                                          │
│      └─ gmail_send_welcome                                                  │
│                                                                             │
│      AFTER project.created:                                                 │
│      ├─ clickup_assign_team                                                │
│      └─ calendly_schedule_meeting                                          │
│                                                                             │
│      AFTER payment.first_invoice_paid:                                     │
│      └─ okta_provision_access                                              │
│                                                                             │
│   4. Execute skills, update data points as each completes                  │
│                                                                             │
│   5. Continue until ALL data points = true                                 │
│                                                                             │
│   TASK STATUS: 8/8 complete → DONE                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Minimize Movement: Optimal Path

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          MINIMIZE MOVEMENT                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   PRINCIPLE: Do the least amount of work to achieve all data points.       │
│                                                                             │
│   BAD PATH (unnecessary steps):                                             │
│   ────────────────────────────                                              │
│   1. Create project                                                        │
│   2. Send welcome email                                                    │
│   3. Wait for response                     ← UNNECESSARY                   │
│   4. Send follow-up                        ← UNNECESSARY                   │
│   5. Create invoice                                                        │
│   6. Send invoice reminder                 ← UNNECESSARY                   │
│   7. Create wiki page                                                      │
│   8. Notify team                           ← COULD BE PARALLEL             │
│   9. Assign team                                                           │
│   10. Schedule kickoff                                                     │
│                                                                             │
│   10 steps, sequential, slow                                               │
│                                                                             │
│   OPTIMAL PATH (minimum movement):                                          │
│   ─────────────────────────────────                                         │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  PARALLEL EXECUTION                                                   │  │
│   │                                                                       │  │
│   │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │  │
│   │  │ Contract  │  │  Create   │  │  Create   │  │  Create   │        │  │
│   │  │  (DocuSign)│ │  Project  │  │   Wiki    │  │  Invoice  │        │  │
│   │  └─────┬─────┘  └─────┬─────┘  └───────────┘  └─────┬─────┘        │  │
│   │        │              │                              │              │  │
│   │        │              ├────────────┐                │              │  │
│   │        │              ▼            ▼                ▼              │  │
│   │        │        ┌───────────┐ ┌───────────┐  ┌───────────┐        │  │
│   │        │        │  Assign   │ │ Schedule  │  │  Welcome  │        │  │
│   │        │        │   Team    │ │  Kickoff  │  │   Email   │        │  │
│   │        │        └───────────┘ └───────────┘  └─────┬─────┘        │  │
│   │        │                                           │              │  │
│   │        └───────────────────────────────────────────┘              │  │
│   │                              │                                     │  │
│   │                              ▼                                     │  │
│   │                        ┌───────────┐                              │  │
│   │                        │ Provision │                              │  │
│   │                        │  Access   │                              │  │
│   │                        └───────────┘                              │  │
│   │                                                                   │  │
│   └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│   7 skills, parallel where possible, fast                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Human-in-the-Loop Architecture

### The Core Principle: Humans Pull, System Doesn't Push

Sorx 2.0 is designed as a **human-centric system** where humans remain in control. The system prepares, suggests, and organizes - but humans dictate when and how things happen.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        PULL vs PUSH MODEL                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ❌ PUSH MODEL (What we DON'T do)                                          │
│   ═══════════════════════════════                                           │
│                                                                             │
│   System decides → Executes → Notifies human                               │
│                                                                             │
│   Problems:                                                                 │
│   • Human feels out of control                                             │
│   • Actions happen without context                                         │
│   • Overwhelmed with notifications                                         │
│   • Can't prioritize what matters                                          │
│   • System runs ahead of human understanding                               │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ✅ PULL MODEL (What we DO)                                                │
│   ═══════════════════════════                                               │
│                                                                             │
│   Human asks → System presents → Human decides → System executes           │
│                                                                             │
│   Benefits:                                                                 │
│   • Human always in control                                                │
│   • Context provided before action                                         │
│   • Human pulls when ready                                                 │
│   • Prioritization by human                                                │
│   • System waits for human direction                                       │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   PULL INTERACTIONS                                                         │
│   ═════════════════                                                         │
│                                                                             │
│   Human: "What's waiting for me?"                                          │
│   System: "You have 3 tasks ready for review:                              │
│            1. Proposal for Acme Corp (draft ready)                         │
│            2. Invoice approval needed ($15K)                               │
│            3. New lead qualification (high score)"                         │
│                                                                             │
│   Human: "Show me the proposal"                                            │
│   System: [Displays proposal with context]                                 │
│                                                                             │
│   Human: "Looks good, send it"                                             │
│   System: [Executes send, updates status]                                  │
│                                                                             │
│   Human: "What's next?"                                                    │
│   System: [Presents next item when human is ready]                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Temperature Control: Human Dictates Autonomy

The **Temperature** is the level of autonomy humans grant to the system. Humans control the dial.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TEMPERATURE CONTROL                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TEMPERATURE DIAL (Controlled by Human)                                    │
│   ══════════════════════════════════════                                    │
│                                                                             │
│   ❄️──────────────────🌡️──────────────────🔥                                │
│   COLD              WARM               HOT                                  │
│   Full Control      Balanced           High Autonomy                        │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ❄️ COLD TEMPERATURE (Full Human Control)                                  │
│   ═════════════════════════════════════════                                 │
│                                                                             │
│   Every action requires explicit approval:                                  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  TASK: Send follow-up email to John                                  │  │
│   │                                                                       │  │
│   │  CONTEXT:                                                            │  │
│   │  • Last contact: 3 days ago                                          │  │
│   │  • Deal stage: Proposal sent                                         │  │
│   │  • Deal value: $25,000                                               │  │
│   │                                                                       │  │
│   │  PROPOSED ACTION:                                                    │  │
│   │  Send email with subject: "Following up on our proposal"            │  │
│   │                                                                       │  │
│   │  DRAFT:                                                              │  │
│   │  "Hi John, I wanted to check in on the proposal I sent..."          │  │
│   │                                                                       │  │
│   │  [APPROVE]  [EDIT]  [REJECT]  [SKIP FOR NOW]                        │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   Use when:                                                                 │
│   • New to the system                                                      │
│   • High-stakes actions                                                    │
│   • Training the system                                                    │
│   • Want to review everything                                              │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   🌡️ WARM TEMPERATURE (Balanced)                                            │
│   ═══════════════════════════════                                           │
│                                                                             │
│   Routine actions auto-execute, complex decisions wait:                     │
│                                                                             │
│   AUTO-EXECUTE (No approval needed):                                        │
│   ✓ Update CRM fields                                                      │
│   ✓ Log activities                                                         │
│   ✓ Send routine notifications                                             │
│   ✓ Create internal tasks                                                  │
│                                                                             │
│   WAIT FOR APPROVAL:                                                        │
│   ⏸ External communications (emails, messages)                             │
│   ⏸ Financial actions (invoices, payments)                                 │
│   ⏸ Commitments (meetings, deadlines)                                      │
│   ⏸ New client interactions                                                │
│                                                                             │
│   OUTPUT DISPLAY:                                                           │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  SUMMARY: 12 actions completed, 3 waiting for you                    │  │
│   │                                                                       │  │
│   │  ✅ COMPLETED:                                                        │  │
│   │  • Updated 5 deal stages                                             │  │
│   │  • Logged 4 call activities                                          │  │
│   │  • Created 3 follow-up tasks                                         │  │
│   │                                                                       │  │
│   │  ⏳ WAITING FOR YOU:                                                  │  │
│   │  1. [Review] Email to Acme Corp CEO                                  │  │
│   │  2. [Approve] Invoice for $15,000                                    │  │
│   │  3. [Confirm] Meeting with new prospect                              │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   🔥 HOT TEMPERATURE (High Autonomy)                                        │
│   ════════════════════════════════════                                      │
│                                                                             │
│   Most actions auto-execute, only critical decisions escalate:              │
│                                                                             │
│   AUTO-EXECUTE:                                                             │
│   ✓ All routine actions                                                    │
│   ✓ Standard emails (templates, follow-ups)                                │
│   ✓ Standard invoices (within parameters)                                  │
│   ✓ Meeting scheduling (within calendar rules)                             │
│                                                                             │
│   ESCALATE ONLY:                                                            │
│   🚨 Actions above financial threshold                                     │
│   🚨 First contact with new clients                                        │
│   🚨 Exceptions to normal process                                          │
│   🚨 Conflicts or errors                                                   │
│                                                                             │
│   Human gets summaries:                                                     │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  DAILY SUMMARY                                                        │  │
│   │                                                                       │  │
│   │  47 actions completed today                                          │  │
│   │  • 15 emails sent                                                    │  │
│   │  • 8 deals updated                                                   │  │
│   │  • 12 tasks completed                                                │  │
│   │  • 5 invoices created                                                │  │
│   │  • 7 meetings scheduled                                              │  │
│   │                                                                       │  │
│   │  🚨 1 item needs your attention:                                      │  │
│   │  → $50K deal requires manual approval (above threshold)              │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Context Display: What Humans See

Before any action, humans see full context:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CONTEXT DISPLAY FORMAT                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Every task presented to human includes:                                   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  📋 TASK: [What needs to happen]                                     │  │
│   │  ════════════════════════════════                                    │  │
│   │                                                                       │  │
│   │  🎯 OBJECTIVE:                                                        │  │
│   │  [Why this task exists, what data point it satisfies]                │  │
│   │                                                                       │  │
│   │  📊 CONTEXT:                                                          │  │
│   │  [All relevant information gathered from systems]                    │  │
│   │  • Source 1: [data]                                                  │  │
│   │  • Source 2: [data]                                                  │  │
│   │  • Source 3: [data]                                                  │  │
│   │                                                                       │  │
│   │  🤖 PROPOSED ACTION:                                                  │  │
│   │  [What the system recommends doing]                                  │  │
│   │                                                                       │  │
│   │  📝 OUTPUT PREVIEW:                                                   │  │
│   │  [What will be created/sent/modified]                                │  │
│   │                                                                       │  │
│   │  ⚡ SKILLS INVOLVED:                                                  │  │
│   │  [Which skills will execute]                                         │  │
│   │                                                                       │  │
│   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │  │
│   │                                                                       │  │
│   │  [APPROVE]  [EDIT]  [REJECT]  [ASK QUESTION]  [SKIP]                │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Role-Specific Human Thinking

Different roles need to **think** about different things. The system facilitates this:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     ROLE-SPECIFIC THINKING                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   @marketing-role needs human to think about:                               │
│   ════════════════════════════════════════════                              │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  TASK: Create campaign landing page                                   │  │
│   │                                                                       │  │
│   │  CONTEXT:                                                            │  │
│   │  • Campaign: Q1 Product Launch                                       │  │
│   │  • Target: Enterprise CTOs                                           │  │
│   │  • Budget: $50K                                                      │  │
│   │  • Timeline: 2 weeks                                                 │  │
│   │                                                                       │  │
│   │  🧠 HUMAN INPUT NEEDED:                                               │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  The system needs your creative direction:                     │  │  │
│   │  │                                                                │  │  │
│   │  │  1. Visual style? [Modern/Corporate/Bold/Minimal]             │  │  │
│   │  │  2. Key message? [_______________________]                    │  │  │
│   │  │  3. Primary CTA? [Demo/Trial/Contact/Download]                │  │  │
│   │  │  4. Hero image concept? [_______________________]             │  │  │
│   │  │                                                                │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   │  Once you provide direction, skills will:                            │  │
│   │  • Generate copy variations                                          │  │
│   │  • Create design mockups                                             │  │
│   │  • Set up landing page                                               │  │
│   │  • Configure analytics                                               │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   @sales-role needs human to think about:                                   │
│   ═══════════════════════════════════════                                   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  TASK: Respond to pricing objection                                   │  │
│   │                                                                       │  │
│   │  CONTEXT:                                                            │  │
│   │  • Deal: Acme Corp - $100K                                           │  │
│   │  • Objection: "Too expensive vs. competitor"                         │  │
│   │  • Competitor: $75K for similar scope                                │  │
│   │  • Deal history: 3 meetings, demo completed                          │  │
│   │                                                                       │  │
│   │  🧠 HUMAN INPUT NEEDED:                                               │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  Strategy decision:                                            │  │  │
│   │  │                                                                │  │  │
│   │  │  ○ Hold price, emphasize value                                │  │  │
│   │  │  ○ Offer discount: [___]%                                     │  │  │
│   │  │  ○ Restructure deal (payment terms)                           │  │  │
│   │  │  ○ Add value (extra services)                                 │  │  │
│   │  │  ○ Walk away                                                  │  │  │
│   │  │                                                                │  │  │
│   │  │  Key point to emphasize? [_______________________]            │  │  │
│   │  │                                                                │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   │  Once you decide, skills will:                                       │  │
│   │  • Draft response email                                              │  │
│   │  • Update deal in CRM                                                │  │
│   │  • Prepare supporting materials                                      │  │
│   │  • Schedule follow-up                                                │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Skill Examples: Simple to Complex

### Example 1: Simple Skill (Tier 1 - Deterministic)

A straightforward skill with hardcoded workflow:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                 SIMPLE SKILL: slack_send_channel_message                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   METADATA                                                                  │
│   ════════                                                                  │
│   ID:             slack_send_channel_message_v2                            │
│   Name:           Send Slack Channel Message                               │
│   Tier:           1 (Deterministic)                                        │
│   Model:          NONE (pure code)                                         │
│   Success Rate:   100% (when Slack API is up)                              │
│   Avg Execution:  ~200ms                                                   │
│   Role:           @any                                                     │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   WORKFLOW (Hardcoded Process)                                              │
│   ═════════════════════════════                                             │
│                                                                             │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐            │
│   │  Input  │────▶│  Auth   │────▶│  Send   │────▶│ Return  │            │
│   │ Params  │     │  Slack  │     │ Message │     │ Result  │            │
│   └─────────┘     └─────────┘     └─────────┘     └─────────┘            │
│                                                                             │
│   Step 1: Receive params (channel_id, message, optional: attachments)      │
│   Step 2: Get Slack credentials from vault                                 │
│   Step 3: Call Slack API chat.postMessage                                  │
│   Step 4: Return message ID and timestamp                                  │
│                                                                             │
│   NO AI INVOLVED. NO DECISIONS. PURE EXECUTION.                            │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   CODE                                                                      │
│   ════                                                                      │

```python
# skills/slack/send_channel_message_v2.py
"""
Skill: Send Slack Channel Message
Tier: 1 (Deterministic - 100% success when API is up)
"""

from sorx import get_credential, return_result

def execute(params: dict):
    # ═══════════════════════════════════════════════════════
    # STEP 1: VALIDATE INPUT (Hardcoded)
    # ═══════════════════════════════════════════════════════
    channel_id = params.get("channel_id")
    message = params.get("message")

    if not channel_id or not message:
        return_result({
            "success": False,
            "error": "channel_id and message are required"
        })
        return

    # ═══════════════════════════════════════════════════════
    # STEP 2: GET CREDENTIALS (Hardcoded)
    # ═══════════════════════════════════════════════════════
    slack_creds = get_credential("slack")
    token = slack_creds["bot_token"]

    # ═══════════════════════════════════════════════════════
    # STEP 3: SEND MESSAGE (Hardcoded)
    # ═══════════════════════════════════════════════════════
    import requests

    response = requests.post(
        "https://slack.com/api/chat.postMessage",
        headers={"Authorization": f"Bearer {token}"},
        json={
            "channel": channel_id,
            "text": message,
            "attachments": params.get("attachments", [])
        }
    )

    result = response.json()

    # ═══════════════════════════════════════════════════════
    # STEP 4: RETURN RESULT (Hardcoded)
    # ═══════════════════════════════════════════════════════
    if result.get("ok"):
        return_result({
            "success": True,
            "message_ts": result["ts"],
            "channel": result["channel"]
        })
    else:
        return_result({
            "success": False,
            "error": result.get("error", "Unknown error")
        })

metadata = {
    "id": "slack_send_channel_message_v2",
    "tier": 1,
    "model": None,
    "credentials_needed": ["slack"],
    "data_points_satisfied": ["notification.sent"],
    "role_affinity": ["any"]
}
```

│                                                                             │
│   WHAT HUMAN SEES                                                           │
│   ═══════════════                                                           │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  ✅ ACTION COMPLETED                                                  │  │
│   │                                                                       │  │
│   │  Sent message to #sales-wins:                                        │  │
│   │  "New deal closed: Acme Corp - $50,000!"                            │  │
│   │                                                                       │  │
│   │  Message ID: 1234567890.123456                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Example 2: Complex Skill (Tier 3/4 - Multi-Connector, Agentic)

A sophisticated skill that:
- Connects to multiple data sources
- Makes agentic calls to other skills
- Transforms and combines data
- Produces a complex output
- Requires human thinking/input

```
┌─────────────────────────────────────────────────────────────────────────────┐
│     COMPLEX SKILL: quarterly_business_review_generator                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   METADATA                                                                  │
│   ════════                                                                  │
│   ID:             quarterly_business_review_generator_v3                   │
│   Name:           Generate Quarterly Business Review                       │
│   Tier:           3-4 (Reasoning + Generative)                             │
│   Model:          OPUS (complex reasoning required)                         │
│   Success Rate:   85-90%                                                   │
│   Avg Execution:  ~45 seconds                                              │
│   Role:           @ops, @finance, @executive                               │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   DATA CONNECTORS (5 sources)                                               │
│   ═══════════════════════════                                               │
│                                                                             │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│   │ HubSpot │  │QuickBooks│  │ ClickUp │  │  Slack  │  │  Notion │        │
│   │  (CRM)  │  │(Finance) │  │ (Tasks) │  │ (Comms) │  │ (Docs)  │        │
│   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│        │            │            │            │            │              │
│        └────────────┴────────────┴────────────┴────────────┘              │
│                                   │                                        │
│                                   ▼                                        │
│                          ┌───────────────┐                                 │
│                          │  Data Fusion  │                                 │
│                          │    Engine     │                                 │
│                          └───────────────┘                                 │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   WORKFLOW (Complex, with Agentic Calls)                                    │
│   ══════════════════════════════════════                                    │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  PHASE 1: DATA COLLECTION (Parallel)                                 │  │
│   │  ════════════════════════════════════                                │  │
│   │                                                                       │  │
│   │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │  │
│   │  │ HubSpot:  │  │QuickBooks:│  │ ClickUp:  │  │  Slack:   │        │  │
│   │  │ Deals,    │  │ Revenue,  │  │ Projects, │  │ Sentiment,│        │  │
│   │  │ Pipeline, │  │ Expenses, │  │ Tasks,    │  │ Activity  │        │  │
│   │  │ Clients   │  │ Cash Flow │  │ Velocity  │  │ Metrics   │        │  │
│   │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘        │  │
│   │        │              │              │              │              │  │
│   │        └──────────────┴──────────────┴──────────────┘              │  │
│   │                              │                                      │  │
│   │                              ▼                                      │  │
│   │  ┌─────────────────────────────────────────────────────────────┐  │  │
│   │  │                  RAW DATA COLLECTED                          │  │  │
│   │  │                                                              │  │  │
│   │  │  HubSpot:    45 deals, $1.2M pipeline, 23 new clients       │  │  │
│   │  │  QuickBooks: $890K revenue, $650K expenses, $240K profit    │  │  │
│   │  │  ClickUp:    156 tasks completed, 89% on-time delivery      │  │  │
│   │  │  Slack:      12,456 messages, positive sentiment 78%         │  │  │
│   │  │                                                              │  │  │
│   │  └─────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                   │                                        │
│                                   ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  PHASE 2: AGENTIC ANALYSIS (Sequential)                              │  │
│   │  ═══════════════════════════════════════                             │  │
│   │                                                                       │  │
│   │  request_agent("@data-analyst", "Analyze revenue trends")           │  │
│   │       │                                                              │  │
│   │       ▼                                                              │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  @data-analyst response:                                       │  │  │
│   │  │  "Revenue up 15% QoQ, driven by enterprise segment.           │  │  │
│   │  │   SMB flat. Churn increased 2% - concerning trend."           │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │       │                                                              │  │
│   │       ▼                                                              │  │
│   │  request_agent("@finance-analyst", "Analyze profitability")         │  │
│   │       │                                                              │  │
│   │       ▼                                                              │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  @finance-analyst response:                                    │  │  │
│   │  │  "Gross margin improved to 72%. CAC stable at $1,200.         │  │  │
│   │  │   LTV:CAC ratio at 4.2x - healthy. OpEx up 8%."               │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │       │                                                              │  │
│   │       ▼                                                              │  │
│   │  request_agent("@ops-analyst", "Analyze team performance")          │  │
│   │       │                                                              │  │
│   │       ▼                                                              │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  @ops-analyst response:                                        │  │  │
│   │  │  "Task velocity up 23%. On-time delivery at 89%.              │  │  │
│   │  │   Engineering bottleneck identified. Sales team exceeded."    │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                   │                                        │
│                                   ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  PHASE 3: HUMAN INPUT REQUIRED                                       │  │
│   │  ═══════════════════════════════                                     │  │
│   │                                                                       │  │
│   │  ⏸️ SKILL PAUSED - WAITING FOR HUMAN                                 │  │
│   │                                                                       │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │                                                                │  │  │
│   │  │  📋 CONTEXT GATHERED:                                          │  │  │
│   │  │  • Revenue: $890K (+15% QoQ)                                  │  │  │
│   │  │  • Profit: $240K (27% margin)                                 │  │  │
│   │  │  • New Clients: 23                                            │  │  │
│   │  │  • Team Performance: 89% on-time                              │  │  │
│   │  │  • Key Concern: Churn up 2%, Engineering bottleneck           │  │  │
│   │  │                                                                │  │  │
│   │  │  🧠 YOUR INPUT NEEDED:                                         │  │  │
│   │  │                                                                │  │  │
│   │  │  1. Key wins to highlight?                                    │  │  │
│   │  │     [________________________________________]                │  │  │
│   │  │                                                                │  │  │
│   │  │  2. Challenges to address?                                    │  │  │
│   │  │     [________________________________________]                │  │  │
│   │  │                                                                │  │  │
│   │  │  3. Strategic priorities for next quarter?                    │  │  │
│   │  │     [________________________________________]                │  │  │
│   │  │                                                                │  │  │
│   │  │  4. Audience for this QBR?                                    │  │  │
│   │  │     ○ Board of Directors                                      │  │  │
│   │  │     ○ Leadership Team                                         │  │  │
│   │  │     ○ All Hands                                               │  │  │
│   │  │     ○ Investors                                               │  │  │
│   │  │                                                                │  │  │
│   │  │  5. Tone?                                                     │  │  │
│   │  │     ○ Optimistic                                              │  │  │
│   │  │     ○ Balanced                                                │  │  │
│   │  │     ○ Cautious                                                │  │  │
│   │  │                                                                │  │  │
│   │  │  [CONTINUE WITH MY INPUT]  [SKIP - USE DEFAULTS]             │  │  │
│   │  │                                                                │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                   │                                        │
│                          (Human provides input)                            │
│                                   │                                        │
│                                   ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  PHASE 4: DOCUMENT GENERATION (AI-Powered)                           │  │
│   │  ═════════════════════════════════════════                           │  │
│   │                                                                       │  │
│   │  Using OPUS model with:                                              │  │
│   │  • All collected data                                                │  │
│   │  • Agent analysis                                                    │  │
│   │  • Human input/direction                                             │  │
│   │                                                                       │  │
│   │  Generate:                                                           │  │
│   │  ├─ Executive Summary                                                │  │
│   │  ├─ Financial Overview (charts, trends)                              │  │
│   │  ├─ Sales Performance (pipeline, wins, losses)                       │  │
│   │  ├─ Operations Review (team, delivery, efficiency)                   │  │
│   │  ├─ Challenges & Risks                                               │  │
│   │  ├─ Strategic Priorities                                             │  │
│   │  └─ Next Quarter Goals                                               │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                   │                                        │
│                                   ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  PHASE 5: OUTPUT & DISTRIBUTION                                      │  │
│   │  ══════════════════════════════                                      │  │
│   │                                                                       │  │
│   │  request_skill("notion_create_page", {...})                          │  │
│   │       └─▶ Creates QBR document in Notion                             │  │
│   │                                                                       │  │
│   │  request_skill("google_slides_create", {...})                        │  │
│   │       └─▶ Creates presentation deck                                  │  │
│   │                                                                       │  │
│   │  request_skill("gmail_send_with_attachment", {...})                  │  │
│   │       └─▶ Sends to stakeholders (after human approval)               │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                   │                                        │
│                                   ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                       │  │
│   │  FINAL OUTPUT (Presented to Human)                                   │  │
│   │  ══════════════════════════════════                                  │  │
│   │                                                                       │  │
│   │  ┌───────────────────────────────────────────────────────────────┐  │  │
│   │  │  ✅ QBR GENERATED                                              │  │  │
│   │  │                                                                │  │  │
│   │  │  DELIVERABLES:                                                │  │  │
│   │  │  📄 Notion Document: Q4 2024 Business Review                  │  │  │
│   │  │     → [View Document]                                         │  │  │
│   │  │                                                                │  │  │
│   │  │  📊 Presentation: Q4_2024_QBR_Deck.pptx                       │  │  │
│   │  │     → [View]  [Download]  [Edit]                              │  │  │
│   │  │                                                                │  │  │
│   │  │  📧 READY TO SEND:                                            │  │  │
│   │  │  • Board of Directors (5 recipients)                          │  │  │
│   │  │  • Subject: "Q4 2024 Quarterly Business Review"               │  │  │
│   │  │                                                                │  │  │
│   │  │  [REVIEW EMAIL]  [SEND NOW]  [SCHEDULE]  [EDIT]              │  │  │
│   │  │                                                                │  │  │
│   │  └───────────────────────────────────────────────────────────────┘  │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Complex Skill Code Structure

```python
# skills/business/quarterly_business_review_generator_v3.py
"""
Skill: Generate Quarterly Business Review
Tier: 3-4 (Reasoning + Generative)
Model: OPUS

This is a COMPLEX skill that:
1. Connects to 5 data sources
2. Makes agentic calls to specialized analysts
3. Requires human input for direction
4. Generates multi-format output
5. Handles distribution
"""

from sorx import (
    get_credential,
    return_result,
    request_skill,      # Call other skills
    request_agent,      # Call agents for analysis
    request_decision,   # Get human input
    notify_user,        # Update human on progress
    get_context,        # Get user context
)
from datetime import datetime, timedelta

def execute(params: dict):
    quarter = params.get("quarter", "Q4")
    year = params.get("year", 2024)

    notify_user(f"Starting {quarter} {year} QBR generation...", "info")

    # ═══════════════════════════════════════════════════════════════════
    # PHASE 1: PARALLEL DATA COLLECTION
    # ═══════════════════════════════════════════════════════════════════

    # These run in parallel (fire off all, collect results)
    hubspot_data = request_skill(
        skill="hubspot_get_quarterly_metrics",
        params={"quarter": quarter, "year": year}
    )

    quickbooks_data = request_skill(
        skill="quickbooks_get_financial_summary",
        params={"quarter": quarter, "year": year}
    )

    clickup_data = request_skill(
        skill="clickup_get_team_metrics",
        params={"quarter": quarter, "year": year}
    )

    slack_data = request_skill(
        skill="slack_get_activity_metrics",
        params={"quarter": quarter, "year": year}
    )

    notify_user("Data collected from all sources", "success")

    # ═══════════════════════════════════════════════════════════════════
    # PHASE 2: AGENTIC ANALYSIS
    # ═══════════════════════════════════════════════════════════════════

    revenue_analysis = request_agent(
        agent="@data-analyst",
        task="Analyze revenue trends and identify key drivers",
        context={
            "hubspot": hubspot_data,
            "quickbooks": quickbooks_data
        }
    )

    profitability_analysis = request_agent(
        agent="@finance-analyst",
        task="Analyze profitability, margins, and financial health",
        context={"quickbooks": quickbooks_data}
    )

    ops_analysis = request_agent(
        agent="@ops-analyst",
        task="Analyze team performance, velocity, and bottlenecks",
        context={
            "clickup": clickup_data,
            "slack": slack_data
        }
    )

    notify_user("Analysis complete", "success")

    # ═══════════════════════════════════════════════════════════════════
    # PHASE 3: HUMAN INPUT (Pull - Wait for human)
    # ═══════════════════════════════════════════════════════════════════

    # Present context and ask for direction
    human_input = request_decision(
        question="Please provide direction for the QBR",
        context={
            "summary": {
                "revenue": quickbooks_data.get("revenue"),
                "profit": quickbooks_data.get("profit"),
                "new_clients": hubspot_data.get("new_clients"),
                "team_performance": clickup_data.get("on_time_rate"),
            },
            "analysis": {
                "revenue": revenue_analysis,
                "profitability": profitability_analysis,
                "operations": ops_analysis
            }
        },
        input_fields=[
            {"id": "wins", "label": "Key wins to highlight", "type": "text"},
            {"id": "challenges", "label": "Challenges to address", "type": "text"},
            {"id": "priorities", "label": "Strategic priorities", "type": "text"},
            {"id": "audience", "label": "Audience", "type": "select",
             "options": ["Board", "Leadership", "All Hands", "Investors"]},
            {"id": "tone", "label": "Tone", "type": "select",
             "options": ["Optimistic", "Balanced", "Cautious"]}
        ]
    )

    # ═══════════════════════════════════════════════════════════════════
    # PHASE 4: DOCUMENT GENERATION (Uses OPUS for complex reasoning)
    # ═══════════════════════════════════════════════════════════════════

    # Combine all data and human direction
    qbr_content = generate_qbr_document(
        data={
            "hubspot": hubspot_data,
            "quickbooks": quickbooks_data,
            "clickup": clickup_data,
            "slack": slack_data
        },
        analysis={
            "revenue": revenue_analysis,
            "profitability": profitability_analysis,
            "operations": ops_analysis
        },
        human_input=human_input,
        quarter=quarter,
        year=year
    )

    # ═══════════════════════════════════════════════════════════════════
    # PHASE 5: CREATE OUTPUTS
    # ═══════════════════════════════════════════════════════════════════

    # Create Notion document
    notion_result = request_skill(
        skill="notion_create_page",
        params={
            "parent_id": "qbr_folder_id",
            "title": f"{quarter} {year} Quarterly Business Review",
            "content": qbr_content["document"]
        }
    )

    # Create presentation
    slides_result = request_skill(
        skill="google_slides_create_from_template",
        params={
            "template_id": "qbr_template",
            "title": f"{quarter} {year} QBR Deck",
            "data": qbr_content["slides_data"]
        }
    )

    # Prepare email (but don't send - human approves)
    email_draft = request_skill(
        skill="gmail_create_draft",
        params={
            "to": get_stakeholder_emails(human_input["audience"]),
            "subject": f"{quarter} {year} Quarterly Business Review",
            "body": qbr_content["email_body"],
            "attachments": [slides_result["file_url"]]
        }
    )

    # ═══════════════════════════════════════════════════════════════════
    # FINAL: RETURN RESULTS TO HUMAN
    # ═══════════════════════════════════════════════════════════════════

    return_result({
        "success": True,
        "outputs": {
            "notion_document": {
                "url": notion_result["url"],
                "title": f"{quarter} {year} Quarterly Business Review"
            },
            "presentation": {
                "url": slides_result["url"],
                "title": f"{quarter}_{year}_QBR_Deck"
            },
            "email_draft": {
                "draft_id": email_draft["draft_id"],
                "recipients": email_draft["to"],
                "subject": email_draft["subject"]
            }
        },
        "data_points_satisfied": [
            "qbr.document_created",
            "qbr.presentation_created",
            "qbr.email_drafted"
        ],
        "awaiting_human_action": ["email.send_approval"]
    })


def generate_qbr_document(data, analysis, human_input, quarter, year):
    """
    Uses OPUS model to generate the QBR document.
    This is where the complex reasoning happens.
    """
    # LLM call to generate document structure and content
    # Based on all inputs
    pass


def get_stakeholder_emails(audience):
    """Get email list based on audience selection."""
    audiences = {
        "Board": ["board@company.com"],
        "Leadership": ["leadership@company.com"],
        "All Hands": ["all@company.com"],
        "Investors": ["investors@company.com"]
    }
    return audiences.get(audience, [])


metadata = {
    "id": "quarterly_business_review_generator_v3",
    "name": "Generate Quarterly Business Review",
    "tier": 4,
    "model": "opus",
    "credentials_needed": ["hubspot", "quickbooks", "clickup", "slack", "notion", "google"],
    "data_connectors": ["hubspot", "quickbooks", "clickup", "slack", "notion", "google_slides", "gmail"],
    "agent_calls": ["@data-analyst", "@finance-analyst", "@ops-analyst"],
    "human_input_required": True,
    "role_affinity": ["ops", "finance", "executive"]
}
```

---

## Core Concept: Skills

### What is a Skill?

A **Skill** is a learned capability that an agent acquires to perform a specific task on a specific system. Skills are:

1. **Acquired** - Generated the first time an agent needs to do something
2. **Saved** - Stored for reuse
3. **Evolved** - Improved based on feedback and experience
4. **Shared** - Can be shared across agents and organizations
5. **Role-Based** - Different agent roles have different skill sets

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SKILL LIFECYCLE                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   1. ACQUISITION                                                            │
│      User: "Send an email to John with the project update"                  │
│      Agent: "I don't have this skill yet. Let me learn it..."              │
│      [Agent generates script based on pattern templates]                    │
│      [Skill saved: "gmail_send_email"]                                      │
│                                                                             │
│   2. EXECUTION                                                              │
│      User: "Email the team about the meeting"                              │
│      Agent: "I have the gmail_send_email skill. Executing..."              │
│      [Reuses saved skill with new parameters]                              │
│                                                                             │
│   3. EVOLUTION                                                              │
│      User: "That email didn't include attachments properly"                │
│      Agent: "Let me improve the skill..."                                  │
│      [Updates skill to handle attachments better]                          │
│      [Skill version incremented]                                           │
│                                                                             │
│   4. MASTERY                                                                │
│      After 50 executions with 98% success rate:                            │
│      [Skill marked as "mastered"]                                          │
│      [Can be shared to skill library]                                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Structure

```python
class Skill:
    # Identity
    id: str                      # "gmail_send_email_v3"
    name: str                    # "Send Email via Gmail"
    description: str             # "Sends an email using Gmail API"

    # Classification
    provider: str                # "google"
    category: str                # "communication"
    role_affinity: list[str]     # ["sales", "support", "marketing"]

    # The actual capability
    script: str                  # Python code to execute
    pattern_used: str            # Which pattern template was used
    interface_type: str          # "rest_api", "database", "desktop_app", etc.

    # Requirements
    credentials_needed: list[str]  # ["google"]
    dependencies: list[str]        # ["requests", "google-auth"]

    # Learning data
    version: int                 # 3
    executions: int              # 47
    success_rate: float          # 0.98
    avg_execution_time: float    # 1.2 seconds
    last_improved: datetime
    improvement_notes: list[str]

    # Evolution history
    previous_versions: list[str]  # Links to older versions
    failure_patterns: list[str]   # What went wrong before
    optimization_history: list[str]  # How it improved

    # Sharing
    is_public: bool              # Available in skill library?
    usage_count: int             # Times used by others
    rating: float                # Community rating
```

---

## Architecture

### The Dynamic Integration Protocol (DIP)

DIP is the underlying protocol that enables Sorx 2.0 to connect to any system:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SORX 2.0 ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   CLOUD LAYER (BusinessOS Backend)                                          │
│   ════════════════════════════════                                          │
│                                                                             │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   │
│   │   Agent     │   │   Skill     │   │  Pattern    │   │  Skill      │   │
│   │   Router    │──▶│  Generator  │◀──│  Library    │   │  Library    │   │
│   └──────┬──────┘   └──────┬──────┘   └─────────────┘   └──────┬──────┘   │
│          │                 │                                    │          │
│          │    ┌────────────┴────────────┐                      │          │
│          │    │                         │                      │          │
│          ▼    ▼                         ▼                      ▼          │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                      Skill Orchestrator                              │ │
│   │  - Routes requests to appropriate skills                            │ │
│   │  - Manages skill acquisition, execution, evolution                  │ │
│   │  - Tracks skill performance and learning                            │ │
│   └─────────────────────────────────────┬───────────────────────────────┘ │
│                                         │                                  │
│                                         │ WebSocket / SSE                  │
│                                         ▼                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   LOCAL LAYER (User's Machine - Desktop App / MIOSA OS)                     │
│   ══════════════════════════════════════════════════════                    │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                        Sorx Engine                                    │ │
│   │                                                                       │ │
│   │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │ │
│   │  │  Credential   │  │    Skill      │  │   Execution   │           │ │
│   │  │    Vault      │  │    Cache      │  │   Sandbox     │           │ │
│   │  └───────────────┘  └───────────────┘  └───────────────┘           │ │
│   │                                                                       │ │
│   │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │ │
│   │  │   Interface   │  │   Protocol    │  │   Result      │           │ │
│   │  │   Adapters    │  │   Handlers    │  │   Processor   │           │ │
│   │  └───────────────┘  └───────────────┘  └───────────────┘           │ │
│   │                                                                       │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                         │                                  │
│                                         ▼                                  │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                     TARGET SYSTEMS                                    │ │
│   │  APIs │ Databases │ Desktop Apps │ Files │ Hardware │ Legacy        │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Orchestration Layer - The Callback Protocol

### The Problem

Skills don't operate in isolation. A skill executing on the Sorx Engine often needs:
- Data from another system that requires a different skill
- A decision from an AI agent in BusinessOS
- Context from the current conversation
- Results from another agent's work

### The Solution: Bidirectional Orchestration

The **Orchestration Layer** provides real-time, bidirectional communication between BusinessOS and the Sorx Engine, enabling skills to **callback** to the operating system mid-execution.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        ORCHESTRATION ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   BusinessOS (Cloud)                                                        │
│   ══════════════════                                                        │
│                                                                             │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                 │
│   │   Agent     │     │   Agent     │     │   Agent     │                 │
│   │   Router    │◀───▶│  Workspace  │◀───▶│   Memory    │                 │
│   └──────┬──────┘     └──────┬──────┘     └─────────────┘                 │
│          │                   │                                              │
│          ▼                   ▼                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                    ORCHESTRATION HUB                                  │ │
│   │                                                                       │ │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │ │
│   │  │  Message    │  │  Callback   │  │  Context    │                  │ │
│   │  │    Bus      │  │   Router    │  │   Manager   │                  │ │
│   │  └─────────────┘  └─────────────┘  └─────────────┘                  │ │
│   │                                                                       │ │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │ │
│   │  │  Execution  │  │   Agent     │  │   Result    │                  │ │
│   │  │   Tracker   │  │  Dispatcher │  │  Aggregator │                  │ │
│   │  └─────────────┘  └─────────────┘  └─────────────┘                  │ │
│   │                                                                       │ │
│   └──────────────────────────────┬──────────────────────────────────────┘ │
│                                  │                                         │
│                                  │ WebSocket (bidirectional)               │
│                                  │                                         │
├──────────────────────────────────┼─────────────────────────────────────────┤
│                                  │                                         │
│   Sorx Engine (Local)            │                                         │
│   ═══════════════════            ▼                                         │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐ │
│   │                     CALLBACK HANDLER                                  │ │
│   │                                                                       │ │
│   │  - Receives callbacks from executing skills                          │ │
│   │  - Forwards requests to BusinessOS                                   │ │
│   │  - Receives responses and resumes skill execution                    │ │
│   │  - Maintains execution context while waiting                         │ │
│   │                                                                       │ │
│   └──────────────────────────────┬──────────────────────────────────────┘ │
│                                  │                                         │
│   ┌──────────────────────────────┼──────────────────────────────────────┐ │
│   │                              ▼                                       │ │
│   │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐              │ │
│   │   │ Skill 1 │  │ Skill 2 │  │ Skill 3 │  │ Skill 4 │              │ │
│   │   │ (active)│  │(waiting)│  │ (active)│  │(pending)│              │ │
│   │   └─────────┘  └─────────┘  └─────────┘  └─────────┘              │ │
│   │                     ▲                                                │ │
│   │                     │                                                │ │
│   │                  PAUSED                                              │ │
│   │            (waiting for callback)                                    │ │
│   │                                                                       │ │
│   └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Callback Flow Example

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              EXAMPLE: CREATE INVOICE (with callback for client data)        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   BusinessOS                               Sorx Engine                      │
│   ══════════                               ════════════                     │
│       │                                         │                           │
│       │  1. EXECUTE                             │                           │
│       │  ─────────────────────────────────────▶ │                           │
│       │  {                                      │                           │
│       │    "skill": "quickbooks_create_invoice",│                           │
│       │    "params": {"client_id": "C-123"}     │                           │
│       │    "execution_id": "exec-abc"           │                           │
│       │  }                                      │                           │
│       │                                         │                           │
│       │                                         │ 2. Skill starts...        │
│       │                                         │    Needs billing info     │
│       │                                         │    for client C-123       │
│       │                                         │                           │
│       │  3. CALLBACK                            │                           │
│       │  ◀───────────────────────────────────── │                           │
│       │  {                                      │                           │
│       │    "type": "agent_request",             │                           │
│       │    "execution_id": "exec-abc",          │                           │
│       │    "request": {                         │                           │
│       │      "need": "client_billing_info",     │                           │
│       │      "client_id": "C-123",              │                           │
│       │      "fields": ["address", "tax_id",    │                           │
│       │                 "payment_terms"]        │                           │
│       │    }                                    │                           │
│       │  }                                      │                           │
│       │                                         │                           │
│       │  4. BusinessOS routes to                │                           │
│       │     @data-agent                         │                           │
│       │         │                               │                           │
│       │         ▼                               │                           │
│       │  5. @data-agent decides to use          │                           │
│       │     hubspot_get_client skill            │                           │
│       │  ─────────────────────────────────────▶ │                           │
│       │  {                                      │                           │
│       │    "skill": "hubspot_get_client",       │                           │
│       │    "params": {"id": "C-123"},           │                           │
│       │    "execution_id": "exec-def",          │ 6. Skill executes         │
│       │    "parent_execution": "exec-abc"       │    fetches from HubSpot   │
│       │  }                                      │                           │
│       │                                         │                           │
│       │  7. Client data returned                │                           │
│       │  ◀───────────────────────────────────── │                           │
│       │  {                                      │                           │
│       │    "execution_id": "exec-def",          │                           │
│       │    "result": {                          │                           │
│       │      "address": "123 Main St...",       │                           │
│       │      "tax_id": "XX-XXXXXXX",            │                           │
│       │      "payment_terms": "net-30"          │                           │
│       │    }                                    │                           │
│       │  }                                      │                           │
│       │                                         │                           │
│       │  8. RESUME with data                    │                           │
│       │  ─────────────────────────────────────▶ │                           │
│       │  {                                      │                           │
│       │    "type": "callback_response",         │                           │
│       │    "execution_id": "exec-abc",          │ 9. Original skill         │
│       │    "data": {                            │    resumes with data      │
│       │      "address": "123 Main St...",       │    creates invoice        │
│       │      "tax_id": "XX-XXXXXXX",            │                           │
│       │      "payment_terms": "net-30"          │                           │
│       │    }                                    │                           │
│       │  }                                      │                           │
│       │                                         │                           │
│       │  10. COMPLETE                           │                           │
│       │  ◀───────────────────────────────────── │                           │
│       │  {                                      │                           │
│       │    "execution_id": "exec-abc",          │                           │
│       │    "status": "success",                 │                           │
│       │    "result": {"invoice_id": "INV-789"}  │                           │
│       │  }                                      │                           │
│       │                                         │                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Message Types

```python
# orchestration/messages.py

from enum import Enum
from dataclasses import dataclass
from typing import Any, Optional
from datetime import datetime

class MessageType(Enum):
    # BusinessOS → Sorx Engine
    EXECUTE_SKILL = "execute_skill"           # Start a skill
    CALLBACK_RESPONSE = "callback_response"   # Response to a callback
    CANCEL_EXECUTION = "cancel_execution"     # Cancel running skill

    # Sorx Engine → BusinessOS
    EXECUTION_STARTED = "execution_started"   # Skill started
    CALLBACK_REQUEST = "callback_request"     # Skill needs something
    EXECUTION_PROGRESS = "execution_progress" # Progress update
    EXECUTION_COMPLETE = "execution_complete" # Skill finished
    EXECUTION_ERROR = "execution_error"       # Skill failed

class CallbackType(Enum):
    AGENT_REQUEST = "agent_request"           # Need an agent to do something
    SKILL_REQUEST = "skill_request"           # Need another skill executed
    DATA_REQUEST = "data_request"             # Need data from BusinessOS
    DECISION_REQUEST = "decision_request"     # Need a decision from agent
    CONTEXT_REQUEST = "context_request"       # Need conversation context
    USER_INPUT = "user_input"                 # Need input from user

@dataclass
class ExecutionContext:
    """Preserved state while waiting for callback response."""
    execution_id: str
    skill_id: str
    started_at: datetime
    params: dict
    local_state: dict              # Variables, loop counters, etc.
    stack_position: int            # Where to resume in the script
    parent_execution: Optional[str] # If this is a nested call
    callback_chain: list[str]       # History of callbacks made

@dataclass
class CallbackRequest:
    """Request from skill back to BusinessOS."""
    execution_id: str
    callback_type: CallbackType
    request_data: dict
    timeout_ms: int = 30000        # How long to wait before failing
    priority: str = "normal"       # "high" for user-facing operations

    # For agent requests
    preferred_agent: Optional[str] = None   # "@data-agent", "@sales-agent"

    # For skill requests
    nested_skill: Optional[str] = None      # "hubspot_get_client"
    nested_params: Optional[dict] = None

@dataclass
class CallbackResponse:
    """Response from BusinessOS back to waiting skill."""
    execution_id: str
    success: bool
    data: Any                      # The requested data/result
    error: Optional[str] = None
    metadata: Optional[dict] = None
```

### Callback Runtime Functions

Skills use these functions to interact with BusinessOS:

```python
# sorx_runtime/callbacks.py

from sorx import (
    get_credential,       # Already existed - local credential access
    return_result,        # Already existed - return final result
    trigger_skill,        # Already existed - fire-and-forget skill
)

# NEW: Callback functions for bidirectional communication

def request_agent(
    agent: str,
    task: str,
    context: dict = None,
    wait: bool = True,
    timeout_ms: int = 30000
) -> dict:
    """
    Request an agent in BusinessOS to perform a task.

    Args:
        agent: Which agent to invoke ("@data-agent", "@sales-agent", etc.)
        task: Natural language description of what's needed
        context: Additional context for the agent
        wait: If True, pause until agent responds. If False, fire-and-forget.
        timeout_ms: How long to wait for response

    Returns:
        Agent's response data

    Example:
        client_info = request_agent(
            agent="@data-agent",
            task="Get billing information for client C-123",
            context={"fields_needed": ["address", "tax_id"]}
        )
    """
    pass

def request_skill(
    skill: str,
    params: dict,
    wait: bool = True,
    timeout_ms: int = 30000
) -> dict:
    """
    Request another skill to be executed and return its result.

    Unlike trigger_skill() which is fire-and-forget, this waits for the result.

    Args:
        skill: Skill ID to execute
        params: Parameters for the skill
        wait: If True, pause until skill completes
        timeout_ms: How long to wait

    Returns:
        Skill execution result

    Example:
        client = request_skill(
            skill="hubspot_get_client",
            params={"client_id": "C-123"}
        )
        # Now use client data in current skill
    """
    pass

def request_data(
    data_type: str,
    query: dict,
    source: str = None
) -> dict:
    """
    Request data from BusinessOS (database, memory, etc.).

    Args:
        data_type: Type of data needed ("client", "project", "user", etc.)
        query: Query parameters
        source: Specific source if known ("hubspot", "database", etc.)

    Returns:
        Requested data

    Example:
        recent_invoices = request_data(
            data_type="invoices",
            query={"client_id": "C-123", "limit": 5},
            source="quickbooks"
        )
    """
    pass

def request_decision(
    question: str,
    options: list[str] = None,
    context: dict = None,
    agent: str = None
) -> str:
    """
    Request a decision from an agent or user.

    Args:
        question: What decision is needed
        options: If provided, constrain to these choices
        context: Additional context for decision
        agent: Specific agent to ask, or None for auto-routing

    Returns:
        The decision made

    Example:
        discount = request_decision(
            question="Client is asking for a discount. What should we offer?",
            options=["0%", "5%", "10%", "15%", "Escalate to manager"],
            context={"client_tier": "enterprise", "deal_size": 50000}
        )
    """
    pass

def get_context(
    context_type: str
) -> dict:
    """
    Get context from the current conversation or session.

    Args:
        context_type: What context is needed
            - "conversation": Current chat context
            - "user": Current user info
            - "session": Current session data
            - "workspace": Workspace settings

    Returns:
        Requested context data

    Example:
        conv = get_context("conversation")
        # Use conversation history to inform skill behavior
    """
    pass

def notify_user(
    message: str,
    notification_type: str = "info",
    actions: list[dict] = None
) -> None:
    """
    Send a notification to the user through BusinessOS.

    This doesn't pause execution - it's fire-and-forget.

    Args:
        message: Message to show user
        notification_type: "info", "success", "warning", "error"
        actions: Optional action buttons [{label, action_id}]

    Example:
        notify_user(
            message="Invoice created successfully!",
            notification_type="success",
            actions=[
                {"label": "View Invoice", "action_id": "view_inv_123"},
                {"label": "Email to Client", "action_id": "email_inv_123"}
            ]
        )
    """
    pass
```

### Skill Example with Callbacks

```python
# skills/quickbooks_create_invoice_v2.py
"""
Skill: Create Invoice in QuickBooks
Version: 2 (added callback for client data)
Category: finance
Provider: intuit

This skill demonstrates the callback pattern:
1. Starts creating invoice
2. Realizes it needs client billing info
3. Callbacks to BusinessOS to get info from HubSpot
4. Resumes and completes the invoice
"""

from sorx import (
    get_credential,
    return_result,
    request_skill,      # Callback to run another skill
    request_decision,   # Callback to get agent decision
    notify_user,        # Fire-and-forget notification
)
from quickbooks import QuickBooks

def execute(params: dict):
    client_id = params["client_id"]
    line_items = params.get("line_items", [])

    # Get QuickBooks credentials from local vault
    qb_creds = get_credential("quickbooks")
    qb = QuickBooks(
        client_id=qb_creds["client_id"],
        client_secret=qb_creds["client_secret"],
        refresh_token=qb_creds["refresh_token"]
    )

    # ═══════════════════════════════════════════════════════════════
    # CALLBACK 1: Get client billing info from HubSpot
    # ═══════════════════════════════════════════════════════════════
    # This pauses execution, sends request to BusinessOS,
    # BusinessOS routes to appropriate agent, agent runs skill,
    # result comes back, execution resumes here

    client_data = request_skill(
        skill="hubspot_get_client",
        params={
            "client_id": client_id,
            "fields": ["company", "address", "email", "tax_id", "payment_terms"]
        }
    )

    if not client_data.get("success"):
        return_result({
            "success": False,
            "error": f"Could not fetch client data: {client_data.get('error')}"
        })
        return

    # ═══════════════════════════════════════════════════════════════
    # CALLBACK 2: Decision on payment terms if not set
    # ═══════════════════════════════════════════════════════════════

    payment_terms = client_data.get("payment_terms")
    if not payment_terms:
        payment_terms = request_decision(
            question=f"Client {client_data['company']} has no payment terms set. What should we use?",
            options=["Net 15", "Net 30", "Net 45", "Net 60", "Due on Receipt"],
            context={
                "client_name": client_data["company"],
                "client_tier": client_data.get("tier", "unknown")
            }
        )

    # ═══════════════════════════════════════════════════════════════
    # Now create the invoice with all the data we gathered
    # ═══════════════════════════════════════════════════════════════

    invoice = qb.create_invoice(
        customer_email=client_data["email"],
        billing_address=client_data["address"],
        line_items=line_items,
        payment_terms=payment_terms,
        tax_id=client_data.get("tax_id")
    )

    # Fire-and-forget notification (doesn't pause execution)
    notify_user(
        message=f"Invoice {invoice['id']} created for {client_data['company']}",
        notification_type="success",
        actions=[
            {"label": "View Invoice", "action_id": f"view_invoice_{invoice['id']}"},
            {"label": "Send to Client", "action_id": f"email_invoice_{invoice['id']}"}
        ]
    )

    return_result({
        "success": True,
        "invoice_id": invoice["id"],
        "invoice_number": invoice["number"],
        "amount": invoice["total"],
        "client": client_data["company"],
        "payment_terms": payment_terms
    })

# Skill metadata
metadata = {
    "id": "quickbooks_create_invoice_v2",
    "name": "Create QuickBooks Invoice",
    "version": 2,
    "category": "finance",
    "provider": "intuit",
    "credentials_needed": ["quickbooks"],
    "callbacks_used": ["request_skill", "request_decision", "notify_user"],
    "role_affinity": ["operations", "finance"]
}
```

### Execution States

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SKILL EXECUTION STATES                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐        │
│   │ PENDING  │────▶│ RUNNING  │────▶│ COMPLETE │     │  FAILED  │        │
│   └──────────┘     └────┬─────┘     └──────────┘     └──────────┘        │
│                         │                                   ▲              │
│                         │                                   │              │
│                         ▼                                   │              │
│                    ┌──────────┐                             │              │
│                    │ WAITING  │◀───────────────────────────┘              │
│                    │ CALLBACK │                                            │
│                    └────┬─────┘                                            │
│                         │                                                  │
│                         │ callback_response                                │
│                         │                                                  │
│                         ▼                                                  │
│                    ┌──────────┐                                            │
│                    │ RESUMING │────────────────▶ RUNNING                   │
│                    └──────────┘                                            │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   State Descriptions:                                                       │
│   ───────────────────                                                       │
│   PENDING    - Skill queued, not yet started                               │
│   RUNNING    - Skill actively executing                                    │
│   WAITING    - Paused, waiting for callback response from BusinessOS       │
│   RESUMING   - Callback received, restoring context before continuing      │
│   COMPLETE   - Skill finished successfully                                 │
│   FAILED     - Skill failed (error, timeout, cancelled)                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Callback Chaining

Complex workflows involve multiple nested callbacks:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     CALLBACK CHAINING EXAMPLE                                │
│                   "Close Deal → Full Onboarding"                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   deal_won_automation (Skill A)                                             │
│   │                                                                         │
│   ├─▶ request_skill("hubspot_update_deal") ────────────────────────┐       │
│   │                                                  Skill B        │       │
│   │   ◀─── result: {deal_updated: true} ───────────────────────────┘       │
│   │                                                                         │
│   ├─▶ request_skill("slack_notify_team")                                   │
│   │       │                                                                 │
│   │       └─▶ request_data("team_members") ────────────────────────┐       │
│   │                                          BusinessOS lookup      │       │
│   │           ◀─── result: [{name, slack_id}, ...] ────────────────┘       │
│   │       │                                                                 │
│   │   ◀─── result: {notified: ["@john", "@jane"]}                          │
│   │                                                                         │
│   ├─▶ request_skill("quickbooks_create_invoice")                           │
│   │       │                                                                 │
│   │       └─▶ request_skill("hubspot_get_client") ─────────────────┐       │
│   │                                                  Skill C        │       │
│   │           ◀─── result: {company, address, ...} ────────────────┘       │
│   │       │                                                                 │
│   │       └─▶ request_decision("payment_terms?") ──────────────────┐       │
│   │                                          @finance-agent         │       │
│   │           ◀─── result: "Net 30" ───────────────────────────────┘       │
│   │       │                                                                 │
│   │   ◀─── result: {invoice_id: "INV-123"}                                 │
│   │                                                                         │
│   ├─▶ request_skill("clickup_create_project") ─────────────────────────┐   │
│   │       │                                                  Skill D   │   │
│   │       └─▶ request_agent("@project-agent", "create template")      │   │
│   │           ◀─── result: {template_id: "TPL-456"}                   │   │
│   │   ◀─── result: {project_id: "PRJ-789"} ────────────────────────────┘   │
│   │                                                                         │
│   └─▶ return_result({                                                       │
│           "onboarding_complete": true,                                      │
│           "deal_id": "D-123",                                              │
│           "invoice_id": "INV-123",                                         │
│           "project_id": "PRJ-789"                                          │
│       })                                                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Timeout and Error Handling

```python
# sorx_runtime/error_handling.py

class CallbackTimeout(Exception):
    """Raised when a callback doesn't receive response in time."""
    pass

class CallbackError(Exception):
    """Raised when a callback fails."""
    pass

# In skill code:
from sorx import request_skill, CallbackTimeout, CallbackError

def execute(params: dict):
    try:
        # Short timeout for critical path
        client = request_skill(
            skill="hubspot_get_client",
            params={"id": params["client_id"]},
            timeout_ms=5000  # 5 seconds max
        )
    except CallbackTimeout:
        # Fallback: use cached data or partial info
        client = get_cached_client(params["client_id"])
        if not client:
            return_result({
                "success": False,
                "error": "Could not fetch client data (timeout)",
                "retry_suggested": True
            })
            return
    except CallbackError as e:
        # The called skill failed
        return_result({
            "success": False,
            "error": f"Client lookup failed: {e.message}",
            "nested_error": e.details
        })
        return

    # Continue with client data...
```

### Orchestration Configuration

```yaml
# config/orchestration.yaml

orchestration:
  # WebSocket connection settings
  connection:
    url: "wss://businessos.app/sorx/orchestrate"
    reconnect_attempts: 5
    reconnect_delay_ms: 1000
    heartbeat_interval_ms: 30000

  # Callback defaults
  callbacks:
    default_timeout_ms: 30000
    max_timeout_ms: 300000      # 5 minutes absolute max
    max_chain_depth: 10         # Prevent infinite callback loops

  # Execution settings
  execution:
    max_concurrent_skills: 10
    max_waiting_callbacks: 20
    context_preservation: true   # Keep state while waiting

  # Agent routing
  agent_routing:
    default_agent: "@orchestrator"
    data_requests: "@data-agent"
    decisions: "@decision-agent"
    fallback: "@general-agent"

  # Priority handling
  priority:
    high:
      timeout_ms: 60000
      retry_attempts: 3
    normal:
      timeout_ms: 30000
      retry_attempts: 1
    low:
      timeout_ms: 120000
      retry_attempts: 0
```

---

## Interface Adapters

The key to universal connectivity is **Interface Adapters** - specialized handlers for different connection types:

### 1. REST API Adapter

```python
# adapters/rest_api.py

class RESTAdapter:
    """Handles REST API connections - the most common interface type."""

    capabilities = [
        "GET", "POST", "PUT", "PATCH", "DELETE",
        "OAuth2", "API Key", "Bearer Token", "Basic Auth",
        "JSON", "XML", "Form Data", "Multipart"
    ]

    def generate_skill(self, task: str, api_docs: str) -> Skill:
        """Generate a skill for a REST API endpoint."""
        # LLM generates Python requests code
        pass

    def execute(self, skill: Skill, params: dict) -> Result:
        """Execute a REST API skill."""
        pass
```

### 2. Database Adapter

```python
# adapters/database.py

class DatabaseAdapter:
    """Handles direct database connections."""

    capabilities = [
        "PostgreSQL", "MySQL", "SQLite", "MSSQL",
        "MongoDB", "Redis", "Elasticsearch",
        "Read", "Write", "Transactions"
    ]

    def generate_skill(self, task: str, schema: str) -> Skill:
        """Generate a skill for database operations."""
        # LLM generates SQL or query code
        pass
```

### 3. Legacy System Adapter

```python
# adapters/legacy.py

class LegacyAdapter:
    """Handles legacy system connections."""

    capabilities = [
        "SOAP", "XML-RPC", "EDI", "AS/400",
        "FTP", "SFTP", "Telnet", "SSH",
        "Mainframe", "COBOL Interfaces"
    ]

    def generate_skill(self, task: str, wsdl: str = None) -> Skill:
        """Generate a skill for legacy system operations."""
        # LLM generates SOAP calls, EDI messages, etc.
        pass
```

### 4. Desktop Automation Adapter

```python
# adapters/desktop.py

class DesktopAdapter:
    """Handles desktop application automation."""

    capabilities = {
        "macos": ["AppleScript", "JXA", "Automator"],
        "windows": ["PowerShell", "COM", "UI Automation"],
        "linux": ["X11", "DBus", "xdotool"]
    }

    def generate_skill(self, task: str, app: str) -> Skill:
        """Generate a skill to automate a desktop application."""
        # LLM generates AppleScript, PowerShell, etc.
        pass
```

### 5. File System Adapter

```python
# adapters/filesystem.py

class FileSystemAdapter:
    """Handles file system operations."""

    capabilities = [
        "Local Files", "Network Shares", "Cloud Storage",
        "FTP", "SFTP", "S3", "GCS", "Azure Blob",
        "Read", "Write", "Watch", "Transform"
    ]

    def generate_skill(self, task: str) -> Skill:
        """Generate a skill for file operations."""
        pass
```

### 6. Hardware/IoT Adapter

```python
# adapters/hardware.py

class HardwareAdapter:
    """Handles hardware and IoT connections."""

    capabilities = [
        "Serial", "USB", "Bluetooth", "Zigbee",
        "MQTT", "CoAP", "Modbus", "OPC-UA",
        "GPIO", "I2C", "SPI"
    ]

    def generate_skill(self, task: str, device_docs: str) -> Skill:
        """Generate a skill to interact with hardware."""
        pass
```

### 7. Communication Adapter

```python
# adapters/communication.py

class CommunicationAdapter:
    """Handles communication protocols."""

    capabilities = [
        "SMTP", "IMAP", "POP3",      # Email
        "SMS Gateways", "Twilio",     # SMS
        "SIP", "WebRTC",              # Voice
        "Fax APIs"                    # Fax
    ]

    def generate_skill(self, task: str) -> Skill:
        """Generate a skill for communication."""
        pass
```

---

## Skill Learning System

### Acquisition Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SKILL ACQUISITION FLOW                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   User Request: "Create a task in ClickUp when a deal closes in HubSpot"   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 1. SKILL CHECK                                                       │  │
│   │    Agent checks skill library:                                       │  │
│   │    - hubspot_watch_deal_close: NOT FOUND                            │  │
│   │    - clickup_create_task: FOUND (v2, 94% success)                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 2. SKILL ACQUISITION                                                 │  │
│   │    Need to learn: hubspot_watch_deal_close                          │  │
│   │                                                                       │  │
│   │    a) Load HubSpot API pattern template                             │  │
│   │    b) Identify interface type: REST API + Webhooks                  │  │
│   │    c) Generate skill script:                                         │  │
│   │       - Subscribe to deal.propertyChange webhook                    │  │
│   │       - Filter for dealstage = "closedwon"                         │  │
│   │       - Extract deal data for downstream use                        │  │
│   │    d) Validate script (security, efficiency)                        │  │
│   │    e) Save skill to library                                         │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 3. WORKFLOW COMPOSITION                                              │  │
│   │    Combine skills into workflow:                                     │  │
│   │                                                                       │  │
│   │    [hubspot_watch_deal_close] ──▶ [clickup_create_task]            │  │
│   │                                                                       │  │
│   │    Workflow saved as compound skill:                                │  │
│   │    "hubspot_deal_to_clickup_task"                                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 4. EXECUTION                                                         │  │
│   │    - Deploy webhook listener                                        │  │
│   │    - Monitor for deal closes                                        │  │
│   │    - Create ClickUp tasks automatically                             │  │
│   │    - Track success/failure                                          │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Evolution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SKILL EVOLUTION FLOW                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Feedback: "The ClickUp task didn't include the deal amount"              │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 1. ANALYZE FAILURE                                                   │  │
│   │    - Load skill: clickup_create_task v2                             │  │
│   │    - Identify gap: Missing field mapping for deal.amount            │  │
│   │    - Root cause: Skill wasn't configured to pass financial data     │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 2. IMPROVE SKILL                                                     │  │
│   │    Agent generates improved version:                                │  │
│   │                                                                       │  │
│   │    # v2: Original                                                    │  │
│   │    task_data = {                                                     │  │
│   │        "name": deal["dealname"],                                    │  │
│   │        "description": deal["description"]                           │  │
│   │    }                                                                 │  │
│   │                                                                       │  │
│   │    # v3: Improved                                                    │  │
│   │    task_data = {                                                     │  │
│   │        "name": deal["dealname"],                                    │  │
│   │        "description": deal["description"],                          │  │
│   │        "custom_fields": [                                           │  │
│   │            {"name": "Deal Amount", "value": deal["amount"]},       │  │
│   │            {"name": "Close Date", "value": deal["closedate"]}      │  │
│   │        ]                                                             │  │
│   │    }                                                                 │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 3. VERSION & SAVE                                                    │  │
│   │    - Increment version: v2 → v3                                     │  │
│   │    - Save improvement notes                                         │  │
│   │    - Archive v2 (don't delete - may need rollback)                  │  │
│   │    - Update success tracking                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 4. LEARN & GENERALIZE                                                │  │
│   │    Agent notes pattern for future:                                  │  │
│   │    "When creating tasks from deals, always include financial data"  │  │
│   │                                                                       │  │
│   │    This learning is applied to:                                     │  │
│   │    - Similar skills (asana_create_task, linear_create_issue)        │  │
│   │    - Future skill generation                                         │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Pattern Library

Patterns are minimal templates that teach agents HOW to connect to different interface types. They're NOT full implementations - just enough for the LLM to generate correct skills.

### Pattern Categories

```
patterns/
├── interfaces/                      # Interface type patterns
│   ├── rest_api.py                  # REST API pattern
│   ├── graphql.py                   # GraphQL pattern
│   ├── soap.py                      # SOAP/XML-RPC pattern
│   ├── grpc.py                      # gRPC pattern
│   ├── websocket.py                 # WebSocket pattern
│   ├── database.py                  # Database pattern
│   ├── file_system.py               # File operations pattern
│   ├── desktop_macos.py             # macOS automation pattern
│   ├── desktop_windows.py           # Windows automation pattern
│   ├── desktop_linux.py             # Linux automation pattern
│   └── hardware.py                  # Hardware/IoT pattern
│
├── auth/                            # Authentication patterns
│   ├── oauth2.py                    # OAuth 2.0 flows
│   ├── api_key.py                   # API key auth
│   ├── basic_auth.py                # Basic auth
│   ├── jwt.py                       # JWT tokens
│   ├── saml.py                      # SAML (enterprise SSO)
│   └── certificate.py               # Certificate-based auth
│
├── providers/                       # Provider-specific patterns
│   ├── google/
│   │   ├── gmail.py
│   │   ├── drive.py
│   │   ├── calendar.py
│   │   └── sheets.py
│   ├── microsoft/
│   │   ├── outlook.py
│   │   ├── teams.py
│   │   └── onedrive.py
│   ├── salesforce/
│   │   └── salesforce.py
│   ├── hubspot/
│   │   └── hubspot.py
│   └── ... (all planned integrations)
│
└── enterprise/                      # Enterprise system patterns
    ├── sap.py                       # SAP integration
    ├── oracle.py                    # Oracle integration
    ├── as400.py                     # AS/400 mainframe
    ├── edi.py                       # EDI messaging
    └── ldap.py                      # LDAP/Active Directory
```

### Pattern Example: REST API

```python
# patterns/interfaces/rest_api.py

REST_API_PATTERN = """
## REST API Skill Pattern

### Structure:
```python
import requests
from sorx import get_credential, return_result, log

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]

    Args:
        params: {
            [PARAM_DEFINITIONS]
        }

    Returns:
        Structured result via return_result()
    '''

    # 1. Get credentials from local vault
    creds = get_credential("[PROVIDER]")

    # 2. Build request
    url = f"{creds['base_url']}/[ENDPOINT]"
    headers = {
        "Authorization": f"Bearer {creds['access_token']}",
        "Content-Type": "application/json"
    }

    # 3. Make request
    response = requests.[METHOD](
        url,
        headers=headers,
        json=params.get("body"),      # For POST/PUT
        params=params.get("query")     # For GET
    )

    # 4. Handle token expiration
    if response.status_code == 401:
        # Trigger token refresh flow
        return_result({"error": "token_expired", "refresh": True})
        return

    # 5. Handle rate limiting
    if response.status_code == 429:
        retry_after = response.headers.get("Retry-After", 60)
        return_result({"error": "rate_limited", "retry_after": retry_after})
        return

    # 6. Return structured result
    if response.ok:
        return_result({
            "success": True,
            "data": response.json()
        })
    else:
        return_result({
            "success": False,
            "error": response.text,
            "status_code": response.status_code
        })
```

### Common Patterns:
- List: GET /resources → returns array
- Get: GET /resources/{id} → returns object
- Create: POST /resources → returns created object
- Update: PUT/PATCH /resources/{id} → returns updated object
- Delete: DELETE /resources/{id} → returns success/empty
- Search: GET /resources?query=... → returns filtered array

### Pagination:
- Cursor-based: ?cursor=abc123
- Offset-based: ?offset=100&limit=50
- Page-based: ?page=2&per_page=50

### Error Handling:
- 400: Bad request (log params, fix request)
- 401: Unauthorized (refresh token)
- 403: Forbidden (check permissions)
- 404: Not found (verify resource exists)
- 429: Rate limited (backoff and retry)
- 500: Server error (retry with backoff)
"""
```

### Pattern Example: Legacy SOAP

```python
# patterns/interfaces/soap.py

SOAP_PATTERN = """
## SOAP/XML-RPC Skill Pattern

For connecting to legacy enterprise systems using SOAP.

### Structure:
```python
from zeep import Client
from zeep.wsse.username import UsernameToken
from sorx import get_credential, return_result

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]
    '''

    # 1. Get credentials
    creds = get_credential("[PROVIDER]")

    # 2. Create SOAP client with WSDL
    wsse = UsernameToken(creds['username'], creds['password'])
    client = Client(
        wsdl=creds['wsdl_url'],
        wsse=wsse
    )

    # 3. Call SOAP method
    try:
        result = client.service.[METHOD_NAME](
            [PARAMETERS]
        )

        return_result({
            "success": True,
            "data": serialize_zeep_result(result)
        })

    except Exception as e:
        return_result({
            "success": False,
            "error": str(e)
        })
```

### Notes:
- Always use zeep library for SOAP
- Get WSDL URL from provider
- Handle complex types properly
- Serialize results to JSON
"""
```

### Pattern Example: Desktop Automation (macOS)

```python
# patterns/interfaces/desktop_macos.py

MACOS_PATTERN = """
## macOS Desktop Automation Pattern

For automating macOS applications using AppleScript/JXA.

### Structure:
```python
import subprocess
from sorx import return_result

def execute(params: dict):
    '''
    [SKILL_DESCRIPTION]
    '''

    # AppleScript to execute
    script = '''
    tell application "[APP_NAME]"
        [COMMANDS]
    end tell
    '''

    # Execute AppleScript
    result = subprocess.run(
        ['osascript', '-e', script],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return_result({
            "success": True,
            "output": result.stdout
        })
    else:
        return_result({
            "success": False,
            "error": result.stderr
        })
```

### Common Apps:
- Finder: File operations
- Mail: Email (alternative to API)
- Calendar: Local calendar
- Notes: Local notes
- Numbers/Pages/Keynote: Office suite
- Any app with AppleScript support

### JXA Alternative:
```javascript
// For JavaScript-based automation
const app = Application('[APP_NAME]');
app.includeStandardAdditions = true;
// ... commands
```
"""
```

---

## Role-Based Skills

Different agent roles have affinities for different skill categories:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          ROLE-BASED SKILL MATRIX                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   SALES AGENT                                                               │
│   ───────────                                                               │
│   Primary Skills:                                                           │
│   - CRM operations (HubSpot, Salesforce, Pipedrive)                        │
│   - Email outreach (Gmail, Outlook)                                        │
│   - Calendar management                                                     │
│   - Document generation (proposals, quotes)                                │
│                                                                             │
│   SUPPORT AGENT                                                             │
│   ─────────────                                                             │
│   Primary Skills:                                                           │
│   - Ticketing systems (Zendesk, Intercom, Freshdesk)                       │
│   - Knowledge base operations                                               │
│   - Customer lookup (CRM)                                                   │
│   - Communication (email, chat, SMS)                                       │
│                                                                             │
│   MARKETING AGENT                                                           │
│   ───────────────                                                           │
│   Primary Skills:                                                           │
│   - Email campaigns (Mailchimp, Klaviyo)                                   │
│   - Social media (Buffer, Hootsuite)                                       │
│   - Analytics (Google Analytics, Mixpanel)                                 │
│   - Content management (WordPress, Webflow)                                │
│                                                                             │
│   OPERATIONS AGENT                                                          │
│   ────────────────                                                          │
│   Primary Skills:                                                           │
│   - Task management (ClickUp, Asana, Monday)                               │
│   - Documentation (Notion, Confluence)                                      │
│   - File management (Drive, Dropbox)                                       │
│   - Team communication (Slack, Teams)                                      │
│                                                                             │
│   FINANCE AGENT                                                             │
│   ─────────────                                                             │
│   Primary Skills:                                                           │
│   - Accounting (QuickBooks, Xero)                                          │
│   - Payments (Stripe, PayPal)                                              │
│   - Invoicing                                                               │
│   - Reporting                                                               │
│                                                                             │
│   DEVELOPER AGENT                                                           │
│   ───────────────                                                           │
│   Primary Skills:                                                           │
│   - Code repositories (GitHub, GitLab)                                     │
│   - Issue tracking (Jira, Linear)                                          │
│   - CI/CD pipelines                                                         │
│   - Monitoring (Datadog, Sentry)                                           │
│                                                                             │
│   IT ADMIN AGENT                                                            │
│   ──────────────                                                            │
│   Primary Skills:                                                           │
│   - Directory services (LDAP, Azure AD)                                    │
│   - Device management                                                       │
│   - Security tools                                                          │
│   - Legacy system access                                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## BusinessOS Skill Catalog

Concrete skill examples mapped to BusinessOS integrations. These skills grow organically as agents learn to perform business workflows.

### Skill File Structure

```
~/.businessos/skills/
├── crm/                                    # CRM & Sales
│   ├── hubspot-deal-pipeline.md            ~3.2k tokens
│   ├── hubspot-contact-enrichment.md       ~2.8k tokens
│   ├── hubspot-company-lookup.md           ~1.9k tokens
│   ├── lead-qualification-workflow.md      ~4.1k tokens
│   ├── client-health-score.md              ~3.5k tokens
│   └── sales-handoff-automation.md         ~2.7k tokens
│
├── communication/                          # Email & Messaging
│   ├── gmail-client-outreach.md            ~2.4k tokens
│   ├── gmail-followup-sequence.md          ~3.8k tokens
│   ├── gmail-meeting-request.md            ~1.7k tokens
│   ├── slack-team-notification.md          ~1.5k tokens
│   ├── slack-client-channel-update.md      ~2.1k tokens
│   ├── slack-standup-collection.md         ~2.9k tokens
│   └── multi-channel-announcement.md       ~3.3k tokens
│
├── tasks/                                  # Task & Project Management
│   ├── clickup-task-from-email.md          ~2.6k tokens
│   ├── clickup-project-setup.md            ~4.2k tokens
│   ├── clickup-sprint-planning.md          ~3.7k tokens
│   ├── asana-milestone-tracking.md         ~2.8k tokens
│   ├── task-delegation-workflow.md         ~3.1k tokens
│   └── weekly-task-summary.md              ~2.3k tokens
│
├── documents/                              # Documents & Knowledge
│   ├── notion-meeting-notes.md             ~2.5k tokens
│   ├── notion-project-wiki.md              ~3.4k tokens
│   ├── notion-client-database.md           ~2.9k tokens
│   ├── drive-proposal-generation.md        ~4.5k tokens
│   ├── drive-contract-organization.md      ~2.2k tokens
│   └── knowledge-base-update.md            ~2.7k tokens
│
├── meetings/                               # Calendar & Meetings
│   ├── calendar-smart-scheduling.md        ~3.1k tokens
│   ├── calendar-availability-check.md      ~1.8k tokens
│   ├── zoom-meeting-setup.md               ~2.3k tokens
│   ├── meeting-prep-workflow.md            ~3.6k tokens
│   ├── meeting-followup-automation.md      ~4.0k tokens
│   └── recurring-meeting-management.md     ~2.4k tokens
│
├── finance/                                # Finance & Billing
│   ├── stripe-invoice-generation.md        ~2.8k tokens
│   ├── stripe-payment-tracking.md          ~2.1k tokens
│   ├── overdue-payment-reminder.md         ~3.2k tokens
│   ├── revenue-report-generation.md        ~3.9k tokens
│   └── expense-categorization.md           ~2.5k tokens
│
├── workflows/                              # Compound Workflows
│   ├── new-client-onboarding.md            ~5.8k tokens
│   ├── deal-won-automation.md              ~4.7k tokens
│   ├── project-kickoff-sequence.md         ~5.2k tokens
│   ├── weekly-report-distribution.md       ~3.8k tokens
│   ├── client-renewal-workflow.md          ~4.4k tokens
│   └── escalation-handling.md              ~3.6k tokens
│
└── meta/                                   # Skills about Skills
    ├── skill-creator.md                    ~2.1k tokens
    ├── workflow-analyzer.md                ~2.8k tokens
    └── skill-optimizer.md                  ~2.4k tokens
```

---

### Skill Examples: CRM (HubSpot)

#### hubspot-deal-pipeline.md

```markdown
# HubSpot Deal Pipeline Management

## Skill Metadata
- **ID**: hubspot-deal-pipeline-v4
- **Category**: crm
- **Provider**: hubspot
- **Role Affinity**: sales, account-management
- **Credentials**: hubspot
- **Version**: 4
- **Executions**: 847
- **Success Rate**: 98.7%

## Description
Manages the full deal lifecycle in HubSpot - creating deals, updating stages,
tracking activities, and triggering downstream workflows.

## Capabilities
- Create new deals with proper associations
- Move deals through pipeline stages
- Add notes and activities to deals
- Track deal value and probability
- Trigger notifications on stage changes

## Script

```python
import requests
from sorx import get_credential, return_result, trigger_skill

def execute(params: dict):
    """
    HubSpot Deal Pipeline Management

    Actions:
        - create: Create new deal
        - update_stage: Move deal to new stage
        - add_activity: Add note/call/email to deal
        - get_pipeline: Get pipeline summary
    """
    creds = get_credential("hubspot")
    base_url = "https://api.hubapi.com"
    headers = {
        "Authorization": f"Bearer {creds['access_token']}",
        "Content-Type": "application/json"
    }

    action = params.get("action")

    if action == "create":
        deal_data = {
            "properties": {
                "dealname": params["name"],
                "amount": params.get("amount", 0),
                "dealstage": params.get("stage", "appointmentscheduled"),
                "pipeline": params.get("pipeline", "default"),
                "closedate": params.get("close_date"),
                "hubspot_owner_id": params.get("owner_id")
            }
        }

        # Associate with contact and company if provided
        if params.get("contact_id") or params.get("company_id"):
            deal_data["associations"] = []
            if params.get("contact_id"):
                deal_data["associations"].append({
                    "to": {"id": params["contact_id"]},
                    "types": [{"associationCategory": "HUBSPOT_DEFINED", "associationTypeId": 3}]
                })
            if params.get("company_id"):
                deal_data["associations"].append({
                    "to": {"id": params["company_id"]},
                    "types": [{"associationCategory": "HUBSPOT_DEFINED", "associationTypeId": 5}]
                })

        response = requests.post(
            f"{base_url}/crm/v3/objects/deals",
            headers=headers,
            json=deal_data
        )

        if response.ok:
            deal = response.json()
            # Trigger notification skill
            trigger_skill("slack-team-notification", {
                "channel": "#sales",
                "message": f"New deal created: {params['name']} (${params.get('amount', 0):,})"
            })
            return_result({"success": True, "deal_id": deal["id"], "deal": deal})
        else:
            return_result({"success": False, "error": response.text})

    elif action == "update_stage":
        deal_id = params["deal_id"]
        new_stage = params["stage"]

        # Get current deal first
        current = requests.get(
            f"{base_url}/crm/v3/objects/deals/{deal_id}",
            headers=headers
        ).json()

        old_stage = current["properties"].get("dealstage")

        response = requests.patch(
            f"{base_url}/crm/v3/objects/deals/{deal_id}",
            headers=headers,
            json={"properties": {"dealstage": new_stage}}
        )

        if response.ok:
            # Check if deal was won
            if new_stage == "closedwon":
                trigger_skill("deal-won-automation", {
                    "deal_id": deal_id,
                    "deal_name": current["properties"]["dealname"],
                    "amount": current["properties"].get("amount")
                })

            return_result({
                "success": True,
                "deal_id": deal_id,
                "old_stage": old_stage,
                "new_stage": new_stage
            })
        else:
            return_result({"success": False, "error": response.text})

    elif action == "add_activity":
        # Add engagement (note, call, email, meeting)
        engagement_data = {
            "engagement": {
                "type": params.get("type", "NOTE"),
                "timestamp": params.get("timestamp", int(time.time() * 1000))
            },
            "associations": {
                "dealIds": [params["deal_id"]]
            },
            "metadata": {
                "body": params["content"]
            }
        }

        response = requests.post(
            f"{base_url}/engagements/v1/engagements",
            headers=headers,
            json=engagement_data
        )

        return_result({"success": response.ok, "engagement": response.json() if response.ok else None})

    elif action == "get_pipeline":
        # Get all deals in pipeline with summary
        response = requests.post(
            f"{base_url}/crm/v3/objects/deals/search",
            headers=headers,
            json={
                "filterGroups": [{
                    "filters": [{
                        "propertyName": "pipeline",
                        "operator": "EQ",
                        "value": params.get("pipeline", "default")
                    }]
                }],
                "properties": ["dealname", "amount", "dealstage", "closedate"],
                "limit": 100
            }
        )

        if response.ok:
            deals = response.json()["results"]
            # Group by stage
            by_stage = {}
            total_value = 0
            for deal in deals:
                stage = deal["properties"]["dealstage"]
                if stage not in by_stage:
                    by_stage[stage] = {"count": 0, "value": 0}
                by_stage[stage]["count"] += 1
                amount = float(deal["properties"].get("amount") or 0)
                by_stage[stage]["value"] += amount
                total_value += amount

            return_result({
                "success": True,
                "total_deals": len(deals),
                "total_value": total_value,
                "by_stage": by_stage,
                "deals": deals
            })
        else:
            return_result({"success": False, "error": response.text})
```

## Evolution History
- **v1**: Basic deal creation
- **v2**: Added stage updates and associations
- **v3**: Added activity tracking, fixed amount formatting
- **v4**: Added pipeline summary, integrated with notification skills

## Learned Patterns
- Always associate deals with contacts AND companies when available
- Include dollar amounts in notifications for context
- Trigger downstream automations on stage changes (especially closedwon)
```

---

### Skill Examples: Communication (Gmail + Slack)

#### gmail-followup-sequence.md

```markdown
# Gmail Follow-up Sequence

## Skill Metadata
- **ID**: gmail-followup-sequence-v3
- **Category**: communication
- **Provider**: google
- **Role Affinity**: sales, account-management, support
- **Credentials**: google
- **Version**: 3
- **Executions**: 1,247
- **Success Rate**: 97.2%

## Description
Manages intelligent email follow-up sequences. Tracks sent emails, schedules
follow-ups, personalizes based on context, and stops when recipient responds.

## Capabilities
- Create follow-up sequences (3-5 touches)
- Personalize emails based on recipient data
- Track opens and responses
- Auto-stop when recipient replies
- Variable timing between touches

## Script

```python
import requests
import base64
from email.mime.text import MIMEText
from datetime import datetime, timedelta
from sorx import get_credential, return_result, schedule_skill, get_context

def execute(params: dict):
    """
    Gmail Follow-up Sequence Management

    Actions:
        - start_sequence: Begin a new follow-up sequence
        - send_followup: Send the next follow-up in sequence
        - check_response: Check if recipient responded
        - stop_sequence: End sequence early
    """
    creds = get_credential("google")
    headers = {"Authorization": f"Bearer {creds['access_token']}"}

    action = params.get("action")

    if action == "start_sequence":
        recipient = params["to"]
        subject = params["subject"]
        sequence_id = f"seq_{recipient}_{int(time.time())}"

        # Get recipient context from CRM if available
        context = get_context("client", recipient)

        # Personalize first email
        body = personalize_email(
            template=params["template"],
            recipient_name=context.get("name", "there"),
            company=context.get("company"),
            custom_fields=params.get("custom_fields", {})
        )

        # Send first email
        message = create_email(recipient, subject, body)
        sent = send_email(headers, message)

        if sent["success"]:
            # Schedule follow-ups
            schedule_followups(sequence_id, params)

            return_result({
                "success": True,
                "sequence_id": sequence_id,
                "first_email_id": sent["message_id"],
                "followups_scheduled": len(params.get("followup_templates", []))
            })
        else:
            return_result({"success": False, "error": sent["error"]})

    elif action == "send_followup":
        sequence_id = params["sequence_id"]
        followup_number = params["followup_number"]

        # Check if recipient already responded
        if check_for_response(headers, params["thread_id"]):
            # Stop sequence - they replied!
            return_result({
                "success": True,
                "action": "sequence_stopped",
                "reason": "recipient_replied"
            })

        # Get the right template
        template = params["templates"][followup_number - 1]

        # Personalize
        context = get_context("client", params["to"])
        body = personalize_email(
            template=template,
            recipient_name=context.get("name", "there"),
            company=context.get("company"),
            followup_number=followup_number
        )

        # Send as reply to thread
        message = create_reply(params["thread_id"], body)
        sent = send_email(headers, message)

        if sent["success"]:
            # Schedule next if more remain
            if followup_number < len(params["templates"]):
                schedule_skill("gmail-followup-sequence", {
                    "action": "send_followup",
                    "sequence_id": sequence_id,
                    "followup_number": followup_number + 1,
                    "thread_id": params["thread_id"],
                    "to": params["to"],
                    "templates": params["templates"]
                }, delay_days=params.get("days_between", 3))

            return_result({
                "success": True,
                "followup_sent": followup_number,
                "remaining": len(params["templates"]) - followup_number
            })
        else:
            return_result({"success": False, "error": sent["error"]})

    elif action == "check_response":
        thread_id = params["thread_id"]
        response = requests.get(
            f"https://gmail.googleapis.com/gmail/v1/users/me/threads/{thread_id}",
            headers=headers
        )

        if response.ok:
            thread = response.json()
            messages = thread.get("messages", [])

            # Check if any message is FROM the recipient (not us)
            our_email = creds.get("email")
            for msg in messages:
                headers_list = msg.get("payload", {}).get("headers", [])
                from_header = next((h["value"] for h in headers_list if h["name"] == "From"), "")
                if our_email not in from_header:
                    return_result({
                        "success": True,
                        "has_response": True,
                        "response_snippet": msg.get("snippet")
                    })

            return_result({"success": True, "has_response": False})
        else:
            return_result({"success": False, "error": response.text})


def personalize_email(template, recipient_name, company=None, **kwargs):
    """Replace placeholders with actual values."""
    result = template
    result = result.replace("{{name}}", recipient_name)
    result = result.replace("{{company}}", company or "your company")
    result = result.replace("{{followup_number}}", str(kwargs.get("followup_number", 1)))

    for key, value in kwargs.get("custom_fields", {}).items():
        result = result.replace(f"{{{{{key}}}}}", str(value))

    return result


def create_email(to, subject, body):
    """Create a MIME email message."""
    message = MIMEText(body)
    message["to"] = to
    message["subject"] = subject
    return {"raw": base64.urlsafe_b64encode(message.as_bytes()).decode()}


def send_email(headers, message):
    """Send email via Gmail API."""
    response = requests.post(
        "https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
        headers={**headers, "Content-Type": "application/json"},
        json=message
    )
    if response.ok:
        return {"success": True, "message_id": response.json()["id"]}
    return {"success": False, "error": response.text}
```

## Sequence Templates

### Sales Follow-up (3 touches)
```
Touch 1: Initial outreach with value prop
Touch 2 (+3 days): "Wanted to make sure you saw my email..."
Touch 3 (+5 days): "One last follow-up..." with different angle
```

### Meeting Request (2 touches)
```
Touch 1: Meeting request with proposed times
Touch 2 (+2 days): "Still hoping to connect..."
```

## Evolution History
- **v1**: Basic single follow-up email
- **v2**: Added sequence support with scheduling
- **v3**: Added response detection, personalization, CRM integration

## Learned Patterns
- Always check for response before sending follow-up
- 3-day spacing works well for sales, 2-day for urgent
- Include "Re:" in subject to improve open rates
- Stop immediately when recipient replies (don't annoy)
```

---

#### slack-team-notification.md

```markdown
# Slack Team Notification

## Skill Metadata
- **ID**: slack-team-notification-v5
- **Category**: communication
- **Provider**: slack
- **Role Affinity**: all
- **Credentials**: slack
- **Version**: 5
- **Executions**: 3,892
- **Success Rate**: 99.4%

## Description
Sends formatted notifications to Slack channels with context-aware formatting,
thread support, and interactive elements.

## Script

```python
import requests
from sorx import get_credential, return_result

def execute(params: dict):
    """
    Slack Team Notification

    Supports:
        - Simple messages
        - Rich formatted blocks
        - Thread replies
        - Mentions (@user, @channel)
        - Attachments with color coding
    """
    creds = get_credential("slack")
    headers = {
        "Authorization": f"Bearer {creds['bot_token']}",
        "Content-Type": "application/json"
    }

    channel = params["channel"]
    message = params.get("message", "")
    notification_type = params.get("type", "info")

    # Build message payload
    payload = {"channel": channel}

    # If simple message, just send text
    if params.get("simple"):
        payload["text"] = message
    else:
        # Build rich blocks
        blocks = []

        # Add header if provided
        if params.get("title"):
            blocks.append({
                "type": "header",
                "text": {"type": "plain_text", "text": params["title"]}
            })

        # Add main message
        if message:
            blocks.append({
                "type": "section",
                "text": {"type": "mrkdwn", "text": message}
            })

        # Add fields if provided (key-value pairs)
        if params.get("fields"):
            fields_block = {
                "type": "section",
                "fields": [
                    {"type": "mrkdwn", "text": f"*{k}:*\n{v}"}
                    for k, v in params["fields"].items()
                ]
            }
            blocks.append(fields_block)

        # Add action buttons if provided
        if params.get("actions"):
            actions_block = {
                "type": "actions",
                "elements": [
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": action["text"]},
                        "url": action.get("url"),
                        "action_id": action.get("action_id", f"action_{i}")
                    }
                    for i, action in enumerate(params["actions"])
                ]
            }
            blocks.append(actions_block)

        # Add context footer
        if params.get("footer"):
            blocks.append({
                "type": "context",
                "elements": [{"type": "mrkdwn", "text": params["footer"]}]
            })

        payload["blocks"] = blocks

        # Add color-coded attachment based on type
        colors = {
            "success": "#36a64f",
            "warning": "#ffcc00",
            "error": "#ff0000",
            "info": "#0066ff"
        }
        if notification_type in colors:
            payload["attachments"] = [{
                "color": colors[notification_type],
                "blocks": blocks
            }]
            payload.pop("blocks")  # Move blocks into attachment

    # Thread reply
    if params.get("thread_ts"):
        payload["thread_ts"] = params["thread_ts"]

    # Send
    response = requests.post(
        "https://slack.com/api/chat.postMessage",
        headers=headers,
        json=payload
    )

    if response.ok and response.json().get("ok"):
        result = response.json()
        return_result({
            "success": True,
            "ts": result["ts"],
            "channel": result["channel"],
            "thread_ts": result.get("message", {}).get("thread_ts")
        })
    else:
        return_result({
            "success": False,
            "error": response.json().get("error", response.text)
        })
```

## Usage Examples

### Deal Won Notification
```python
trigger_skill("slack-team-notification", {
    "channel": "#sales-wins",
    "type": "success",
    "title": "Deal Won!",
    "message": "Congratulations! We just closed *Acme Corp*",
    "fields": {
        "Deal Value": "$45,000",
        "Sales Rep": "@john",
        "Close Date": "Jan 4, 2026"
    },
    "actions": [
        {"text": "View in HubSpot", "url": "https://app.hubspot.com/deals/123"}
    ],
    "footer": "Via BusinessOS Automation"
})
```

### Error Alert
```python
trigger_skill("slack-team-notification", {
    "channel": "#alerts",
    "type": "error",
    "title": "Integration Error",
    "message": "Failed to sync contacts from HubSpot",
    "fields": {
        "Error": "Rate limit exceeded",
        "Retry In": "15 minutes"
    }
})
```

## Evolution History
- **v1**: Simple text messages
- **v2**: Added blocks formatting
- **v3**: Added attachments with colors
- **v4**: Added thread support, actions
- **v5**: Added fields layout, footer, type-based coloring
```

---

### Skill Examples: Tasks (ClickUp)

#### clickup-task-from-email.md

```markdown
# ClickUp Task from Email

## Skill Metadata
- **ID**: clickup-task-from-email-v4
- **Category**: tasks
- **Provider**: clickup
- **Role Affinity**: operations, project-management, support
- **Credentials**: clickup
- **Version**: 4
- **Executions**: 562
- **Success Rate**: 96.8%

## Description
Intelligently converts emails into ClickUp tasks with proper categorization,
priority detection, assignee matching, and deadline extraction.

## Script

```python
import requests
import re
from datetime import datetime, timedelta
from sorx import get_credential, return_result, call_llm

def execute(params: dict):
    """
    Create ClickUp task from email content.

    Automatically extracts:
        - Task name from subject
        - Description from body
        - Priority from urgency keywords
        - Due date from mentioned dates
        - Assignee from @mentions or context
    """
    creds = get_credential("clickup")
    headers = {"Authorization": creds["api_key"]}

    email = params["email"]
    list_id = params.get("list_id") or creds.get("default_list_id")

    # Extract task details using LLM
    extraction = call_llm(
        prompt=f"""Extract task details from this email:

Subject: {email['subject']}
From: {email['from']}
Body: {email['body']}

Return JSON with:
- task_name: Clear, actionable task title
- description: Key details and context
- priority: 1 (urgent), 2 (high), 3 (normal), 4 (low)
- due_date: ISO date if mentioned, null otherwise
- tags: relevant tags (max 3)
""",
        response_format="json"
    )

    # Build task payload
    task_data = {
        "name": extraction["task_name"],
        "description": f"""**From Email**
From: {email['from']}
Subject: {email['subject']}
Date: {email['date']}

---

{extraction['description']}

---
*Original email:*
{email['body'][:2000]}
""",
        "priority": extraction["priority"],
        "tags": extraction.get("tags", []),
        "custom_fields": []
    }

    # Add due date if extracted
    if extraction.get("due_date"):
        task_data["due_date"] = int(datetime.fromisoformat(
            extraction["due_date"]
        ).timestamp() * 1000)

    # Add source tracking
    task_data["custom_fields"].append({
        "id": creds.get("source_field_id"),
        "value": "email"
    })

    # Link to original email if we have message ID
    if email.get("message_id"):
        task_data["custom_fields"].append({
            "id": creds.get("email_link_field_id"),
            "value": f"gmail://message/{email['message_id']}"
        })

    # Try to match assignee
    assignee = match_assignee(email, creds.get("team_members", []))
    if assignee:
        task_data["assignees"] = [assignee["id"]]

    # Create task
    response = requests.post(
        f"https://api.clickup.com/api/v2/list/{list_id}/task",
        headers=headers,
        json=task_data
    )

    if response.ok:
        task = response.json()

        # Notify assignee
        if assignee:
            trigger_skill("slack-team-notification", {
                "channel": f"@{assignee['username']}",
                "type": "info",
                "message": f"New task assigned from email: *{extraction['task_name']}*",
                "actions": [{"text": "View Task", "url": task["url"]}]
            })

        return_result({
            "success": True,
            "task_id": task["id"],
            "task_url": task["url"],
            "task_name": extraction["task_name"],
            "priority": extraction["priority"],
            "assignee": assignee["name"] if assignee else None
        })
    else:
        return_result({"success": False, "error": response.text})


def match_assignee(email, team_members):
    """Match email sender or @mentions to team member."""
    from_email = email["from"].lower()

    # Check if sender is a team member
    for member in team_members:
        if member["email"].lower() in from_email:
            return member

    # Check for @mentions in body
    mentions = re.findall(r'@(\w+)', email["body"])
    for mention in mentions:
        for member in team_members:
            if mention.lower() in member["username"].lower():
                return member

    return None
```

## Example Transformation

**Input Email:**
```
From: john@client.com
Subject: URGENT: Need proposal updates by Friday
Body: Hi team, we need the updated proposal with the new pricing
by end of day Friday. @sarah can you handle this?
```

**Output Task:**
```
Name: Update proposal with new pricing for John
Priority: 1 (Urgent)
Due: Friday EOD
Assignee: Sarah
Tags: [proposal, client-request]
Description: Client needs updated proposal with new pricing...
```

## Evolution History
- **v1**: Basic email-to-task with manual fields
- **v2**: Added LLM extraction for smart parsing
- **v3**: Added assignee matching, priority detection
- **v4**: Added email linking, source tracking, notifications

## Learned Patterns
- Include original email snippet in description for context
- "URGENT" in subject = priority 1
- Dates like "by Friday" should be parsed to actual dates
- Always link back to source email for reference
```

---

### Skill Examples: Compound Workflows

#### deal-won-automation.md

```markdown
# Deal Won Automation

## Skill Metadata
- **ID**: deal-won-automation-v3
- **Category**: workflows
- **Provider**: multi (hubspot, clickup, slack, gmail, notion)
- **Role Affinity**: sales, operations, account-management
- **Credentials**: hubspot, clickup, slack, google, notion
- **Version**: 3
- **Executions**: 234
- **Success Rate**: 98.3%

## Description
Comprehensive automation triggered when a deal is marked as won in HubSpot.
Orchestrates multiple downstream actions across systems.

## Workflow Diagram

```
Deal Won in HubSpot
        │
        ├──► Slack: Celebrate in #sales-wins
        │
        ├──► ClickUp: Create onboarding project
        │       └──► With checklist tasks
        │
        ├──► Gmail: Send welcome email to client
        │
        ├──► Notion: Create client wiki page
        │
        ├──► HubSpot: Update deal properties
        │       ├──► Set closed date
        │       └──► Move contact to "Customer" lifecycle
        │
        └──► Calendar: Schedule kickoff meeting
```

## Script

```python
import asyncio
from sorx import get_credential, return_result, trigger_skill, call_llm
from datetime import datetime, timedelta

def execute(params: dict):
    """
    Deal Won Automation - Full post-sale workflow

    Triggered when: HubSpot deal stage = closedwon
    """
    deal_id = params["deal_id"]
    deal_name = params["deal_name"]
    amount = params.get("amount", 0)

    # Get full deal details from HubSpot
    deal = get_deal_details(deal_id)
    contact = get_associated_contact(deal_id)
    company = get_associated_company(deal_id)

    results = {
        "deal_id": deal_id,
        "deal_name": deal_name,
        "actions_completed": []
    }

    # ═══════════════════════════════════════════════════════════════
    # 1. CELEBRATE - Notify the team
    # ═══════════════════════════════════════════════════════════════

    slack_result = trigger_skill("slack-team-notification", {
        "channel": "#sales-wins",
        "type": "success",
        "title": f"Deal Won: {deal_name}",
        "message": f"*{deal.get('owner_name', 'Team')}* just closed a deal!",
        "fields": {
            "Client": company.get("name", "Unknown"),
            "Value": f"${amount:,.2f}",
            "Sales Cycle": f"{deal.get('days_to_close', '?')} days"
        },
        "actions": [
            {"text": "View Deal", "url": f"https://app.hubspot.com/deals/{deal_id}"}
        ]
    })
    results["actions_completed"].append({"action": "slack_notification", "success": slack_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 2. CREATE PROJECT - Set up onboarding in ClickUp
    # ═══════════════════════════════════════════════════════════════

    project_result = trigger_skill("clickup-project-setup", {
        "template": "client-onboarding",
        "name": f"Onboarding: {company.get('name', deal_name)}",
        "custom_fields": {
            "client_name": company.get("name"),
            "deal_value": amount,
            "hubspot_deal_id": deal_id,
            "primary_contact": contact.get("email")
        },
        "tasks": [
            {"name": "Send welcome packet", "assignee": "onboarding-team", "due_days": 1},
            {"name": "Schedule kickoff call", "assignee": "account-manager", "due_days": 2},
            {"name": "Set up client workspace", "assignee": "ops-team", "due_days": 3},
            {"name": "Create project plan", "assignee": "project-manager", "due_days": 5},
            {"name": "Send first invoice", "assignee": "finance", "due_days": 7}
        ]
    })
    results["actions_completed"].append({"action": "clickup_project", "success": project_result["success"]})
    results["project_url"] = project_result.get("project_url")

    # ═══════════════════════════════════════════════════════════════
    # 3. WELCOME EMAIL - Send to client
    # ═══════════════════════════════════════════════════════════════

    if contact.get("email"):
        # Generate personalized welcome email
        email_content = call_llm(
            prompt=f"""Write a warm, professional welcome email for a new client.

Client: {contact.get('firstname', 'there')} at {company.get('name')}
Deal: {deal_name}
Our company: [Your Company]

Include:
- Gratitude for choosing us
- What happens next (onboarding process)
- Who their point of contact will be
- How to reach us

Keep it concise and friendly.
"""
        )

        email_result = trigger_skill("gmail-client-outreach", {
            "to": contact["email"],
            "subject": f"Welcome to [Your Company], {contact.get('firstname', '')}!",
            "body": email_content,
            "track": True
        })
        results["actions_completed"].append({"action": "welcome_email", "success": email_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 4. DOCUMENTATION - Create client wiki in Notion
    # ═══════════════════════════════════════════════════════════════

    notion_result = trigger_skill("notion-client-database", {
        "action": "create_page",
        "database": "Clients",
        "properties": {
            "Name": company.get("name", deal_name),
            "Status": "Onboarding",
            "Deal Value": amount,
            "Primary Contact": contact.get("email"),
            "Start Date": datetime.now().isoformat(),
            "HubSpot ID": deal_id
        },
        "content": f"""
# {company.get('name', deal_name)}

## Overview
- **Deal Closed**: {datetime.now().strftime('%B %d, %Y')}
- **Value**: ${amount:,.2f}
- **Primary Contact**: {contact.get('firstname', '')} {contact.get('lastname', '')}

## Onboarding Checklist
- [ ] Welcome email sent
- [ ] Kickoff call scheduled
- [ ] Workspace set up
- [ ] Project plan created
- [ ] First invoice sent

## Notes
*Add meeting notes, decisions, and important context here*

## Links
- [HubSpot Deal](https://app.hubspot.com/deals/{deal_id})
- [ClickUp Project]({project_result.get('project_url', '#')})
"""
    })
    results["actions_completed"].append({"action": "notion_page", "success": notion_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 5. UPDATE CRM - Set lifecycle stage
    # ═══════════════════════════════════════════════════════════════

    if contact.get("id"):
        hubspot_result = trigger_skill("hubspot-contact-enrichment", {
            "action": "update",
            "contact_id": contact["id"],
            "properties": {
                "lifecyclestage": "customer",
                "hs_lead_status": "CONNECTED",
                "became_customer_date": datetime.now().isoformat()
            }
        })
        results["actions_completed"].append({"action": "hubspot_update", "success": hubspot_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # 6. SCHEDULE KICKOFF - Book meeting
    # ═══════════════════════════════════════════════════════════════

    if contact.get("email"):
        calendar_result = trigger_skill("calendar-smart-scheduling", {
            "action": "propose_meeting",
            "attendees": [contact["email"], deal.get("owner_email")],
            "title": f"Kickoff Call: {company.get('name', deal_name)}",
            "duration": 45,
            "preferred_days": 3,  # Within next 3 business days
            "description": f"Kickoff call to begin onboarding for {company.get('name')}."
        })
        results["actions_completed"].append({"action": "kickoff_scheduled", "success": calendar_result["success"]})

    # ═══════════════════════════════════════════════════════════════
    # SUMMARY
    # ═══════════════════════════════════════════════════════════════

    successful = sum(1 for a in results["actions_completed"] if a["success"])
    total = len(results["actions_completed"])

    results["summary"] = f"{successful}/{total} actions completed successfully"
    results["success"] = successful == total

    return_result(results)


def get_deal_details(deal_id):
    """Fetch deal from HubSpot."""
    creds = get_credential("hubspot")
    # ... API call
    pass

def get_associated_contact(deal_id):
    """Get primary contact for deal."""
    # ... API call
    pass

def get_associated_company(deal_id):
    """Get company for deal."""
    # ... API call
    pass
```

## Trigger Configuration

```yaml
trigger:
  type: webhook
  source: hubspot
  event: deal.propertyChange
  filter:
    property: dealstage
    value: closedwon
```

## Evolution History
- **v1**: Slack notification only
- **v2**: Added ClickUp project creation
- **v3**: Full workflow - email, Notion, calendar, CRM updates

## Learned Patterns
- Run independent tasks in parallel for speed
- Always include links between systems (HubSpot ↔ ClickUp ↔ Notion)
- Personalize communications using client data
- Track which actions succeeded for debugging
```

---

### Skill Examples: Meta Skills

#### skill-creator.md

```markdown
# Skill Creator

## Skill Metadata
- **ID**: skill-creator-v2
- **Category**: meta
- **Role Affinity**: all
- **Version**: 2
- **Executions**: 89
- **Success Rate**: 94.4%

## Description
Meta-skill that creates new skills. When the agent encounters a task it doesn't
have a skill for, this skill generates a new one.

## Script

```python
from sorx import return_result, call_llm, save_skill, get_pattern, validate_skill

def execute(params: dict):
    """
    Create a new skill from a task description.

    Steps:
    1. Analyze the task requirements
    2. Identify the interface type and provider
    3. Load the appropriate pattern template
    4. Generate the skill script
    5. Validate and save
    """
    task = params["task"]
    context = params.get("context", {})

    # Step 1: Analyze task
    analysis = call_llm(
        prompt=f"""Analyze this task and determine what's needed:

Task: {task}
Context: {context}

Return JSON with:
- skill_name: snake_case name for the skill
- category: crm|communication|tasks|documents|meetings|finance|workflows
- provider: The primary service (hubspot, gmail, clickup, etc.)
- interface_type: rest_api|graphql|database|desktop|file_system
- description: What this skill does
- required_credentials: List of credential keys needed
- inputs: List of expected input parameters
- outputs: What the skill returns
"""
    )

    # Step 2: Load pattern template
    pattern = get_pattern(
        interface_type=analysis["interface_type"],
        provider=analysis["provider"]
    )

    # Step 3: Generate skill script
    script = call_llm(
        prompt=f"""Generate a Sorx skill script.

Skill: {analysis['skill_name']}
Description: {analysis['description']}
Provider: {analysis['provider']}
Inputs: {analysis['inputs']}
Outputs: {analysis['outputs']}

Use this pattern as a template:
{pattern}

Requirements:
- Use get_credential() for auth, never hardcode
- Use return_result() for all outputs
- Include proper error handling
- Add docstring explaining usage
- Handle common edge cases

Return only the Python code.
"""
    )

    # Step 4: Validate
    validation = validate_skill(script, analysis)

    if not validation["valid"]:
        # Try to fix issues
        script = call_llm(
            prompt=f"""Fix these issues in the skill:

Script:
{script}

Issues:
{validation['issues']}

Return the corrected code only.
"""
        )
        validation = validate_skill(script, analysis)

    if validation["valid"]:
        # Step 5: Save skill
        skill = save_skill(
            name=analysis["skill_name"],
            category=analysis["category"],
            provider=analysis["provider"],
            description=analysis["description"],
            script=script,
            credentials=analysis["required_credentials"],
            version=1
        )

        return_result({
            "success": True,
            "skill_id": skill["id"],
            "skill_name": analysis["skill_name"],
            "message": f"Created new skill: {analysis['skill_name']}"
        })
    else:
        return_result({
            "success": False,
            "error": "Could not generate valid skill",
            "issues": validation["issues"]
        })
```

## Example Usage

**User Request:**
"I need to check our Stripe balance and send a Slack message if it's below $10,000"

**Skill Creator Output:**
```python
# Generated skill: stripe-balance-alert-v1

from sorx import get_credential, return_result, trigger_skill

def execute(params: dict):
    """
    Check Stripe balance and alert if below threshold.

    Params:
        threshold: Minimum balance (default: 10000)
        channel: Slack channel for alerts (default: #finance)
    """
    creds = get_credential("stripe")
    threshold = params.get("threshold", 10000)

    response = requests.get(
        "https://api.stripe.com/v1/balance",
        headers={"Authorization": f"Bearer {creds['secret_key']}"}
    )

    if response.ok:
        balance = response.json()
        available = sum(b["amount"] for b in balance["available"]) / 100

        if available < threshold:
            trigger_skill("slack-team-notification", {
                "channel": params.get("channel", "#finance"),
                "type": "warning",
                "title": "Low Stripe Balance Alert",
                "message": f"Current balance: ${available:,.2f}",
                "fields": {"Threshold": f"${threshold:,}"}
            })

        return_result({
            "success": True,
            "balance": available,
            "alert_sent": available < threshold
        })
    else:
        return_result({"success": False, "error": response.text})
```

## Evolution History
- **v1**: Basic skill generation
- **v2**: Added validation, auto-fix, pattern loading
```

---

## Skill Library

A shared repository of learned skills that can be discovered and reused:

### Skill Discovery

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SKILL LIBRARY                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Search: [hubspot deal notification____________] [Search]                  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_deal_stage_notification                                      │  │
│   │ ─────────────────────────────────                                    │  │
│   │ Sends Slack notification when HubSpot deal changes stage            │  │
│   │                                                                       │  │
│   │ Version: 5  │  Executions: 12,847  │  Success: 99.2%                │  │
│   │ Author: BusinessOS Team  │  Rating: 4.8/5                           │  │
│   │                                                                       │  │
│   │ Requires: hubspot, slack                                             │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_deal_to_clickup_task                                         │  │
│   │ ────────────────────────────                                         │  │
│   │ Creates ClickUp task when HubSpot deal closes                       │  │
│   │                                                                       │  │
│   │ Version: 3  │  Executions: 3,241  │  Success: 97.8%                 │  │
│   │ Author: Community  │  Rating: 4.5/5                                 │  │
│   │                                                                       │  │
│   │ Requires: hubspot, clickup                                           │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ hubspot_closed_deal_summary_email                                    │  │
│   │ ───────────────────────────────                                      │  │
│   │ Sends daily summary email of closed deals                           │  │
│   │                                                                       │  │
│   │ Version: 2  │  Executions: 892  │  Success: 98.1%                   │  │
│   │ Author: Community  │  Rating: 4.2/5                                 │  │
│   │                                                                       │  │
│   │ Requires: hubspot, gmail                                             │  │
│   │ [Install] [View Code] [Fork]                                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   Categories: [All] [CRM] [Communication] [Tasks] [Files] [Custom]        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Publishing

When a skill reaches "mastery" (high execution count + success rate), users can publish to the library:

```python
class SkillPublisher:
    def can_publish(self, skill: Skill) -> bool:
        """Check if skill meets publishing criteria."""
        return (
            skill.executions >= 50 and
            skill.success_rate >= 0.95 and
            skill.version >= 2  # Has been improved at least once
        )

    def publish(self, skill: Skill, user: User) -> PublishedSkill:
        """Publish skill to community library."""
        # Security review
        self.security_review(skill)

        # Remove user-specific data
        sanitized = self.sanitize_skill(skill)

        # Publish
        return PublishedSkill(
            skill=sanitized,
            author=user,
            published_at=now()
        )
```

---

## Comparison: Sorx 2.0 vs Composio vs Traditional

| Feature | Composio | Traditional MCP | Sorx 2.0 |
|---------|----------|-----------------|----------|
| **Tool Count** | 250+ pre-built | Custom built | Unlimited (generated) |
| **New Integrations** | Wait for dev team | Build yourself | Agent learns on-demand |
| **Legacy Systems** | Limited | Build yourself | Universal adapters |
| **Credential Security** | Cloud-stored | Context-exposed | Local vault only |
| **Learning** | None | None | Skills improve over time |
| **Customization** | Limited config | Full code | Auto-customized |
| **Context Usage** | Tool definitions | Tool definitions | Minimal patterns |
| **Enterprise Systems** | Some | Manual | Full support |
| **Desktop Apps** | No | No | Yes (OS automation) |
| **Hardware/IoT** | No | No | Yes |
| **Offline Capable** | No | No | Yes |
| **Skill Sharing** | No | No | Community library |
| **Small Business** | Works | Complex | Efficient |
| **Enterprise** | Works | Complex | Carrier-grade |

---

## Enterprise Features

### Carrier-Grade Reliability

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      ENTERPRISE RELIABILITY FEATURES                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   HIGH AVAILABILITY                                                         │
│   - Multi-region Sorx Engine deployment                                    │
│   - Automatic failover                                                      │
│   - Zero-downtime updates                                                   │
│                                                                             │
│   SECURITY                                                                  │
│   - SOC 2 Type II compliant                                                │
│   - GDPR compliant                                                          │
│   - End-to-end encryption                                                   │
│   - Credential vault with HSM backing                                      │
│   - Audit logging for compliance                                           │
│                                                                             │
│   SCALABILITY                                                               │
│   - Horizontal scaling of skill execution                                  │
│   - Rate limit handling                                                     │
│   - Queue-based execution for high volume                                  │
│                                                                             │
│   MONITORING                                                                │
│   - Real-time skill performance metrics                                    │
│   - Alerting on failures                                                   │
│   - Detailed execution logs                                                │
│                                                                             │
│   GOVERNANCE                                                                │
│   - Skill approval workflows                                               │
│   - Access control by role                                                 │
│   - Data classification enforcement                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Small Business Efficiency

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      SMALL BUSINESS EFFICIENCY                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ZERO CONFIGURATION                                                        │
│   - Skills generated automatically                                         │
│   - No developer required                                                   │
│   - Natural language requests                                              │
│                                                                             │
│   COST EFFICIENT                                                            │
│   - Only pay for what you use                                              │
│   - No pre-built tool licensing                                            │
│   - Skills reused across requests                                          │
│                                                                             │
│   FAST TIME TO VALUE                                                        │
│   - First skill in minutes                                                 │
│   - No integration project                                                 │
│   - Immediate automation                                                   │
│                                                                             │
│   SIMPLE MANAGEMENT                                                         │
│   - Single dashboard                                                       │
│   - Clear skill inventory                                                  │
│   - Easy credential management                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Complete System Architecture

### Master Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                      SORX 2.0 MASTER ARCHITECTURE                                     │
├──────────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                                       │
│   EXTERNAL CLIENTS (Any LLM, Any System)                                                             │
│   ══════════════════════════════════════                                                             │
│                                                                                                       │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐          │
│   │   ChatGPT   │   │   Claude    │   │   Gemini    │   │  Custom LLM │   │  3rd Party  │          │
│   │   Plugin    │   │   Agent     │   │   Agent     │   │   Backend   │   │    Apps     │          │
│   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘          │
│          │                 │                 │                 │                 │                   │
│          └─────────────────┴─────────────────┴─────────────────┴─────────────────┘                   │
│                                              │                                                        │
│                                              ▼                                                        │
│   ┌──────────────────────────────────────────────────────────────────────────────────────────────┐  │
│   │                              SORX API GATEWAY (Public)                                        │  │
│   │                                                                                               │  │
│   │  POST /api/v1/skill/execute     - Execute a skill                                            │  │
│   │  POST /api/v1/skill/discover    - Find skills for a task                                     │  │
│   │  POST /api/v1/skill/generate    - Generate new skill on-demand                               │  │
│   │  GET  /api/v1/skill/{id}        - Get skill details                                          │  │
│   │  GET  /api/v1/skill/library     - Browse skill library                                       │  │
│   │  POST /api/v1/request           - Natural language request (skill selection handled)         │  │
│   │                                                                                               │  │
│   │  Auth: API Key, OAuth2, JWT                                                                  │  │
│   │  Rate Limiting: Per-tenant, burst protection                                                 │  │
│   │                                                                                               │  │
│   └────────────────────────────────────────────┬─────────────────────────────────────────────────┘  │
│                                                │                                                     │
├────────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│                                                │                                                     │
│   BUSINESSOS CLOUD                             │                                                     │
│   ════════════════                             │                                                     │
│                                                ▼                                                     │
│   ┌──────────────────────────────────────────────────────────────────────────────────────────────┐  │
│   │                                    SKILL ROUTER                                               │  │
│   │                                                                                               │  │
│   │  ┌─────────────────────────────────────────────────────────────────────────────────────┐   │  │
│   │  │                            SKILL SELECTION ENGINE                                     │   │  │
│   │  │                                                                                       │   │  │
│   │  │  1. INTENT ANALYSIS                                                                  │   │  │
│   │  │     └─ Parse natural language → Extract intent, entities, parameters                │   │  │
│   │  │                                                                                       │   │  │
│   │  │  2. SKILL MATCHING                                                                   │   │  │
│   │  │     └─ Vector search → Find skills by semantic similarity                           │   │  │
│   │  │     └─ Capability match → Filter by required capabilities                           │   │  │
│   │  │     └─ Permission check → Filter by tenant/user access                              │   │  │
│   │  │                                                                                       │   │  │
│   │  │  3. SKILL SELECTION                                                                  │   │  │
│   │  │     └─ Exact match → Use existing skill directly                                    │   │  │
│   │  │     └─ Adapt match → Modify skill parameters                                        │   │  │
│   │  │     └─ Compose match → Chain multiple skills                                        │   │  │
│   │  │     └─ Generate → Create new skill if no match                                      │   │  │
│   │  │                                                                                       │   │  │
│   │  │  4. EXECUTION ROUTING                                                                │   │  │
│   │  │     └─ Route to user's Sorx Engine (local execution)                                │   │  │
│   │  │     └─ OR route to shared Sorx Engine (cloud execution)                             │   │  │
│   │  │                                                                                       │   │  │
│   │  └─────────────────────────────────────────────────────────────────────────────────────┘   │  │
│   │                                                                                               │  │
│   └────────────────────────────────────────────┬─────────────────────────────────────────────────┘  │
│                                                │                                                     │
│          ┌─────────────────────────────────────┼─────────────────────────────────────────┐          │
│          │                                     │                                         │          │
│          ▼                                     ▼                                         ▼          │
│   ┌─────────────┐                    ┌─────────────┐                           ┌─────────────┐     │
│   │   Agent     │                    │   Skill     │                           │   Skill     │     │
│   │  Workspace  │◀──────────────────▶│ Generator   │                           │  Library    │     │
│   │             │                    │   (LLM)     │                           │  (Global)   │     │
│   │ @sales      │                    └─────────────┘                           └─────────────┘     │
│   │ @support    │                           │                                        │             │
│   │ @ops        │                           │ Generates skills                       │             │
│   │ @finance    │                           ▼                                        ▼             │
│   └──────┬──────┘                    ┌─────────────┐                           ┌─────────────┐     │
│          │                           │  Pattern    │                           │  Tenant     │     │
│          │ Callbacks                 │  Library    │                           │   Skills    │     │
│          │                           │ (Templates) │                           │ (Private)   │     │
│          │                           └─────────────┘                           └─────────────┘     │
│          │                                                                                          │
│   ┌──────▼──────────────────────────────────────────────────────────────────────────────────────┐  │
│   │                            ORCHESTRATION HUB                                                  │  │
│   │                                                                                               │  │
│   │  ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐            │  │
│   │  │  Message  │   │ Callback  │   │ Execution │   │  Context  │   │  Result   │            │  │
│   │  │    Bus    │   │  Router   │   │  Tracker  │   │  Manager  │   │ Aggregator│            │  │
│   │  └───────────┘   └───────────┘   └───────────┘   └───────────┘   └───────────┘            │  │
│   │                                                                                               │  │
│   └────────────────────────────────────────────┬─────────────────────────────────────────────────┘  │
│                                                │                                                     │
│                                                │ WebSocket                                          │
│                                                │                                                     │
├────────────────────────────────────────────────┼─────────────────────────────────────────────────────┤
│                                                │                                                     │
│   SORX ENGINE (Local on User's Machine)        │                                                     │
│   ═════════════════════════════════════        ▼                                                     │
│                                                                                                       │
│   ┌──────────────────────────────────────────────────────────────────────────────────────────────┐  │
│   │                                    EXECUTION RUNTIME                                          │  │
│   │                                                                                               │  │
│   │  ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐             │  │
│   │  │  Credential   │   │    Skill      │   │   Sandbox     │   │   Callback    │             │  │
│   │  │    Vault      │   │    Cache      │   │   Executor    │   │   Handler     │             │  │
│   │  │              │   │               │   │               │   │               │             │  │
│   │  │  AES-256     │   │  Local copy   │   │  Isolated     │   │  Bidirectional│             │  │
│   │  │  encrypted   │   │  of skills    │   │  Python env   │   │  with cloud   │             │  │
│   │  └───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘             │  │
│   │                                                                                               │  │
│   │  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │  │
│   │  │                              INTERFACE ADAPTERS                                        │  │  │
│   │  │                                                                                        │  │  │
│   │  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐            │  │  │
│   │  │  │ REST  │ │  DB   │ │Legacy │ │Desktop│ │ File  │ │ IoT   │ │ Comm  │            │  │  │
│   │  │  │ API   │ │       │ │ SOAP  │ │ Auto  │ │ Sys   │ │       │ │       │            │  │  │
│   │  │  └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘            │  │  │
│   │  │      │         │         │         │         │         │         │                 │  │  │
│   │  └──────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘  │  │
│   │                                                                                               │  │
│   └────────────────────────────────────────────┬─────────────────────────────────────────────────┘  │
│                                                │                                                     │
│                                                ▼                                                     │
│   ┌──────────────────────────────────────────────────────────────────────────────────────────────┐  │
│   │                                    TARGET SYSTEMS                                             │  │
│   │                                                                                               │  │
│   │   Modern APIs        Legacy Systems        Desktop Apps        Hardware/IoT        Files     │  │
│   │   ────────────       ──────────────        ────────────        ────────────        ─────     │  │
│   │   HubSpot            SOAP Services         Excel               IoT Devices         Local     │  │
│   │   Slack              EDI                   Outlook             Serial Ports        S3/GCS    │  │
│   │   Gmail              FTP/SFTP              QuickBooks          Printers            SFTP      │  │
│   │   ClickUp            Mainframe             Photoshop           Scanners            Network   │  │
│   │   Notion             AS/400                Terminal            Modbus              Shares    │  │
│   │                                                                                               │  │
│   └──────────────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                                       │
└──────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Communication Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COMMUNICATION PROTOCOLS                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   LAYER                    PROTOCOL                PURPOSE                  │
│   ═════                    ════════                ═══════                  │
│                                                                             │
│   External → Gateway       HTTPS + REST            Public API access        │
│                            gRPC (optional)         High-performance calls   │
│                                                                             │
│   Gateway → Skill Router   Internal gRPC           Service-to-service      │
│                            mTLS                    Secure internal          │
│                                                                             │
│   Router → Orchestrator    Redis Pub/Sub           Event-driven routing    │
│                            NATS (optional)         Message queue           │
│                                                                             │
│   Orchestrator → Engine    WebSocket               Bidirectional real-time │
│                            SSE (fallback)          One-way streaming       │
│                                                                             │
│   Engine → Target Systems  Per-adapter             HTTP, DB drivers, etc.  │
│                            (varies)                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Selection Decision Tree

The Skill Router uses this decision tree to select the right skill:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SKILL SELECTION DECISION TREE                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Incoming Request                                                          │
│   ════════════════                                                          │
│         │                                                                   │
│         ▼                                                                   │
│   ┌─────────────┐                                                           │
│   │  EXPLICIT   │ Request includes skill_id?                               │
│   │   SKILL?    │                                                           │
│   └──────┬──────┘                                                           │
│          │                                                                   │
│     YES  │   NO                                                             │
│          │    │                                                             │
│          │    ▼                                                             │
│          │   ┌─────────────┐                                                │
│          │   │   PARSE     │ Extract intent, entities, parameters          │
│          │   │   REQUEST   │ from natural language                         │
│          │   └──────┬──────┘                                                │
│          │          │                                                       │
│          │          ▼                                                       │
│          │   ┌─────────────┐                                                │
│          │   │   SEARCH    │ Vector similarity search +                    │
│          │   │   SKILLS    │ Capability matching                           │
│          │   └──────┬──────┘                                                │
│          │          │                                                       │
│          │          ▼                                                       │
│          │   ┌─────────────┐                                                │
│          │   │   MATCH     │                                                │
│          │   │   FOUND?    │                                                │
│          │   └──────┬──────┘                                                │
│          │          │                                                       │
│          │     ┌────┴────┐                                                  │
│          │     │         │                                                  │
│          │    YES       NO                                                  │
│          │     │         │                                                  │
│          │     ▼         ▼                                                  │
│          │   ┌─────────────┐    ┌─────────────┐                            │
│          │   │  SINGLE OR  │    │  GENERATE   │ Create new skill           │
│          │   │  MULTIPLE?  │    │    NEW      │ using pattern library      │
│          │   └──────┬──────┘    └──────┬──────┘                            │
│          │          │                  │                                    │
│          │     ┌────┴────┐             │                                    │
│          │     │         │             │                                    │
│          │  SINGLE   MULTIPLE          │                                    │
│          │     │         │             │                                    │
│          │     │         ▼             │                                    │
│          │     │   ┌─────────────┐     │                                    │
│          │     │   │   SELECT    │     │ Rank by: success rate,            │
│          │     │   │    BEST     │     │ execution time, version,          │
│          │     │   │    MATCH    │     │ user preference                   │
│          │     │   └──────┬──────┘     │                                    │
│          │     │          │            │                                    │
│          │     └────┬─────┘            │                                    │
│          │          │                  │                                    │
│          └──────────┴──────────────────┘                                    │
│                     │                                                       │
│                     ▼                                                       │
│            ┌─────────────┐                                                  │
│            │   COMPOSE   │ Is this a multi-step task?                      │
│            │   NEEDED?   │                                                  │
│            └──────┬──────┘                                                  │
│                   │                                                         │
│              ┌────┴────┐                                                    │
│              │         │                                                    │
│             YES       NO                                                    │
│              │         │                                                    │
│              ▼         │                                                    │
│         ┌─────────────┐│                                                    │
│         │   CREATE    ││ Build workflow from                               │
│         │  COMPOUND   ││ multiple skills                                   │
│         │   SKILL     ││                                                    │
│         └──────┬──────┘│                                                    │
│                │       │                                                    │
│                └───┬───┘                                                    │
│                    │                                                        │
│                    ▼                                                        │
│            ┌─────────────┐                                                  │
│            │  VALIDATE   │ Check credentials, permissions                  │
│            │   ACCESS    │                                                  │
│            └──────┬──────┘                                                  │
│                   │                                                         │
│                   ▼                                                         │
│            ┌─────────────┐                                                  │
│            │   ROUTE     │ Send to appropriate Sorx Engine                 │
│            │  EXECUTION  │ (local or cloud)                                │
│            └─────────────┘                                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Where Does Selection Happen?

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SKILL SELECTION: CLOUD vs LOCAL                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                 SKILL SELECTION (BusinessOS Cloud)                   │  │
│   │                                                                       │  │
│   │   WHY IN CLOUD:                                                      │  │
│   │   ✓ Access to global skill library                                  │  │
│   │   ✓ Vector search across all skills                                 │  │
│   │   ✓ LLM access for intent parsing                                   │  │
│   │   ✓ Permission enforcement                                          │  │
│   │   ✓ Usage tracking and billing                                      │  │
│   │   ✓ Skill generation with pattern library                           │  │
│   │                                                                       │  │
│   │   DOES:                                                              │  │
│   │   • Parse request intent                                            │  │
│   │   • Find matching skills                                            │  │
│   │   • Generate new skills if needed                                   │  │
│   │   • Check permissions                                               │  │
│   │   • Select execution target                                         │  │
│   │                                                                       │  │
│   │   DOES NOT:                                                          │  │
│   │   • Access user credentials                                         │  │
│   │   • Execute skill code                                              │  │
│   │   • Touch target systems                                            │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                 SKILL EXECUTION (Sorx Engine - Local)                │  │
│   │                                                                       │  │
│   │   WHY LOCAL:                                                         │  │
│   │   ✓ Credentials NEVER leave user's machine                          │  │
│   │   ✓ Can access local resources (files, desktop apps)                │  │
│   │   ✓ Can access on-prem systems                                      │  │
│   │   ✓ Lower latency for target systems                                │  │
│   │   ✓ User has full control                                           │  │
│   │                                                                       │  │
│   │   DOES:                                                              │  │
│   │   • Execute skill Python code                                       │  │
│   │   • Access credential vault                                         │  │
│   │   • Connect to target systems                                       │  │
│   │   • Handle callbacks                                                │  │
│   │   • Return results                                                  │  │
│   │                                                                       │  │
│   │   DOES NOT:                                                          │  │
│   │   • Parse intent (already done)                                     │  │
│   │   • Select skills (already done)                                    │  │
│   │   • Access other tenants' data                                      │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                 CLOUD EXECUTION (Optional - Sorx Cloud Engine)       │  │
│   │                                                                       │  │
│   │   WHEN TO USE:                                                       │  │
│   │   • User doesn't have local Sorx Engine                             │  │
│   │   • Task doesn't need local resources                               │  │
│   │   • User prefers cloud-managed credentials                          │  │
│   │   • Simpler setup for small operations                              │  │
│   │                                                                       │  │
│   │   SECURITY:                                                          │  │
│   │   • Credentials stored in cloud HSM                                 │  │
│   │   • Isolated per-tenant execution                                   │  │
│   │   • SOC 2 compliant infrastructure                                  │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Scalability Architecture

### Target: 100 Million+ Executions

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     SCALABILITY TARGETS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   METRIC                          TARGET                 STRATEGY          │
│   ══════                          ══════                 ════════          │
│   Skill executions/month          100,000,000           Distributed        │
│   Concurrent executions           50,000                Horizontal scale   │
│   Execution latency (P95)         < 500ms               Edge routing       │
│   API latency (P95)               < 100ms               CDN + caching      │
│   Skills in library               1,000,000+            Sharded storage    │
│   Tenants                         100,000+              Multi-tenant       │
│   Uptime                          99.99%                Multi-region       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Distributed Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                              GLOBAL DISTRIBUTION                                          │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                           │
│   ┌───────────────────────────────────────────────────────────────────────────────────┐  │
│   │                                 CLOUDFLARE / CDN                                   │  │
│   │                                                                                    │  │
│   │   • API Gateway routing                                                           │  │
│   │   • DDoS protection                                                               │  │
│   │   • Rate limiting at edge                                                         │  │
│   │   • SSL termination                                                               │  │
│   │                                                                                    │  │
│   └────────────────────────────────────────┬──────────────────────────────────────────┘  │
│                                            │                                              │
│                    ┌───────────────────────┼───────────────────────┐                     │
│                    │                       │                       │                     │
│                    ▼                       ▼                       ▼                     │
│            ┌─────────────┐         ┌─────────────┐         ┌─────────────┐              │
│            │  US-WEST    │         │  US-EAST    │         │   EU-WEST   │              │
│            │  (Primary)  │         │ (Secondary) │         │ (Secondary) │              │
│            └──────┬──────┘         └──────┬──────┘         └──────┬──────┘              │
│                   │                       │                       │                     │
│                   ▼                       ▼                       ▼                     │
│   ┌───────────────────────────────────────────────────────────────────────────────────┐  │
│   │                              REGIONAL CLUSTER (each)                               │  │
│   │                                                                                    │  │
│   │   ┌─────────────────────────────────────────────────────────────────────────┐    │  │
│   │   │                        KUBERNETES (GKE / EKS)                            │    │  │
│   │   │                                                                          │    │  │
│   │   │   API Gateway Pods (HPA: 3-100)                                         │    │  │
│   │   │   ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ... ┌────┐                       │    │  │
│   │   │   │ GW │ │ GW │ │ GW │ │ GW │ │ GW │     │ GW │                       │    │  │
│   │   │   └────┘ └────┘ └────┘ └────┘ └────┘     └────┘                       │    │  │
│   │   │                                                                          │    │  │
│   │   │   Skill Router Pods (HPA: 5-200)                                        │    │  │
│   │   │   ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ... ┌────┐                       │    │  │
│   │   │   │ SR │ │ SR │ │ SR │ │ SR │ │ SR │     │ SR │                       │    │  │
│   │   │   └────┘ └────┘ └────┘ └────┘ └────┘     └────┘                       │    │  │
│   │   │                                                                          │    │  │
│   │   │   Orchestration Pods (HPA: 3-50)                                        │    │  │
│   │   │   ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐                                  │    │  │
│   │   │   │ OH │ │ OH │ │ OH │ │ OH │ │ OH │                                  │    │  │
│   │   │   └────┘ └────┘ └────┘ └────┘ └────┘                                  │    │  │
│   │   │                                                                          │    │  │
│   │   │   Skill Generator Pods (HPA: 2-20) - LLM calls                          │    │  │
│   │   │   ┌────┐ ┌────┐ ┌────┐                                                │    │  │
│   │   │   │ SG │ │ SG │ │ SG │                                                │    │  │
│   │   │   └────┘ └────┘ └────┘                                                │    │  │
│   │   │                                                                          │    │  │
│   │   └──────────────────────────────────────────────────────────────────────────┘    │  │
│   │                                                                                    │  │
│   │   ┌─────────────────────────────────────────────────────────────────────────┐    │  │
│   │   │                          DATA LAYER                                      │    │  │
│   │   │                                                                          │    │  │
│   │   │   ┌────────────────┐   ┌────────────────┐   ┌────────────────┐         │    │  │
│   │   │   │  PostgreSQL    │   │     Redis      │   │  Elasticsearch │         │    │  │
│   │   │   │  (Cloud SQL)   │   │   (Memorystore)│   │  (Skill Search)│         │    │  │
│   │   │   │                │   │                │   │                │         │    │  │
│   │   │   │  • Skills      │   │  • Skill cache │   │  • Vector      │         │    │  │
│   │   │   │  • Executions  │   │  • Session     │   │    embeddings  │         │    │  │
│   │   │   │  • Tenants     │   │  • Rate limits │   │  • Full-text   │         │    │  │
│   │   │   └────────────────┘   └────────────────┘   └────────────────┘         │    │  │
│   │   │                                                                          │    │  │
│   │   └──────────────────────────────────────────────────────────────────────────┘    │  │
│   │                                                                                    │  │
│   │   ┌─────────────────────────────────────────────────────────────────────────┐    │  │
│   │   │                         MESSAGE QUEUE                                    │    │  │
│   │   │                                                                          │    │  │
│   │   │   Cloud Pub/Sub (or NATS)                                               │    │  │
│   │   │   • Execution requests                                                  │    │  │
│   │   │   • Callback routing                                                    │    │  │
│   │   │   • Event distribution                                                  │    │  │
│   │   │                                                                          │    │  │
│   │   └──────────────────────────────────────────────────────────────────────────┘    │  │
│   │                                                                                    │  │
│   └────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                           │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

### Scaling Strategies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SCALING STRATEGIES                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   1. HORIZONTAL POD AUTOSCALING (HPA)                                       │
│   ═══════════════════════════════════                                       │
│                                                                             │
│   Component          Min    Max    Metric              Threshold            │
│   ─────────          ───    ───    ──────              ─────────            │
│   API Gateway        3      100    CPU                 70%                  │
│   Skill Router       5      200    Request queue       1000 pending         │
│   Orchestrator       3      50     WebSocket conns     5000/pod            │
│   Skill Generator    2      20     LLM queue depth     10 pending          │
│                                                                             │
│   2. DATABASE SHARDING                                                      │
│   ════════════════════                                                      │
│                                                                             │
│   Table              Shard Key           Shards     Strategy               │
│   ─────              ─────────           ──────     ────────               │
│   skills             tenant_id           16         Hash                   │
│   executions         tenant_id + date    64         Hash + Time            │
│   skill_versions     skill_id            16         Hash                   │
│                                                                             │
│   3. CACHING LAYERS                                                         │
│   ═════════════════                                                         │
│                                                                             │
│   Layer              Data                  TTL        Hit Rate Target       │
│   ─────              ────                  ───        ───────────────       │
│   L1 (Pod)           Skill metadata        5m         95%                  │
│   L2 (Redis)         Skill code            1h         90%                  │
│   L3 (CDN)           Public skills         24h        80%                  │
│                                                                             │
│   4. ASYNC PROCESSING                                                       │
│   ═══════════════════                                                       │
│                                                                             │
│   • Skill generation queued via Pub/Sub                                    │
│   • Execution results batched for analytics                                │
│   • Non-blocking callbacks via message queue                               │
│   • Eventual consistency for skill metrics                                 │
│                                                                             │
│   5. CONNECTION POOLING                                                     │
│   ═════════════════════                                                     │
│                                                                             │
│   • pgBouncer for PostgreSQL (transaction mode)                            │
│   • Redis Cluster with client-side pooling                                 │
│   • WebSocket connection reuse                                             │
│   • HTTP/2 for API Gateway                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Load Patterns

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           LOAD PATTERNS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   PATTERN 1: BURST (Webhook triggers)                                       │
│   ═════════════════════════════════════                                     │
│                                                                             │
│   [──────▲──────────────]   10,000 skills triggered by HubSpot webhook     │
│         ████                 in 5 seconds                                   │
│        █    █                                                               │
│       █      █               Solution:                                      │
│      █        █              • Queue with Pub/Sub                          │
│   ──█──────────█───          • Process at steady rate                      │
│                              • Acknowledge immediately                      │
│                                                                             │
│   PATTERN 2: STEADY (Background jobs)                                       │
│   ════════════════════════════════════                                      │
│                                                                             │
│   [████████████████████]   100 skills/second consistently                  │
│                              24/7                                           │
│                                                                             │
│                              Solution:                                      │
│                              • Baseline scaling                            │
│                              • Predictive autoscaling                      │
│                              • Reserved capacity                           │
│                                                                             │
│   PATTERN 3: SPIKE (Business hours)                                         │
│   ══════════════════════════════════                                        │
│                                                                             │
│   [──▲──────▲──────▲──]     3x traffic during business hours              │
│      █       █       █                                                      │
│     ███     ███     ███     Solution:                                       │
│    █   █   █   █   █   █    • Scheduled scaling                            │
│   ─     ───     ───     ─   • Time-zone aware                              │
│                              • Warm pools                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Testing Strategy

### Testing Pyramid

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           TESTING PYRAMID                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                            ┌───────────┐                                    │
│                           /│    E2E    │\                                   │
│                          / │   Tests   │ \      5% - Real integrations     │
│                         /  │  (Slow)   │  \     Real HubSpot, Slack, etc.  │
│                        /   └───────────┘   \                               │
│                       /                     \                               │
│                      /   ┌───────────────┐   \                              │
│                     /    │  Integration  │    \    15% - Service tests     │
│                    /     │    Tests      │     \   Mocked external APIs    │
│                   /      │  (Medium)     │      \                           │
│                  /       └───────────────┘       \                          │
│                 /                                 \                          │
│                /     ┌───────────────────────┐     \                         │
│               /      │       Unit Tests       │      \  80% - Fast,        │
│              /       │        (Fast)          │       \ isolated tests     │
│             /        │                        │        \                    │
│            /         └───────────────────────┘         \                    │
│           ─────────────────────────────────────────────                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Test Categories

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          TEST CATEGORIES                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   1. UNIT TESTS                                                             │
│   ═════════════                                                             │
│                                                                             │
│   Component               Tests                           Coverage Target   │
│   ─────────               ─────                           ───────────────   │
│   Skill Selection         Intent parsing, matching         95%             │
│   Callback Handler        Message routing, timeouts        90%             │
│   Credential Vault        Encryption, access control       100%            │
│   Interface Adapters      Protocol handling                85%             │
│   Skill Generator         Template processing              80%             │
│                                                                             │
│   2. INTEGRATION TESTS                                                      │
│   ════════════════════                                                      │
│                                                                             │
│   Scenario                Mock                    Real                      │
│   ────────                ────                    ────                      │
│   Skill Router → Gen      ✓                                                │
│   Orchestrator → Engine   ✓                                                │
│   Engine → Adapters       ✓                                                │
│   Full callback flow      ✓                                                │
│                                                                             │
│   3. E2E TESTS (Sandbox accounts)                                           │
│   ═══════════════════════════════                                           │
│                                                                             │
│   Provider          Sandbox Account      Tests                             │
│   ────────          ───────────────      ─────                             │
│   HubSpot           Dev portal account   CRUD contacts, deals             │
│   Slack             Test workspace       Send messages, create channels   │
│   Gmail             Test account         Send/read emails                 │
│   ClickUp           Test workspace       CRUD tasks, projects             │
│   QuickBooks        Sandbox company      CRUD invoices                    │
│                                                                             │
│   4. SKILL-SPECIFIC TESTS                                                   │
│   ═══════════════════════                                                   │
│                                                                             │
│   Every skill must have:                                                    │
│   □ Input validation tests                                                 │
│   □ Success path tests                                                     │
│   □ Error handling tests                                                   │
│   □ Callback tests (if applicable)                                         │
│   □ Timeout tests                                                          │
│                                                                             │
│   5. CHAOS ENGINEERING                                                      │
│   ════════════════════                                                      │
│                                                                             │
│   Test                          Expected Behavior                          │
│   ────                          ─────────────────                          │
│   Kill random pods              Requests reroute, no failures             │
│   Network partition             Graceful degradation, retries             │
│   Redis failure                 Fallback to DB, slower                    │
│   Callback timeout              Skill fails gracefully                    │
│   LLM rate limit                Queue, retry with backoff                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Testing Framework

```python
# testing/skill_test_framework.py

import pytest
from sorx.testing import SkillTest, MockCredential, MockCallback

class TestHubSpotGetClient(SkillTest):
    """Test suite for hubspot_get_client skill."""

    skill_id = "hubspot_get_client_v2"

    @pytest.fixture
    def mock_credentials(self):
        """Provide mock HubSpot credentials."""
        return MockCredential(
            provider="hubspot",
            data={"access_token": "test_token_xxx"}
        )

    @pytest.fixture
    def mock_hubspot_api(self, requests_mock):
        """Mock HubSpot API responses."""
        requests_mock.get(
            "https://api.hubapi.com/crm/v3/objects/contacts/123",
            json={
                "id": "123",
                "properties": {
                    "firstname": "John",
                    "lastname": "Doe",
                    "email": "john@example.com"
                }
            }
        )
        return requests_mock

    def test_success_path(self, mock_credentials, mock_hubspot_api):
        """Test successful client fetch."""
        result = self.execute_skill(
            params={"client_id": "123"},
            credentials=mock_credentials
        )

        assert result["success"] is True
        assert result["client"]["email"] == "john@example.com"

    def test_not_found(self, mock_credentials, requests_mock):
        """Test handling of non-existent client."""
        requests_mock.get(
            "https://api.hubapi.com/crm/v3/objects/contacts/999",
            status_code=404
        )

        result = self.execute_skill(
            params={"client_id": "999"},
            credentials=mock_credentials
        )

        assert result["success"] is False
        assert "not found" in result["error"].lower()

    def test_callback(self, mock_credentials, mock_hubspot_api):
        """Test skill that uses callbacks."""
        callback_response = MockCallback(
            callback_type="data_request",
            response={"additional_fields": ["company"]}
        )

        result = self.execute_skill(
            params={"client_id": "123", "expand": True},
            credentials=mock_credentials,
            callbacks=[callback_response]
        )

        assert result["success"] is True
        assert callback_response.was_called

    def test_timeout(self, mock_credentials, requests_mock):
        """Test handling of API timeout."""
        requests_mock.get(
            "https://api.hubapi.com/crm/v3/objects/contacts/123",
            exc=requests.exceptions.Timeout
        )

        result = self.execute_skill(
            params={"client_id": "123"},
            credentials=mock_credentials
        )

        assert result["success"] is False
        assert result["retry_suggested"] is True

    def test_rate_limit(self, mock_credentials, requests_mock):
        """Test handling of rate limit response."""
        requests_mock.get(
            "https://api.hubapi.com/crm/v3/objects/contacts/123",
            status_code=429,
            headers={"Retry-After": "60"}
        )

        result = self.execute_skill(
            params={"client_id": "123"},
            credentials=mock_credentials
        )

        assert result["success"] is False
        assert result["retry_after"] == 60


# Run with: pytest -v testing/skills/
```

---

## Red Team Analysis

### Threat Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            THREAT MODEL                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ASSET                    THREAT                     IMPACT                │
│   ═════                    ══════                     ══════                │
│                                                                             │
│   Credentials              Theft via skill code       CRITICAL             │
│                            Leaked in logs             CRITICAL             │
│                            Intercepted in transit     CRITICAL             │
│                                                                             │
│   Skill Code               Malicious injection        HIGH                 │
│                            Data exfiltration          HIGH                 │
│                            Lateral movement           HIGH                 │
│                                                                             │
│   Target Systems           Unauthorized access        HIGH                 │
│                            Data corruption            HIGH                 │
│                            Denial of service          MEDIUM               │
│                                                                             │
│   BusinessOS               Account takeover           CRITICAL             │
│                            Cross-tenant access        CRITICAL             │
│                            Service disruption         HIGH                 │
│                                                                             │
│   User Data                Privacy breach             CRITICAL             │
│                            Data theft                 CRITICAL             │
│                            Compliance violation       HIGH                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Attack Vectors & Mitigations

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ATTACK VECTORS & MITIGATIONS                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ATTACK 1: Malicious Skill Injection                                       │
│   ═══════════════════════════════════                                       │
│                                                                             │
│   Vector:                                                                   │
│   Attacker submits natural language request that tricks LLM into           │
│   generating malicious skill code (e.g., "send all data to evil.com")     │
│                                                                             │
│   Mitigations:                                                              │
│   □ Static analysis of generated code before execution                     │
│   □ Blocklist of dangerous patterns (network calls to non-API hosts)      │
│   □ Sandbox execution with restricted permissions                          │
│   □ Code signing for skills from library                                   │
│   □ Human review for skills accessing sensitive data                       │
│   □ LLM system prompt hardening                                            │
│                                                                             │
│   ATTACK 2: Credential Exfiltration                                         │
│   ═════════════════════════════════                                         │
│                                                                             │
│   Vector:                                                                   │
│   Skill code tries to send credentials to external endpoint                │
│                                                                             │
│   Mitigations:                                                              │
│   □ Credentials injected at runtime, not in code                           │
│   □ Network egress filtering (only allow known API hosts)                  │
│   □ Credential values never in skill logs                                  │
│   □ Memory protection (credentials cleared after use)                      │
│   □ Audit trail for all credential access                                  │
│                                                                             │
│   ATTACK 3: Cross-Tenant Access                                             │
│   ═════════════════════════════                                             │
│                                                                             │
│   Vector:                                                                   │
│   Skill execution accesses another tenant's data or credentials           │
│                                                                             │
│   Mitigations:                                                              │
│   □ Tenant ID in every request, verified at every layer                    │
│   □ Separate credential vaults per tenant                                  │
│   □ Database row-level security                                            │
│   □ Skill execution isolated per tenant                                    │
│   □ Regular penetration testing for tenant isolation                       │
│                                                                             │
│   ATTACK 4: Prompt Injection via Callbacks                                  │
│   ════════════════════════════════════                                      │
│                                                                             │
│   Vector:                                                                   │
│   Skill returns malicious content in callback that manipulates agent      │
│                                                                             │
│   Mitigations:                                                              │
│   □ Callback data sanitization                                             │
│   □ Strict typing for callback responses                                   │
│   □ Agent prompt hardening against data injection                          │
│   □ Content validation before agent processing                             │
│                                                                             │
│   ATTACK 5: Denial of Service                                               │
│   ═══════════════════════════                                               │
│                                                                             │
│   Vector:                                                                   │
│   Skill creates infinite loop or resource exhaustion                       │
│                                                                             │
│   Mitigations:                                                              │
│   □ Execution timeout (max 5 minutes)                                      │
│   □ Memory limit per skill                                                 │
│   □ CPU throttling                                                         │
│   □ Rate limiting per tenant                                               │
│   □ Circuit breaker for repeated failures                                  │
│                                                                             │
│   ATTACK 6: LLM Manipulation                                                │
│   ══════════════════════                                                    │
│                                                                             │
│   Vector:                                                                   │
│   User crafts request to manipulate LLM skill generation                  │
│                                                                             │
│   Mitigations:                                                              │
│   □ Strict system prompts for skill generator                              │
│   □ Input sanitization before LLM                                          │
│   □ Output validation after LLM                                            │
│   □ Pattern-based generation (not free-form)                               │
│   □ Human-in-the-loop for sensitive skills                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Security Controls

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SECURITY CONTROLS                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   LAYER 1: NETWORK                                                          │
│   ════════════════                                                          │
│   □ mTLS between all services                                              │
│   □ Network policies (pod-to-pod restrictions)                             │
│   □ Egress filtering (allowlist of external APIs)                          │
│   □ WAF at API gateway                                                     │
│   □ DDoS protection                                                        │
│                                                                             │
│   LAYER 2: APPLICATION                                                      │
│   ════════════════════                                                      │
│   □ Input validation on all endpoints                                      │
│   □ Output encoding                                                        │
│   □ Rate limiting (per-user, per-tenant)                                   │
│   □ Request signing                                                        │
│   □ OWASP Top 10 protections                                               │
│                                                                             │
│   LAYER 3: DATA                                                             │
│   ═════════════                                                             │
│   □ Encryption at rest (AES-256)                                           │
│   □ Encryption in transit (TLS 1.3)                                        │
│   □ Key rotation (90 days)                                                 │
│   □ Database encryption                                                    │
│   □ Credential encryption (separate keys per tenant)                       │
│                                                                             │
│   LAYER 4: IDENTITY                                                         │
│   ═════════════════                                                         │
│   □ OAuth 2.0 / OpenID Connect                                             │
│   □ MFA for admin access                                                   │
│   □ API key rotation                                                       │
│   □ Least privilege access                                                 │
│   □ Regular access reviews                                                 │
│                                                                             │
│   LAYER 5: AUDIT                                                            │
│   ══════════════                                                            │
│   □ All actions logged                                                     │
│   □ Tamper-proof logs                                                      │
│   □ Anomaly detection                                                      │
│   □ Real-time alerting                                                     │
│   □ 90-day retention (1 year for compliance)                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Skill Sandbox

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SKILL SANDBOX                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                       SANDBOX ENVIRONMENT                            │  │
│   │                                                                       │  │
│   │   ┌───────────────────────────────────────────────────────────────┐ │  │
│   │   │                     SKILL EXECUTION                            │ │  │
│   │   │                                                                │ │  │
│   │   │   Restrictions:                                                │ │  │
│   │   │   ✗ No file system access (except /tmp)                       │ │  │
│   │   │   ✗ No subprocess execution                                   │ │  │
│   │   │   ✗ No network access except allowlisted APIs                 │ │  │
│   │   │   ✗ No access to environment variables                        │ │  │
│   │   │   ✗ No system calls                                           │ │  │
│   │   │                                                                │ │  │
│   │   │   Limits:                                                      │ │  │
│   │   │   • Memory: 512MB max                                         │ │  │
│   │   │   • CPU: 0.5 cores                                            │ │  │
│   │   │   • Time: 5 minutes max                                       │ │  │
│   │   │   • Temp storage: 100MB                                       │ │  │
│   │   │                                                                │ │  │
│   │   │   Allowlist:                                                   │ │  │
│   │   │   ✓ Standard library (limited)                                │ │  │
│   │   │   ✓ requests (to allowlisted hosts only)                      │ │  │
│   │   │   ✓ Provider SDKs (hubspot, slack, etc.)                      │ │  │
│   │   │   ✓ sorx runtime library                                      │ │  │
│   │   │                                                                │ │  │
│   │   └───────────────────────────────────────────────────────────────┘ │  │
│   │                                                                       │  │
│   │   ┌───────────────────────────────────────────────────────────────┐ │  │
│   │   │                    NETWORK EGRESS FILTER                       │ │  │
│   │   │                                                                │ │  │
│   │   │   ✓ api.hubapi.com                                            │ │  │
│   │   │   ✓ slack.com, api.slack.com                                  │ │  │
│   │   │   ✓ api.notion.com                                            │ │  │
│   │   │   ✓ api.clickup.com                                           │ │  │
│   │   │   ✓ gmail.googleapis.com                                      │ │  │
│   │   │   ✗ * (all other hosts blocked)                               │ │  │
│   │   │                                                                │ │  │
│   │   └───────────────────────────────────────────────────────────────┘ │  │
│   │                                                                       │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## External API Gateway

### API for External LLMs

Any LLM or application can use Sorx 2.0 as a skill execution layer:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SORX API GATEWAY                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   BASE URL: https://api.sorx.io/v1                                          │
│                                                                             │
│   AUTHENTICATION:                                                           │
│   ───────────────                                                           │
│   Header: Authorization: Bearer <api_key>                                   │
│   Or: X-API-Key: <api_key>                                                  │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 1: Natural Language Request                                      │
│   ═════════════════════════════════════                                     │
│                                                                             │
│   POST /request                                                             │
│                                                                             │
│   Request:                                                                  │
│   {                                                                         │
│     "request": "Send an email to john@example.com saying the invoice is due",│
│     "context": {                                                            │
│       "user_id": "user_123",                                               │
│       "conversation_id": "conv_456"                                        │
│     },                                                                      │
│     "options": {                                                            │
│       "auto_execute": true,        // Execute immediately if skill exists  │
│       "require_approval": false,   // Need user approval for new skills    │
│       "timeout_ms": 30000                                                  │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "request_id": "req_789",                                               │
│     "status": "completed",                                                 │
│     "skill_used": "gmail_send_email_v3",                                   │
│     "skill_was_generated": false,                                          │
│     "execution_time_ms": 1245,                                             │
│     "result": {                                                            │
│       "success": true,                                                     │
│       "message_id": "msg_abc",                                             │
│       "thread_id": "thread_def"                                            │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 2: Execute Specific Skill                                        │
│   ════════════════════════════════════                                      │
│                                                                             │
│   POST /skill/{skill_id}/execute                                            │
│                                                                             │
│   Request:                                                                  │
│   {                                                                         │
│     "params": {                                                            │
│       "to": "john@example.com",                                            │
│       "subject": "Invoice Due",                                            │
│       "body": "Your invoice is due..."                                     │
│     },                                                                      │
│     "options": {                                                            │
│       "timeout_ms": 30000,                                                 │
│       "async": false                                                       │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "execution_id": "exec_123",                                            │
│     "status": "completed",                                                 │
│     "result": { ... }                                                      │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 3: Discover Skills                                               │
│   ═══════════════════════════                                               │
│                                                                             │
│   POST /skill/discover                                                      │
│                                                                             │
│   Request:                                                                  │
│   {                                                                         │
│     "task": "create a task in my project management tool",                 │
│     "filters": {                                                           │
│       "providers": ["clickup", "asana", "notion"],                        │
│       "min_success_rate": 0.9                                              │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "skills": [                                                            │
│       {                                                                    │
│         "id": "clickup_create_task_v4",                                   │
│         "name": "Create ClickUp Task",                                    │
│         "match_score": 0.95,                                              │
│         "success_rate": 0.98,                                             │
│         "required_params": ["list_id", "name"],                           │
│         "optional_params": ["description", "assignee", "due_date"]        │
│       },                                                                   │
│       {                                                                    │
│         "id": "notion_create_page_v2",                                    │
│         "name": "Create Notion Page",                                     │
│         "match_score": 0.82,                                              │
│         ...                                                                │
│       }                                                                    │
│     ],                                                                      │
│     "can_generate": true,                                                  │
│     "generation_confidence": 0.9                                           │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 4: Generate New Skill                                            │
│   ══════════════════════════════                                            │
│                                                                             │
│   POST /skill/generate                                                      │
│                                                                             │
│   Request:                                                                  │
│   {                                                                         │
│     "task": "Send a Slack message to a channel when a HubSpot deal closes",│
│     "provider_hints": ["slack", "hubspot"],                                │
│     "test_after_generation": true                                          │
│   }                                                                         │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "skill_id": "hubspot_deal_close_to_slack_v1",                          │
│     "status": "generated",                                                 │
│     "test_result": {                                                       │
│       "passed": true,                                                      │
│       "test_execution_id": "test_456"                                      │
│     },                                                                      │
│     "skill": {                                                             │
│       "id": "hubspot_deal_close_to_slack_v1",                             │
│       "name": "Notify Slack on HubSpot Deal Close",                       │
│       "description": "...",                                                │
│       "required_params": ["channel_id"],                                   │
│       "credentials_needed": ["hubspot", "slack"]                           │
│     }                                                                       │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 5: Skill Library                                                 │
│   ═════════════════════════                                                 │
│                                                                             │
│   GET /skill/library                                                        │
│                                                                             │
│   Query params:                                                             │
│   ?category=crm                                                            │
│   ?provider=hubspot                                                        │
│   ?role=sales                                                              │
│   ?search=create+contact                                                   │
│   ?page=1&limit=20                                                         │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "skills": [ ... ],                                                     │
│     "total": 156,                                                          │
│     "page": 1,                                                             │
│     "limit": 20                                                            │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ENDPOINT 6: Credential Status                                             │
│   ═════════════════════════════                                             │
│                                                                             │
│   GET /credentials                                                          │
│                                                                             │
│   Response:                                                                 │
│   {                                                                         │
│     "credentials": [                                                       │
│       {                                                                    │
│         "provider": "hubspot",                                            │
│         "status": "active",                                               │
│         "last_used": "2024-01-15T10:30:00Z",                              │
│         "expires_at": null                                                 │
│       },                                                                   │
│       {                                                                    │
│         "provider": "slack",                                              │
│         "status": "active",                                               │
│         "last_used": "2024-01-15T09:15:00Z",                              │
│         "expires_at": "2024-02-15T00:00:00Z"                              │
│       }                                                                    │
│     ]                                                                       │
│   }                                                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### OpenAI Function Calling Integration

```json
{
  "name": "sorx_execute",
  "description": "Execute a skill or natural language request via Sorx 2.0",
  "parameters": {
    "type": "object",
    "properties": {
      "request": {
        "type": "string",
        "description": "Natural language description of what to do"
      },
      "skill_id": {
        "type": "string",
        "description": "Specific skill ID if known"
      },
      "params": {
        "type": "object",
        "description": "Parameters for the skill"
      }
    },
    "required": ["request"]
  }
}
```

### Claude Tool Use Integration

```python
# For Claude Agent SDK integration

SORX_TOOL = {
    "name": "sorx",
    "description": """
Execute skills via Sorx 2.0 - the universal integration framework.
Use this to:
- Send emails, Slack messages
- Create tasks in ClickUp, Asana, Notion
- Look up CRM data in HubSpot, Salesforce
- Create invoices in QuickBooks
- Any other business workflow automation
    """,
    "input_schema": {
        "type": "object",
        "properties": {
            "action": {
                "type": "string",
                "description": "What you want to do in natural language"
            },
            "skill_id": {
                "type": "string",
                "description": "Specific skill ID if you know it"
            },
            "params": {
                "type": "object",
                "description": "Parameters for the skill"
            }
        },
        "required": ["action"]
    }
}
```

---

## Implementation Priority

### Phase 1: Core Engine (CRITICAL)

1. Sorx Engine (Python sidecar)
2. Credential Vault
3. REST API Adapter
4. Basic skill generation
5. Skill storage and retrieval

### Phase 2: Provider Patterns (HIGH)

Priority based on planned integrations:
1. Google (Gmail, Drive, Calendar)
2. Slack
3. Notion
4. HubSpot
5. ClickUp, Asana

### Phase 3: Skill Learning (HIGH)

1. Skill versioning
2. Evolution based on feedback
3. Success tracking
4. Performance metrics

### Phase 4: Universal Adapters (MEDIUM)

1. Database adapter
2. Legacy system adapter (SOAP)
3. Desktop automation adapter
4. File system adapter

### Phase 5: Skill Library (MEDIUM)

1. Skill publishing
2. Discovery/search
3. Installation
4. Ratings/reviews

### Phase 6: Enterprise Features (MEDIUM)

1. SOC 2 compliance
2. Audit logging
3. Access control
4. Governance workflows

---

## Summary

**Sorx 2.0** is not just an integration platform - it's a **skill-based learning system** where AI agents:

1. **Learn** new skills on-demand by generating code from patterns
2. **Execute** skills locally with secure credential access
3. **Improve** skills over time based on feedback and experience
4. **Share** mastered skills with other agents and organizations
5. **Connect** to ANYTHING - APIs, databases, legacy systems, desktop apps, hardware

This is the **USB-C of AI integrations** - one universal interface that works with any system, learns from experience, and gets better over time.

---

## Related Documents

- `INTEGRATION_IMPLEMENTATION_PLAN.md` - Traditional integrations for data sync
- `INTEGRATION_INFRASTRUCTURE.md` - Backend architecture patterns
- `INTEGRATIONS_MASTER_LIST.md` - Complete provider inventory

---

**Sorx 2.0: System of Reasoning - Where AI Agents Learn to Connect**
