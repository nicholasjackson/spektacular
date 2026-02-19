**Date:** 2026-02-19
**Status:** Idea / Research
**Tags:** #idea #ai #agents #development #tooling

## The Concept

Spectacular is a **multi-surface orchestration platform** that takes a markdown specification as input and orchestrates AI coding agents through a structured pipeline: **analyse → plan → execute → validate**. It intelligently routes work to different models based on complexity, optimising cost without sacrificing quality.

**Key insight:** Spectacular is the **control plane for AI coding at organizational scale**, not skills that run inside individual agents. Like Kubernetes orchestrates containers, Spectacular orchestrates AI coding agents across multiple platforms, providing centralized routing, cost optimization, policy enforcement, and governance for entire development teams.

### Core Flow

```
spec.md → [Analyser] → complexity score → [Planner] → plan.md → [Executor] → code → [Validator] → result
              ↑                                ↑                      ↑                    ↑
           cheap model              smart model (scaled)        coding agent          validation agent
```

1. **Ingest** — Parse markdown spec, extract requirements, constraints, acceptance criteria
2. **Analyse** — Evaluate complexity (scope, dependencies, ambiguity, risk)
3. **Plan** — Generate implementation plan with task breakdown
4. **Execute** — Hand tasks to coding agent(s) for implementation
5. **Validate** — Verify output against spec, run tests, check acceptance criteria

### Key Differentiator: Intelligent Model Routing

The tool doesn't just throw everything at Opus/GPT-4. It:
- Uses a **cheap model** (Haiku/Flash) for spec parsing and complexity scoring
- Scales the **planning model** based on assessed complexity (Haiku for simple, Sonnet for medium, Opus for complex)
- Routes **execution** to the best coding agent for the task type
- Uses a **dedicated validation agent** to check work against the original spec

## Enterprise Control Plane Vision

**The Big Picture:** Spectacular isn't just a better way to run one coding session. It's **infrastructure for managing AI coding at enterprise scale** — a control plane that orchestrates fleets of coding agents across platforms, teams, and projects.

### Multi-Platform Agent Orchestration

```
┌─────────────────────────────────────────────────────────────┐
│                  SPECTACULAR CONTROL PLANE                  │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Complexity  │  │   Policy    │  │   Resource  │          │
│  │  Analyzer   │  │ Enforcement │  │ Allocator   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│           │              │              │                   │
│  ┌────────┴──────────────┴──────────────┴────────┐          │
│  │            ORCHESTRATION LAYER                 │         │
│  └────────────────────┬───────────────────────────┘         │
└───────────────────────┼─────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌─────────┐    ┌─────────┐    ┌─────────┐
   │MacBook Pro   │Linux Server│ │Windows VM │
   │Claude Code│   │   Aider   │ │ Cursor   │
   │$0.03/1K  │   │ $0.01/1K  │ │$0.02/1K  │
   │Available │   │   Busy    │ │Available │
   └─────────┘    └─────────┘    └─────────┘
```

### Enterprise Capabilities

#### **Workload Distribution**
- Spectacular receives `auth-spec.md` → complexity score: 0.8 (high)
- Evaluates available agents: Claude Code (busy), Aider (available, Linux), Cursor (available, Windows)
- Routes authentication tasks to Aider (lower cost + security focus)
- Routes UI tasks to Cursor (better at frontend)
- Coordinates dependencies: Aider finishes backend before Cursor starts frontend

#### **Policy Enforcement**
```yaml
# spectacular-enterprise.yaml
policies:
  security_review_required: 
    - complexity > 0.7
    - keywords: ["auth", "payment", "admin"]
  approved_models:
    - claude-sonnet-4-20250514  # Cost-efficient default
    - claude-opus-4-6          # Only for complex architecture
  code_standards:
    - eslint_config: "@company/strict"
    - test_coverage: 85%
    - security_scan: true
```

#### **Cost Management**
- **Global budget tracking:** $2000/month across all coding agents
- **Per-project allocation:** Frontend team: $500, Backend: $800, DevOps: $300
- **Real-time alerts:** "Project X at 90% of budget, switching to Haiku models"
- **Optimization recommendations:** "Moving auth tasks to Aider saved 40% vs Claude Code"

#### **Audit & Compliance**
- **Complete trail:** Who requested what spec, which agents worked on it, what was changed
- **Code provenance:** "Lines 45-67 generated by Claude-Sonnet on 2026-02-19, validated against spec-auth-v2.md"
- **Security attestation:** All AI-generated code passed security scans and human review
- **Regulatory compliance:** SOX/SOC2 requirements for financial services coding

### Competitive Moat

