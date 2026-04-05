---
name: reviewing-multi-agent-architecture
description: >
  Review and evaluate multi-agent system architectures for correctness, resilience, and
  operational readiness. Use when the user asks to review an agent architecture, evaluate
  an agentic system design, assess agent orchestration patterns, check guardian/safety
  layers, or mentions multi-agent, agent swarm, orchestrator-worker, or agent hierarchy.
  Covers layer separation, failure modes, identity persistence, escalation paths, and
  self-improvement boundaries.
license: MIT
metadata:
  author: jverhoeks
  version: "1.0.0"
  team: data-ai
---

# Reviewing Multi-Agent Architecture

Structured review of multi-agent system designs against a proven three-layer reference
architecture. Produces a written assessment with severity-ranked findings and actionable
recommendations.

## When to Use

- User shares an agent architecture diagram, spec, or codebase and asks for review
- User is designing a new multi-agent system and wants guidance
- User wants to validate orchestration, safety, or identity patterns
- User mentions terms like: orchestrator, sub-agent, guardian, agent swarm, tool-use
  chain, soul, agent memory, escalation, human-in-the-loop

## Reference Architecture — Three Layers

Evaluate every design against these three layers. A layer may be implicit, but its
*responsibilities* must exist somewhere.

### Layer 1: Guardian / Control Plane

**Purpose:** Safety envelope, policy enforcement, budget controls, human escalation.

Responsibilities:
- Global policy enforcement (what agents may never do)
- Token / cost / time budget caps per task and per session
- Circuit breakers: halt execution on anomaly detection
- Human-in-the-loop escalation triggers
- Audit logging of all inter-agent decisions
- Rate limiting and concurrency control

Red flags if missing:
- No global stop mechanism
- Agents can spend unbounded tokens/money
- No audit trail of agent-to-agent delegation
- Escalation path undefined or circular

### Layer 2: Orchestration / Control Flow

**Purpose:** Task decomposition, delegation, result aggregation, retry logic.

Responsibilities:
- Break complex goals into discrete tasks with acceptance criteria
- Select or spawn appropriate worker agents
- Route tasks based on capability matching
- Aggregate and verify results from workers
- Retry with context on failure (max 2-3 retries, then escalate)
- Maintain task DAG / dependency graph
- Detect and prevent circular delegation

Red flags if missing:
- Flat agent topology with no coordination
- No task dependency tracking
- Retry logic without backoff or escalation
- Single point of failure in orchestrator
- No verification of sub-agent output

### Layer 3: Worker Agents / Execution

**Purpose:** Perform specific tasks using tools, produce artifacts, report status.

Responsibilities:
- Execute scoped tasks with clear inputs and expected outputs
- Use tools (code execution, search, API calls) within granted permissions
- Report structured status (success / failure / blocked / needs-input)
- Operate within granted tool and data permissions only
- Maintain task-local context, not global state

Red flags if missing:
- Agents with unbounded tool access
- No structured output contract
- Agents modifying shared state without coordination
- No distinction between agent capabilities

## Review Procedure

Follow these steps in order. Produce findings as you go.

### Step 1: Identify the Layers

Map the user's architecture onto the three-layer model:
1. List every agent, service, or component
2. Classify each into Guardian, Orchestration, or Worker
3. Note any component that spans multiple layers (a smell)
4. Note any layer with zero components (a gap)

Output a layer mapping table:

```
| Component         | Layer          | Notes                    |
|-------------------|----------------|--------------------------|
| PolicyGuard       | Guardian       | Enforces content policy  |
| TaskRouter        | Orchestration  | Decomposes user requests |
| CodeWriter        | Worker         | Generates code           |
```

### Step 2: Evaluate Each Layer

For each layer, check responsibilities against the reference list above.
Score each responsibility: present, partial, or missing.

### Step 3: Assess Cross-Cutting Concerns

Check these concerns that span all layers:

**Identity and Memory**
- Do agents have persistent identity across sessions? (soul / profile)
- Is memory scoped correctly? (task-local vs agent-local vs global)
- Can agents modify their own identity or instructions? If yes, what guards exist?
- Is there generational evolution? If so, what prevents drift?

**Failure Modes**
- What happens when a worker agent fails?
- What happens when the orchestrator fails?
- What happens when the guardian itself fails? (fail-open vs fail-closed)
- Are there deadlock scenarios in the task DAG?
- Is there a maximum cascade depth?

**Security Boundaries**
- Can a worker agent escalate its own permissions?
- Are tool grants scoped per-task or per-agent?
- Is prompt injection addressed at agent boundaries?
- Are inter-agent messages validated?

**Observability**
- Is there a trace ID across the full agent chain?
- Can a human reconstruct what happened and why?
- Are token costs tracked per agent and per task?

### Step 4: Check Self-Improvement Boundaries

If the system includes any form of self-improvement (soul.md loops, prompt tuning,
memory consolidation, capability expansion):

- Is there a diff-based review of changes? (before/after comparison)
- Is there a rollback mechanism?
- Are improvement cycles rate-limited?
- Is there a human approval gate for capability changes?
- Can an agent modify another agent's instructions?
- What prevents reward hacking or metric gaming?

Rate the self-improvement maturity:
- **Level 0:** No self-improvement
- **Level 1:** Memory accumulation only (append-only context)
- **Level 2:** Prompt/config tuning with human review
- **Level 3:** Autonomous prompt modification with guardrails
- **Level 4:** Autonomous capability expansion (high risk — requires strong Guardian)

### Step 5: Produce the Assessment

Structure your output as follows:

```markdown
## Architecture Review: [System Name]

### Layer Mapping
[table from Step 1]

### Findings

#### Critical (must fix before production)
- [C1] ...

#### Warning (should fix, creates risk)
- [W1] ...

#### Advisory (best practice, nice to have)
- [A1] ...

### Self-Improvement Maturity: Level [0-4]
[rationale]

### Recommendations
1. [Highest priority first]
2. ...

### Architecture Diagram Suggestion
[If helpful, suggest a revised component diagram]
```

### Severity Definitions

| Severity | Meaning |
|----------|---------|
| Critical | System can cause harm, lose data, or spend unbounded resources |
| Warning  | System will degrade under stress, edge cases unhandled |
| Advisory | Improvement opportunity, not blocking |

## Examples of Common Anti-Patterns

When reviewing, watch for these:

- **God Agent:** Single agent handles orchestration, execution, and safety — no separation
- **Trust Cascade:** Agent A trusts B trusts C, but A has no visibility into C's actions
- **Infinite Retry Loop:** Failed task retried without escalation or backoff
- **Shared Mutable State:** Multiple agents writing to the same store without coordination
- **Permission Creep:** Agents accumulate tool access over time with no revocation
- **Phantom Guardian:** Safety layer exists in docs but has no runtime enforcement
- **Soul Drift:** Self-improving agent gradually changes its core behavior with no anchor
