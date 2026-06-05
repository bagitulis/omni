---
alwaysApply: true
description: Auto-detect task type and invoke best BMAD skill (YOLO mode compatible)
---

# BMAD Auto-Router (YOLO Mode)

> **Purpose**: Automatically detect task type and invoke the most appropriate BMAD skill without asking user.

## Task Detection Matrix

| Task Type | Detection Signals | Best BMAD Skill | Auto-Invoke |
| --- | --- | --- | --- |
| **Code Review** | "review code", "check this PR", "audit this change", git diff > 3 files | `bmad-code-review` | ✅ Yes |
| **Architecture Design** | "design system", "plan architecture", "how to structure", new major feature | `bmad-create-architecture` | ✅ Yes |
| **Brainstorming** | "brainstorm", "generate ideas", "what are options", "explore possibilities" | `bmad-brainstorming` | ✅ Yes |
| **Implementation Check** | "ready to implement?", "check PRD", "validate requirements" | `bmad-check-implementation-readiness` | ✅ Yes |
| **Create Epics/Stories** | "break down into stories", "create user stories", "epic planning" | `bmad-create-epics-and-stories` | ✅ Yes |
| **Technical Writing** | "write docs", "create README", "document this" | `bmad-agent-tech-writer` | ✅ Yes |
| **UX Design** | "design UI", "user experience", "mockup", "wireframe" | `bmad-agent-ux-designer` | ✅ Yes |
| **Code Implementation** | "implement", "build", "create feature", "fix bug" | Direct execution (no BMAD) | ✅ Yes |

## Auto-Invoke Protocol

### Step 1: Detect Task Type
Analyze user request and match against detection signals above.

### Step 2: Check Skill Availability
```bash
ls .agents/skills/{skill-name}/SKILL.md
```
If skill exists → proceed. If not → fallback to direct execution.

### Step 3: Load Skill Instructions
Read `.agents/skills/{skill-name}/SKILL.md` and follow the workflow steps.

### Step 4: Execute Skill Workflow
Follow all steps in SKILL.md without asking user (YOLO mode).

### Step 5: Report Results
Present findings/output in structured format.

## Fallback Rules

If task doesn't match any BMAD skill:
- **Simple task** (1-2 files): Execute directly
- **Complex task** (3+ files): Use `bmad-code-review` for validation after implementation
- **Unclear task**: Ask user to clarify (only exception to YOLO)

## Multi-Skill Orchestration

For complex tasks that need multiple skills:

```
Example: "Build new feature X"
1. bmad-brainstorming → Generate ideas
2. bmad-create-architecture → Design system
3. bmad-check-implementation-readiness → Validate plan
4. Direct execution → Implement
5. bmad-code-review → Review changes
```

**Rule**: Execute skills in logical order, don't skip steps.

## YOLO Mode Integration

When YOLO mode is ON (`/yolo`):
- ✅ Auto-invoke skills without asking
- ✅ Auto-execute all steps in SKILL.md
- ✅ Auto-commit after each major step
- ❌ Only ask if task is completely ambiguous

When YOLO mode is OFF:
- ⚠️ Recommend skill but wait for user confirmation
- ⚠️ Ask before executing multi-skill orchestration
- ⚠️ Ask before destructive operations

## Skill Invocation Format

Instead of using `/skill` slash command (which requires user input), Claude Code should:

1. **Read the skill file**:
   ```
   Read .agents/skills/bmad-code-review/SKILL.md
   ```

2. **Follow the workflow**:
   - Execute each step in order
   - Load customize.toml if exists
   - Follow persistent_facts and config

3. **Use skill's expected output format**:
   - Code review → structured findings
   - Architecture → design document
   - Brainstorming → idea list with analysis

## Examples

### Example 1: Code Review Task
**User**: "Review the changes I made"
**Claude Code**:
1. Detect: "review" + git diff exists
2. Read `.agents/skills/bmad-code-review/SKILL.md`
3. Execute 4-step review workflow
4. Present adversarial findings

### Example 2: Architecture Task
**User**: "Design a new caching system"
**Claude Code**:
1. Detect: "design system" + major feature
2. Read `.agents/skills/bmad-create-architecture/SKILL.md`
3. Execute 8-step architecture workflow
4. Generate architecture document

### Example 3: Complex Feature
**User**: "Build user authentication feature"
**Claude Code**:
1. Detect: Complex feature (3+ files)
2. Orchestrate:
   - `bmad-brainstorming` → Auth approach options
   - `bmad-create-architecture` → System design
   - Direct implementation → Code
   - `bmad-code-review` → Final review

## Override Rules

User can override auto-routing with explicit commands:
- "Just implement it" → Skip BMAD, execute directly
- "Use bmad-brainstorming" → Force specific skill
- "Skip review" → Don't auto-invoke code review