**This is what nobody else does.** Current tools are single-agent, single-platform:
- **Cursor:** Great IDE, but locked to Cursor
- **Claude Code:** Excellent agent, but one session at a time
- **GitHub Copilot:** Broad adoption, but no orchestration
- **AWS Kiro:** IDE-specific, vendor lock-in

**Spectacular operates above all of them:**
- Use your team's existing tools (Cursor, Claude Code, Aider, Copilot)
- Central orchestration with enterprise governance
- Cost optimization across the entire coding AI spend
- Policy enforcement that scales to hundreds of developers

### Enterprise Sales Story

*"Your development teams are already using AI coding tools — Cursor, Claude Code, Copilot. But you have no visibility into cost, quality, or compliance. Spectacular gives you a control plane: centralized orchestration, intelligent cost optimization, policy enforcement, and audit trails. Think Kubernetes for AI coding agents."*

**ROI calculation:**
- **Cost savings:** 40-60% through intelligent model routing
- **Productivity gains:** Coordinated agents working in parallel vs. sequential
- **Risk reduction:** Policy enforcement, security scans, compliance audit trails
- **Team scaling:** New developers get instant access to organizational coding knowledge

## Existing Landscape (Prior Art)

### GitHub Spec Kit (Sep 2025)
- **What:** Open-source toolkit for spec-driven development
- **Flow:** Specify → Plan → Tasks → Code (4 phases)
- **Agents:** Works with Copilot, Claude Code, Gemini CLI
- **Key insight:** Specs as "living, executable artifacts" not static docs
- **Limitation:** No model routing or complexity analysis — uses whatever agent you point it at
- **Repo:** github.com/github/spec-kit
- **Relevance to Spectacular:** Very similar philosophy. Spec Kit is agent-agnostic scaffolding. Spectacular could wrap or extend this with the intelligence layer (complexity scoring + model routing).

### Kiro (AWS, Jul 2025)
- **What:** Full IDE (VS Code fork) with built-in spec-driven workflow
- **Flow:** Requirements → Design → Tasks → Implementation (enforced structure)
- **Agents:** Built-in, tied to AWS/Bedrock
- **Key feature:** "Agent hooks" that auto-trigger on file save (test sync, doc update, security scan)
- **Limitation:** Locked into Kiro's IDE. Not composable. Can't use your own agents.
- **Relevance to Spectacular:** Proves the spec → design → tasks → code pipeline works. But it's a monolith. Spectacular should be the composable, agent-agnostic version.

### BMAD Method (2025)
- **What:** "Breakthrough Method for Agile AI-Driven Development"
- **Flow:** Two phases — "agentic planning" then "context-engineered development"
- **Agents:** 12+ specialised agents (Analyst, PM, Architect, Scrum Master, Developer, etc.)
- **Key insight:** Each SDLC role gets its own agent persona with specific expertise
- **Limitation:** Complex setup, lots of personas to manage, no built-in cost optimisation
- **Repo:** github.com/bmad-code-org/BMAD-METHOD
- **Relevance to Spectacular:** The role-based agent idea is solid. Spectacular could define 3-4 focused agents (Planner, Executor, Validator) rather than 12+ — keep it lean.

### ThoughtWorks Analysis (Dec 2025)
- **Key observation:** SDD separates design and implementation phases explicitly
- **Warning:** "The quality of the specification determines the quality of the output" — garbage in, garbage out still applies
- **Practical note:** Works best when combined with human review checkpoints between phases

### LLM Router / Model Routing (General Pattern)
- IBM research: routing queries to smaller models can cut inference costs by **85%**
- Pattern: cheap classifier → routes to appropriate model tier
- Already used in production by OpenRouter, Requesty, and others
- **Relevance to Spectacular:** This is the economic engine. Spec analysis is the classifier; complexity score drives model selection.

## What Makes Spectacular Different

| Feature | Spec Kit | Kiro | BMAD | **Spectacular** |
|---------|----------|------|------|-----------------|
| Agent-agnostic | ✅ | ❌ | ✅ | ✅ |
| Complexity scoring | ❌ | ❌ | Partial | ✅ |
| Model routing | ❌ | ❌ | ❌ | ✅ |
| Cost optimisation | ❌ | ❌ | ❌ | ✅ |
| Spec format | Markdown | Proprietary | Markdown | Markdown |
| Standalone CLI | ✅ | ❌ (IDE) | ✅ | ✅ |
| Validation phase | Manual | Hooks | Manual | **Automated** |
| Parallel execution | ❌ | ❌ | ❌ | ✅ (planned) |

## Architecture Sketch

### Components

```
spectacular/
├── cli/                    # CLI entry point
│   └── spectacular.ts      # Main command: spectacular run spec.md
├── core/
│   ├── parser.ts           # Markdown spec parser (extract sections, criteria)
│   ├── analyser.ts         # Complexity scoring engine
│   ├── planner.ts          # Plan generation (calls planning agent)
│   ├── executor.ts         # Task execution orchestrator
│   ├── validator.ts        # Spec compliance checker
│   └── router.ts           # Model/agent selection based on complexity
├── agents/
│   ├── profiles/           # Agent persona definitions (markdown)
│   │   ├── planner.md      # Planning agent system prompt
│   │   ├── executor.md     # Execution agent system prompt
│   │   └── validator.md    # Validation agent system prompt
│   └── adapters/           # Agent backend adapters
│       ├── claude-code.ts  # Claude Code / API
│       ├── openai.ts       # OpenAI / Codex
│       └── local.ts        # Local models (Ollama, etc.)
├── config/
│   └── models.yaml         # Model tier definitions + pricing
└── output/
    ├── plan.md             # Generated plan
    ├── tasks/              # Individual task files
    └── report.md           # Validation report
```

### Complexity Scoring

The analyser would evaluate specs on multiple dimensions:

| Dimension | Weight | Signals |
|-----------|--------|---------|
| **Scope** | 0.25 | Number of features, endpoints, components |
| **Dependencies** | 0.20 | External APIs, databases, auth, third-party services |
| **Ambiguity** | 0.20 | Vague language, missing acceptance criteria, open questions |
| **Technical risk** | 0.20 | New tech, concurrency, security requirements, perf constraints |
| **Integration** | 0.15 | Existing codebase size, coupling, migration needs |

**Score → Model tier mapping:**
- 0.0–0.3 (Simple): Haiku/Flash for planning, Sonnet for execution
- 0.3–0.6 (Medium): Sonnet for planning, Sonnet for execution
- 0.6–0.8 (Complex): Opus for planning, Sonnet for execution
- 0.8–1.0 (Critical): Opus for planning, Opus for execution + human review gates

### Spec Format (Input)

```markdown
# Feature: User Authentication

## Overview
Add OAuth2 login with Google and GitHub providers.

## Requirements
- [ ] Users can sign in with Google OAuth2
- [ ] Users can sign in with GitHub OAuth2
- [ ] Session persists across browser refreshes
- [ ] Logout clears session and revokes token

## Constraints
- Must use existing Express backend
- PostgreSQL for session storage
- No new dependencies over 50KB gzipped

## Acceptance Criteria
- [ ] Login redirects to provider, returns with valid session
- [ ] Session cookie is httpOnly, secure, sameSite=strict
- [ ] Logout returns 200 and invalidates server-side session
- [ ] Rate limit: 10 login attempts per minute per IP
```

### Agent Personas

**Planner Agent** — Senior architect mindset. Reads spec, understands constraints, produces a step-by-step implementation plan with file-level changes, dependency order, and risk flags. Doesn't write code.

**Executor Agent** — Coding agent (Claude Code, Copilot, Aider, etc.). Takes individual tasks from the plan, writes code, runs tests. Focuses on one task at a time. Gets the plan + relevant context, not the whole spec.

**Validator Agent** — QA mindset. Takes the original spec + the produced code, checks each acceptance criterion, runs any defined tests, flags gaps. Returns a structured pass/fail report with remediation suggestions.

## Open Questions

1. **Should Spectacular wrap existing tools (Claude Code, Aider) or call APIs directly?**
   - Wrapping = easier adoption, users keep their setup
   - Direct API = more control over model routing, but reinvents the wheel
   - **Hybrid?** Use Claude Code/Aider for execution, but API calls for analysis/planning/validation

2. **How to handle spec ambiguity?**
   - Option A: Fail fast — refuse to plan until ambiguity score drops below threshold
   - Option B: Generate clarifying questions, present to user, iterate
   - Option C: Make assumptions, document them in plan, let user approve
   - **Probably B for interactive, A for CI/headless mode**

3. **Parallel execution?**
   - Independent tasks could run in parallel across multiple agent instances
   - Need dependency graph from planner
   - Risk: merge conflicts, context drift
   - **v2 feature — get serial working first**

4. **State management between phases?**
   - Each phase produces a markdown artifact (plan.md, tasks/*.md, report.md)
   - These become the input to the next phase
   - Git-friendly, human-readable, diffable
   - Can checkpoint and resume

5. **How to price/package?**
   - Open source CLI + BYO API keys (like Aider)
   - Or hosted service with markup on model costs
   - **Start open source, monetise later if it gets traction**

6. **Integration with existing codebases?**
   - Spectacular needs codebase context for accurate planning
   - Could consume AGENTS.md, README, file tree
   - Maybe integrate with code search (ripgrep, tree-sitter) for context gathering

## Multi-Surface Strategy

Spectacular operates through three complementary interfaces, each serving different users and contexts:

### 1. OpenClaw Plugin (Premium Experience)
- **Canvas UI:** Full orchestration dashboard with real-time agent monitoring
- **Cost tracking:** Live model usage, budget alerts, complexity visualization
- **Voice integration:** "spectacular status", "pause execution", "validate current task"
- **Multi-channel notifications:** Telegram updates, Discord progress reports
- **Sub-agent orchestration:** Leverages OpenClaw's existing `sessions_spawn` and `subagents` management
- **MCP integration:** Uses OpenClaw's mcporter skill for tool access scoping
- **Target:** Power users, teams with OpenClaw already deployed

### 2. GitHub Issues Integration (Workflow Integration)
```bash
spectacular run --github 123  # Pulls issue #123 as spec
```
- Issue → spec parsing (title + body + comments + labels as requirements)
- Auto-creates branches, updates issue with plan links and progress
- PR creation with structured summaries and validation reports
- Labels drive complexity hints (`complexity:high` → Opus planner)
- **Target:** Existing GitHub-based teams, minimal adoption friction

### 3. CLI (Universal Access)
```bash
spectacular run spec.md  # Works anywhere
```
- Outputs markdown artifacts (plan.md, tasks.md, report.md)
- CI/CD friendly, scriptable, no UI dependencies
- Agent-agnostic — works with any coding tool
- **Target:** Individual developers, automation, environments without OpenClaw

### Unified Backend Architecture
```
GitHub Issues ──┐
                ├─→ Spectacular Core Engine ─→ Coding Agents
CLI ────────────┤   (complexity scoring,       (Claude Code,
                │    model routing,             Aider, Cursor,
OpenClaw ───────┘    validation, etc.)         etc.)
```

**Adoption Path:** CLI first (broadest compatibility) → GitHub integration (team workflow) → OpenClaw plugin (premium experience)

## Spec Kit Compatibility Strategy

Rather than reinvent the wheel, Spectacular adopts and extends GitHub Spec Kit's proven practices:

### What We Adopt from Spec Kit
- ✅ **4-phase workflow:** Specify → Plan → Tasks → Code (proven structure)
- ✅ **Markdown conventions:** Compatible spec format, section headers, acceptance criteria
- ✅ **"Living artifacts" philosophy:** Specs evolve with the project
- ✅ **Agent-agnostic approach:** Works with Claude Code, Copilot, Gemini CLI, etc.
- ✅ **File naming:** plan.md, tasks.md, context.md structure

### What Spectacular Adds (Intelligence Layer)
- 🆕 **Phase 0:** Complexity analysis before planning
- 🆕 **Model routing:** Cost optimization through intelligent agent selection
- 🆕 **Phase 5:** Automated validation against original spec  
- 🆕 **Knowledge layer:** Learnings capture, institutional memory
- 🆕 **Multi-surface:** CLI, GitHub, OpenClaw interfaces
- 🆕 **Orchestration:** Sub-agent management, parallel execution

### Compatibility Promise
```bash
# Should work on existing Spec Kit projects
spectacular run existing-spec-kit-spec.md

# Should produce Spec Kit compatible artifacts
spectacular run new-spec.md  # → generates plan.md, tasks.md

# Should enhance existing workflows
spec-kit generate spec.md     # → Standard Spec Kit planning
spectacular validate          # → Add validation layer
```

**Positioning:** Spectacular as **"Spec Kit Pro"** — same standards, enhanced intelligence. Teams can start with Spec Kit, upgrade when they need cost optimization, validation, and orchestration.

## Dogfooding Strategy: Building Spectacular with Spectacular

**The ultimate validation:** Use Spectacular to build itself.

### Meta-Workflow
1. **Bootstrap:** Build minimal CLI manually (v0.1)
2. **Self-hosting:** Write `spectacular-spec.md` for v0.2 features
3. **Recursive improvement:** Use v0.1 to orchestrate building v0.2
4. **Full stack:** Eventually every feature (GitHub integration, OpenClaw plugin) gets spec-driven development

### What This Proves
- ✅ Complexity scoring works on real software (not toy examples)
- ✅ Model routing handles actual development tasks efficiently  
- ✅ Validation catches real bugs against real acceptance criteria
- ✅ Knowledge layer accumulates genuine learnings from complex project
- ✅ Orchestration manages real sub-agents doing real work

### Pitch Value
*"We're so confident in Spectacular that we built it using itself. Every feature was spec-driven, every task was intelligently routed to the right model, every validation ran against the original requirements. The tool that built Spectacular is the same tool we're selling to you."*

### Practical Benefits
- Immediate feedback loop on UX pain points
- Real performance data (costs, time savings, accuracy)
- Authentic case study with concrete metrics
- Forces solving real problems, not imaginary ones
- Compelling "tool that builds itself" narrative for developer mindshare

## Implementation Roadmap

### v0.1 — Bootstrap CLI (Manual Build)
- [ ] Minimal CLI: `spectacular run spec.md`
- [ ] Spec Kit compatible markdown parser
- [ ] Basic complexity scorer (heuristic rules)
- [ ] Single-model planning (Claude API with model selection)
- [ ] Outputs: plan.md, tasks.md, context.md (Spec Kit compatible)
- [ ] Manual execution documentation (user runs with their coding agent)
- [ ] **Meta-goal:** Build enough to run `spectacular run spectacular-v0.2-spec.md`

### v0.2 — Self-Hosted Development (Built with v0.1)
- [ ] **Spec:** Write `spectacular-v0.2-spec.md` with execution features
- [ ] **Use v0.1:** Let Spectacular orchestrate its own development
- [ ] Automated execution via coding agent subprocess (Claude Code, Aider)
- [ ] Basic validation (acceptance criteria checker)
- [ ] Model routing engine (complexity → Haiku/Sonnet/Opus)
- [ ] Knowledge layer: `.spectacular/knowledge/` with learnings capture
- [ ] **Meta-goal:** Prove the workflow works on real development

### v0.3 — Multi-Surface (Built with v0.2)
- [ ] **GitHub integration:** `spectacular run --github 123`
- [ ] Issue parsing, branch creation, PR summaries
- [ ] **OpenClaw plugin foundation:** MCP server interface
- [ ] Canvas UI prototype for status/progress visualization
- [ ] Multiple coding agent backends (Claude Code, Aider, Cursor adapters)
- [ ] Validation with remediation loop (fail → re-execute → validate)

### v0.4 — Control Plane Foundation (Built with v0.3)
- [ ] **Multi-platform agent registry:** Discover and manage Claude Code, Aider, Cursor across network
- [ ] **Workload distribution:** Route tasks to available agents based on capability and cost
- [ ] **OpenClaw Canvas UI:** Real-time fleet management dashboard
- [ ] Resource allocation and parallel task coordination
- [ ] Basic policy enforcement (model restrictions, cost budgets)
- [ ] Cross-platform agent lifecycle management

### v0.5 — Enterprise Orchestration (Built with v0.4)
- [ ] **Fleet-wide cost optimization:** Global budget tracking, per-project allocation
- [ ] **Policy engine:** Security reviews, code standards, compliance rules
- [ ] **Audit trail:** Complete provenance tracking for regulatory compliance
- [ ] **Performance analytics:** Agent utilization, cost per feature, quality metrics
- [ ] Multi-tenant architecture for enterprise deployment
- [ ] SSO integration, RBAC, team management

### v1.0 — Production Control Plane (Built with v0.5)
- [ ] **Enterprise marketplace:** Shareable agent configurations, policy templates
- [ ] **API gateway:** Third-party integrations, custom agent plugins
- [ ] **Global deployment:** Multi-region, high availability, disaster recovery
- [ ] **Advanced orchestration:** ML-based task routing, predictive scaling
- [ ] **Compliance frameworks:** SOX, SOC2, GDPR attestation reports
- [ ] **Meta-achievement:** Entire enterprise platform built and maintained using its own control plane

### Dogfooding Milestones
- **v0.1 → v0.2:** First recursive build proves core concept
- **v0.2 → v0.3:** Multi-feature orchestration validates complexity
- **v0.3 → v0.4:** Enterprise features built with own tooling
- **v0.4 → v1.0:** Full platform self-development demonstrates scalability

## Technical Decisions for Control Plane Architecture

### Core Platform
- **Language:** TypeScript (Node.js ecosystem, async-native, good for both CLI and web APIs) vs Rust (performance, but smaller AI tooling ecosystem) vs Go (excellent for distributed systems, microservices)
- **Architecture:** Monolithic (easier to start) vs Microservices (better for enterprise scale) vs Hybrid (core monolith + plugin microservices)
- **Database:** PostgreSQL (relational, audit trails) vs MongoDB (flexibility) vs Redis (performance) for agent state, job queues, audit logs

### Agent Communication & Discovery  
- **Agent discovery:** mDNS/Bonjour (local network) vs service registry (Consul/etcd) vs database-backed registry
- **Agent communication:** HTTP REST APIs vs gRPC vs WebSockets vs MCP (Model Context Protocol)
- **Message queuing:** Redis pub/sub vs RabbitMQ vs Apache Kafka for task distribution
- **Load balancing:** Round-robin vs capability-based vs cost-optimized routing

### Enterprise Features
- **Authentication:** JWT tokens vs OAuth2/OIDC vs SAML for enterprise SSO
- **Authorization:** RBAC (role-based) vs ABAC (attribute-based) vs custom policy engine
- **Audit logging:** Structured JSON logs vs dedicated audit database vs blockchain for immutable trails
- **Multi-tenancy:** Database-per-tenant vs schema-per-tenant vs row-level security

### Deployment & Operations
- **Containerization:** Docker + Kubernetes vs native binaries vs serverless (Lambda/Cloud Functions)
- **Configuration:** YAML files vs environment variables vs distributed config (etcd/Consul) vs database
- **Monitoring:** Prometheus/Grafana vs ELK stack vs cloud-native (CloudWatch/Azure Monitor)
- **High availability:** Active-passive vs active-active vs distributed consensus (Raft)

### Development Approach
- **MVP Strategy:** CLI-first (broad compatibility) vs OpenClaw plugin (faster enterprise features) vs web-first (universal access)
- **Testing:** Unit tests + integration tests vs end-to-end via dogfooding vs chaos engineering for distributed components
- **Deployment:** Self-hosted (enterprise control) vs SaaS (easier adoption) vs hybrid (local agents, cloud control plane)

## Name Ideas (backup)
- Spectacular ← **current favourite** (spec + spectacular, memorable)
- Specflow (taken by .NET testing framework)
- Spectre (cool but edgy)
- Specular (mirror metaphor — spec reflects in code)

## Expanded Vision: Virtual Agent Environment

The original concept focused on the pipeline (spec → plan → execute → validate). But thinking bigger, Spectacular should be a **virtual agent environment** — a managed workspace where agents operate with proper tooling, knowledge, and governance.

### Sub-Agent Management

The core problem: when you spawn sub-agents for parallel work, who manages them? How do they share context without stepping on each other?

**Spectacular as orchestrator:**
- Spawns sub-agents with scoped permissions and context windows
- Each agent gets: its task, relevant project knowledge, and tool access — nothing more
- Central coordinator tracks progress, detects conflicts, handles failures
- Agents report back structured results, not raw chat output

**Patterns from the wild:**
- **Claude Code sub-agents** (native, Jul 2025): markdown frontmatter defines agent name, tools, model — Claude auto-delegates or you invoke explicitly. Lightweight but no orchestration layer.
- **claude-flow** (github.com/ruvnet/claude-flow): multi-agent swarms with knowledge graph (PageRank + Jaccard), MCP integration. Heavy but proves the pattern works.
- **Agentrooms** (claudecode.run): @mention routing to specialised agents in shared workspace. Good UX metaphor.
- **VoltAgent/awesome-claude-code-subagents**: 100+ pre-built sub-agents. Shows the ecosystem wants composable, specialised agents.

**Spectacular's approach:**
- Define agent roles in markdown (like Claude Code sub-agents)
- But add an **orchestration layer** that manages lifecycle, context isolation, and result aggregation
- Agents are stateless between tasks — all state lives in the project docs
- Think Kubernetes pods but for AI agents: spawn, run, collect, kill

### Project Knowledge Management

Every coding agent session starts cold. They don't know your architecture, conventions, past decisions, or gotchas. The `.docs/knowledge/` pattern from IW is the right idea.

**Spectacular's knowledge layer:**

```
.spectacular/
├── knowledge/
│   ├── architecture.md      # System architecture, patterns, decisions
│   ├── conventions.md       # Code style, naming, file structure rules
│   ├── gotchas.md          # Known issues, workarounds, landmines
│   ├── learnings/          # Auto-captured corrections from past runs
│   │   ├── 2026-02-19.md
│   │   └── ...
│   └── dependencies.md     # External services, APIs, constraints
├── agents/
│   ├── planner.md          # Planner persona + instructions
│   ├── executor.md         # Executor persona + instructions
│   ├── validator.md        # Validator persona + instructions
│   └── custom/             # User-defined specialist agents
├── specs/                  # Active and completed specs
│   ├── active/
│   └── archive/
├── plans/                  # Generated plans with task breakdowns
├── config.yaml             # Model routing, MCP servers, budgets
└── history/                # Run logs, cost tracking
```

**Key principle:** Agents don't just consume knowledge — they **produce** it. Every correction, every deviation from plan, every "oh that didn't work because X" gets captured in `learnings/`. Future runs search this before planning. This is what IW calls "continuous learning" and it's genuinely the killer feature.

**Knowledge injection per phase:**
- **Analyser**: gets architecture.md + dependencies.md (to assess complexity accurately)
- **Planner**: gets everything in knowledge/ (to avoid known pitfalls)
- **Executor**: gets conventions.md + gotchas.md + relevant learnings (focused context)
- **Validator**: gets the spec + conventions.md (to check compliance)

### MCP Server Integration

MCP (Model Context Protocol) is now the standard way agents interact with external tools. Spectacular should be a first-class MCP citizen.

**Two directions:**

#### 1. Spectacular consumes MCP servers (gives agents tools)
```yaml
# config.yaml
mcp_servers:
  - name: github
    command: npx @modelcontextprotocol/server-github
    env:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
  - name: postgres
    command: npx @modelcontextprotocol/server-postgres
    args: ["postgresql://localhost/mydb"]
  - name: filesystem
    command: npx @modelcontextprotocol/server-filesystem
    args: ["/path/to/project"]

# Per-agent tool access
agents:
  planner:
    mcp_tools: [github.search_issues, github.get_file]
  executor:
    mcp_tools: [filesystem.*, github.create_branch]
  validator:
    mcp_tools: [filesystem.read, postgres.query]
```

Each agent gets **scoped MCP tool access** — the planner can read GitHub issues but can't write files. The executor can write files but can't merge PRs. Principle of least privilege.

#### 2. Spectacular exposes itself as an MCP server
Other tools (Claude Code, VS Code, etc.) could invoke Spectacular via MCP:
```
spectacular.analyse_spec → returns complexity score
spectacular.create_plan → returns plan.md
spectacular.execute_plan → runs execution pipeline
spectacular.validate → runs validation
```

This means you could use Spectacular from inside any MCP-compatible tool without leaving your IDE.

### Issue Tracker Integration (GitHub/GitLab/Local)

Specs shouldn't only live as local markdown files. They should flow from wherever work is tracked.

**Three modes:**

#### Local Markdown (default)
```
spectacular run spec.md
```
Pure filesystem. Specs, plans, tasks all in `.spectacular/`. Good for solo work, prototyping.

#### GitHub Issues
```
spectacular run --github 123
```
- Pulls issue #123 as the spec (title + body + comments + labels)
- Creates a branch automatically
- Updates issue with plan link, progress comments
- Creates PR on completion with structured summary
- Maps issue labels → complexity hints (e.g., `complexity:high` → Opus planner)

#### GitLab Issues
```
spectacular run --gitlab 456
```
Same pattern, GitLab API.

**The IW model here is excellent:** their `/iw-plan 123` pulls a GitHub issue, researches the codebase, creates a plan in `.docs/issues/123/`, then `/iw-implement 123` executes it with worktree isolation and phase-based commits. Spectacular should adopt this exact flow but add the intelligence layer on top.

### Best Practices Engine

Rather than hard-coding best practices, Spectacular should **learn** them from the project and make them enforceable.

**Auto-detected practices:**
- Scan existing codebase for patterns (test file locations, import conventions, naming)
- Read AGENTS.md, .editorconfig, ESLint/Prettier configs, CI files
- Build a conventions profile automatically on first run

**Enforceable at each phase:**
- **Planning**: "This project uses Jest for testing — plan should include test files"
- **Execution**: "This project uses absolute imports — don't generate relative imports"
- **Validation**: "This project requires 80% coverage — check coverage report"

**Community practices (future):**
- Shareable practice packs: "TypeScript/React best practices", "Go microservices patterns"
- Think ESLint configs but for agent behaviour

## Process Model: Lessons from IW (jumppad-labs/iw)

The Implementation Workflow (IW) by Jumppad Labs is the closest thing to what Spectacular's process should look like. It's a Claude Code skill set that enforces:

### What IW Gets Right
1. **Research before planning** — launches parallel research agents to explore the codebase before writing a plan. Not just "read the spec and go."
2. **Structured artifacts** — every plan produces: `plan.md`, `tasks.md`, `context.md`, `research.md`. Human-readable, diffable, resumable.
3. **Phase-based execution** — implements task-by-task with commits after each phase. User confirms at milestones. Not one big YOLO commit.
4. **Learnings capture** — corrections and discoveries get stored in `.docs/knowledge/learnings/` for future sessions. Institutional memory.
5. **Git workflow** — creates worktrees for isolation, phase commits, automatic PR creation. Clean git history by default.
6. **Obsidian integration** — research reports can go straight to your knowledge vault.

### What IW Doesn't Do (Spectacular's additions)
1. **No complexity analysis** — every task gets the same treatment regardless of difficulty
2. **No model routing** — uses whatever model Claude Code is configured with
3. **No cost awareness** — no budget tracking or optimisation
4. **No validation agent** — relies on human review and tests
5. **No parallel execution** — serial only
6. **Single agent backend** — Claude Code only, not agent-agnostic
7. **No MCP server integration** — limited to Claude Code's built-in tools

### Spectacular = IW's Process + Intelligence Layer

The ideal: take IW's proven workflow structure and layer on:
- Complexity-driven model selection
- Multi-agent backend support
- Automated validation
- MCP tool scoping
- Cost tracking and budget enforcement
- Parallel task execution (when safe)

```
┌─────────────────────────────────────────────────────────┐
│                    SPECTACULAR                            │
│                                                          │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐            │
│  │ Analyser │──▶│ Planner  │──▶│ Executor │──┐         │
│  │ (Haiku)  │   │ (Sonnet/ │   │ (Claude/ │  │         │
│  │          │   │  Opus)   │   │  Aider)  │  │         │
│  └──────────┘   └──────────┘   └──────────┘  │         │
│       │              │              │         ▼         │
│       │              │              │    ┌──────────┐   │
│       │              │              │    │Validator │   │
│       │              │              │    │ (Sonnet) │   │
│       │              │              │    └────┬─────┘   │
│       │              │              │         │         │
│  ┌────┴──────────────┴──────────────┴─────────┴──┐     │
│  │            Knowledge Layer                      │     │
│  │  architecture │ conventions │ learnings │ gotchas│     │
│  └─────────────────────────────────────────────────┘     │
│       │              │              │         │         │
│  ┌────┴──────────────┴──────────────┴─────────┴──┐     │
│  │              MCP Tool Layer                     │     │
│  │  github │ filesystem │ postgres │ custom        │     │
│  └─────────────────────────────────────────────────┘     │
│       │                                                  │
│  ┌────┴──────────────────────────────────────────┐      │
│  │           Issue Tracker Adapter                │      │
│  │  local markdown │ github │ gitlab │ linear     │      │
│  └───────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────┘
```

## Updated Roadmap

### v0.1 — IW-Inspired Foundation
- [ ] CLI: `spectacular init` (creates `.spectacular/` structure)
- [ ] CLI: `spectacular run spec.md` (local markdown mode)
- [ ] Complexity analyser (heuristic scoring)
- [ ] Planner with model routing (Haiku/Sonnet/Opus based on score)
- [ ] Structured output: plan.md, tasks.md, context.md
- [ ] Knowledge directory with learnings capture
- [ ] Single agent backend (Claude API)

### v0.2 — Execution + Validation
- [ ] Automated execution via Claude Code subprocess
- [ ] Phase-based commits with progress tracking
- [ ] Validator agent with structured pass/fail
- [ ] Remediation loop (fail → re-execute)
- [ ] Git workflow (branch, worktree, PR)
- [ ] GitHub issue integration (`spectacular run --github 123`)

### v0.3 — Agent Environment
- [ ] MCP server consumption (scoped per agent)
- [ ] Spectacular as MCP server (expose pipeline to other tools)
- [ ] Multiple agent backends (Claude, OpenAI, Ollama)
- [ ] Sub-agent spawning with context isolation
- [ ] Cost tracking and budget enforcement
- [ ] Best practices auto-detection

### v1.0 — Production
- [ ] Parallel task execution with conflict detection
- [ ] GitLab + Linear integration
- [ ] Community practice packs
- [ ] Plugin system for custom agents/validators/adapters
- [ ] CI integration (run in pipelines)
- [ ] Dashboard/reporting (web UI or Obsidian integration)

## References

- [GitHub Spec Kit](https://github.com/github/spec-kit) — Open source SDD toolkit
- [Kiro IDE](https://kiro.dev/) — AWS spec-driven IDE
- [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD) — Multi-agent SDLC framework
- [Spec-Driven Development (arxiv)](https://arxiv.org/html/2602.00180v1) — Academic paper on SDD
- [Addy Osmani — How to Write Good Specs](https://addyosmani.com/blog/good-spec/)
- [ThoughtWorks on SDD](https://www.thoughtworks.com/en-us/insights/blog/agile-engineering-practices/spec-driven-development-unpacking-2025-new-engineering-practices)
- [LLM Routing for Cost Reduction](https://www.requesty.ai/blog/intelligent-llm-routing-in-enterprise-ai-uptime-cost-efficiency-and-model)
- [IW — Implementation Workflow](https://github.com/jumppad-labs/iw) — Claude Code skills for plan → execute → learn cycle
- [Claude Code Sub-Agents](https://code.claude.com/docs/en/sub-agents) — Native sub-agent support
- [claude-flow](https://github.com/ruvnet/claude-flow) — Multi-agent swarm orchestration with knowledge graph
- [mcp-agent](https://github.com/lastmile-ai/mcp-agent) — Build agents using MCP + workflow patterns
- [Anthropic: Code Execution with MCP](https://www.anthropic.com/engineering/code-execution-with-mcp)
- [awesome-claude-code-subagents](https://github.com/VoltAgent/awesome-claude-code-subagents) — 100+ specialised sub-agents
- [awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code) — Skills, hooks, orchestrators index
