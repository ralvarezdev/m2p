# PLC Code Generation Agent — Initial Plan

Status: **draft, pre-implementation**
Author: planning session, 2026-05-26
Supersedes: nothing (companion to `2026-05-18-INITIAL_PLAN.md`, which covers the Python sandbox agent)

---

## 1. Goal and Scope

Add a second specialization to the existing ADK agent system: a conversational agent that emits **IEC 61131-3 Structured Text (ST)** Program Organization Units (POUs) on behalf of the user. The agent runs the same multi-turn loop as the Python sandbox agent — user prompt in, tool calls out, results back — but its tools produce `.st` source files instead of executing Python.

### In scope (Phase 0)
- Single tool surface for emitting and reading ST source.
- Per-conversation working tree on disk (`pous/*.st`), snapshotted to git per turn (reuses existing `Committer`).
- Strict **IEC 61131-3** dialect only — no CODESYS-specific extensions, no vendor function blocks.
- A proposed POU/variable style guide that ships with the system prompt.
- Standalone POU mode only — each POU is independent; no project-level wiring.

### Out of scope (deferred to later phases)
- **Compiler validation.** No `matiec`, no CODESYS scripting. The agent emits source; the user takes it to CODESYS manually for now.
- **Behavioral simulation.** No soft-PLC runtime, no I/O traces.
- **Single growing CODESYS project model.** Defer to Phase 1.
- **Reading existing CODESYS projects.** Defer to Phase 1 (requires CODESYS access on the host).
- **Graphical languages (LD, FBD, SFC).** ST only.

---

## 2. Phased Roadmap

| Phase | Adds | Requires | Status |
|---|---|---|---|
| **0 — Generation only** | Tool surface, POU file persistence, system prompt, style guide. No execution feedback. | Nothing new; runs on existing infra. | This plan. |
| **1 — CODESYS bridge** | `build_project` returning CODESYS diagnostics; `read_pou` for existing projects; optional single-project mode. | CODESYS IDE installed on agent host; Scripting Engine wired. | Planned, not designed. |
| **2 — Behavioral validation** | `simulate_cycles` against CODESYS Control runtime in Docker. | Control runtime license decision. | Planned, not designed. |

Phase 0 is shippable on its own. Phase 1 is an additive sandbox backend swap; it does not require Phase 0 to be rewritten.

---

## 3. Conversation Routing

Reuse the existing `sandbox_type` discriminator on `conversations`. Add one new value:

```
sandbox_type: anthropic | docker | e2b | daytona | plc
```

When `sandbox_type == "plc"`, the agent loop:
- Uses the PLC tool surface (section 5) instead of `execute_python`.
- Uses the PLC system prompt (section 7).
- Routes tool calls to a `PLCBackend` instead of a Python sandbox.

No existing Python conversations are affected. UI exposes the choice at conversation creation.

---

## 4. Domain Changes

Minimum surface that keeps the existing schema usable.

### 4.1 Enum additions

```go
// internal/domain/enums.go

type SandboxType string
const (
    SandboxAnthropic SandboxType = "anthropic"
    SandboxDocker    SandboxType = "docker"
    SandboxE2B       SandboxType = "e2b"
    SandboxDaytona   SandboxType = "daytona"
    SandboxPLC       SandboxType = "plc"   // NEW
)

type CodeLanguage string
const (
    LangPython         CodeLanguage = "python"
    LangStructuredText CodeLanguage = "structured_text" // NEW
)
```

### 4.2 `CodeExecution` widened

Add `Language CodeLanguage` to `domain.CodeExecution`. Existing rows default to `python` via migration. `stdout` is reused to carry tool result text (e.g. `"wrote pous/PRG_Main.st (412 bytes)"`); `stderr` stays empty in Phase 0; `exit_code` is `0` on success, `1` on bad tool input. This keeps `code_executions` schema and the SSE stream unchanged.

### 4.3 No new tables

POU files live on disk under the existing snapshot tree (section 8). They are not persisted as DB rows — git history is the audit log, matching the Python agent's design.

---

## 5. Tool Surface (Phase 0)

Four tools. All write/read are scoped to the current conversation's POU directory; the agent cannot escape it.

| Tool | Purpose | Input | Output |
|---|---|---|---|
| `write_pou` | Create or overwrite a POU `.st` file. | `name`, `kind` (`PROGRAM`\|`FUNCTION`\|`FUNCTION_BLOCK`), `return_type` (functions only), `declarations` (VAR blocks as text), `body` (ST statements), `header_comment` | Path written, byte count. |
| `write_gvl` | Create or overwrite a Global Variable List `.st`. | `name`, `declarations`, `header_comment` | Path written, byte count. |
| `list_pous` | Inventory of current conversation's POUs. | (none) | Array of `{name, kind, path, bytes}`. |
| `read_pou` | Read back a POU's current content. | `name` | Full file content as string. |

Notes:
- POU body is assembled server-side from `declarations + body` so the LLM emits structured pieces; the backend interleaves the IEC keywords (`PROGRAM Name`, `VAR ... END_VAR`, `... END_PROGRAM`). This guards against the LLM omitting boilerplate.
- All four tools share one input envelope: see section 6.
- `list_pous` and `read_pou` are required so the agent can self-inspect across turns when the model loses thread of what it already produced.

### 5.1 Sandbox interface generalization

The existing `Sandbox.Execute(ctx, code string)` assumes one tool. To support multiple tools without forcing every backend to do JSON-in-`code` envelopes, the cleanest move is:

```go
// internal/sandbox/sandbox.go

type ToolCall struct {
    Name  string
    Input map[string]any
}

type Sandbox interface {
    Invoke(ctx context.Context, call ToolCall) (*domain.SandboxResult, error)
}
```

Existing Python backends become a thin shim: `Invoke` switches on `call.Name == "execute_python"` and pulls `call.Input["code"].(string)`. This is a small, mechanical migration of four files; the alternative (keep `Execute`, add `MultiToolSandbox` alongside) duplicates dispatch logic at the agent layer.

**Recommendation:** do the generalization. It's load-bearing for Phase 1 (CODESYS scripting will also be multi-tool).

---

## 6. PLC Sandbox Backend

Phase 0 backend is a **filesystem-only sandbox** — no container, no subprocess.

```
internal/sandbox/plc/
    backend.go      # Factory + Invoke dispatch
    pou.go          # POU/GVL file assembly (boilerplate insertion, sanitization)
    paths.go        # Conversation-scoped path helpers; rejects '..' and absolute paths
    constants.go    # File suffixes, kind keywords, max byte limits per POU
```

`Invoke` switches on tool name, validates input, writes the file (or reads it), and returns a `SandboxResult` whose `stdout` summarizes the operation. No git interaction here — the existing `Committer` runs after the turn closes, picks up all files written this turn, and commits them.

There is no container to acquire or release. The Manager's `Acquire(conversationID)` returns a per-conversation backend struct (cheap to allocate) that holds the resolved working-tree path. `release()` is a no-op.

---

## 7. System Prompt Outline

Five sections, kept short:

1. **Role.** "You generate IEC 61131-3 Structured Text POUs. You do not write Python, Go, or pseudocode. You do not generate ladder logic."
2. **Strict dialect rule.** "Emit only standard IEC 61131-3 constructs. Do not use CODESYS-specific extensions (`__VARINFO`, `THIS^`, `__ISVALIDREF`, attribute pragmas) or vendor function blocks. Code must be portable across CODESYS, TwinCAT, Studio 5000."
3. **Style guide.** Inlined from section 9.
4. **Tool usage protocol.** "Use `list_pous` at the start of any turn that may edit existing code. Use `read_pou` before modifying a POU. Use `write_pou` for the final emit. Prefer one POU per call."
5. **No execution feedback.** "There is no compiler or simulator in this environment. Do not claim a POU was tested. Comment any assumption that would normally be verified by a build."

The full text lives in `internal/agent/plc/prompt.go` as a `//go:embed`'d markdown file (mirrors the project's "embedded SQL" pattern from the architecture doc).

---

## 8. Persistence Layout

Reuses the existing snapshot root.

```
{snapshotRoot}/{conversationID}/
    pous/
        PRG_Main.st
        FB_MotorControl.st
        GVL_IO.st
    docs/                   # reserved for Phase 1+ generated docs (not Phase 0)
    .git/                   # existing per-conversation repo, managed by Committer
```

Per turn, the `Committer` (unchanged) stages `pous/` and commits with message `"Turn N: <user prompt summary>"`. File paths are returnable by the agent — `write_pou` returns the relative path (`pous/PRG_Main.st`), which the agent can quote back to the user and which is stable for the user to open in CODESYS.

`read_pou` and `list_pous` operate over the working tree directly, not git — they always see the most recent write in the current turn.

---

## 9. POU and Variable Style Guide (proposed)

This is the proposal; iterate before locking it into the prompt.

### 9.1 POU naming
- **Programs:** `PRG_<Purpose>` — `PRG_Main`, `PRG_Conveyor`.
- **Function blocks:** `FB_<Noun>` — `FB_MotorControl`, `FB_PIDLoop`.
- **Functions:** `FC_<VerbNoun>` — `FC_ScaleAnalog`, `FC_ClampInt`.
- **GVLs:** `GVL_<Domain>` — `GVL_IO`, `GVL_Setpoints`.
- All PascalCase. No leading underscores. No vendor prefixes (`P_`, `M_`).

### 9.2 Variable naming
Scope-prefix + PascalCase:

| Prefix | Scope |
|---|---|
| `i` | `VAR_INPUT` |
| `o` | `VAR_OUTPUT` |
| `iq` | `VAR_IN_OUT` |
| (none) | `VAR` (local) |
| `r` | `VAR RETAIN` |
| `g` | global (declared in a GVL) |
| `k` | `VAR CONSTANT` (SCREAMING_SNAKE_CASE for the rest of the name) |

Examples: `iStartButton`, `oMotorRun`, `rCycleCount`, `gActiveRecipe`, `kMAX_SPEED`.

### 9.3 POU comment header
Every POU starts with a block comment:

```
(*
 * POU:         PRG_Main
 * Kind:        PROGRAM
 * Purpose:     Top-level scan-cycle dispatcher.
 * Author:      <agent>
 * Generated:   2026-05-26
 * Assumptions: <one line per assumption that should be verified at build>
 *)
```

### 9.4 Body conventions
- Indent **4 spaces** (CODESYS default is 4; matches most house styles).
- One statement per line. No more than 80 columns where avoidable.
- `IF`/`CASE` keywords aligned with their `END_`.
- Boolean conditions wrapped in parens when combined: `IF (xRunning AND NOT xFault) THEN`.
- No bare magic numbers in body — promote to a `VAR CONSTANT` or GVL constant.
- Comments only where the *why* is non-obvious. Do not narrate the *what*.

### 9.5 Things the agent must NOT emit
- `GOTO`, `EXIT` from `FOR`/`WHILE` (use a guarded loop condition).
- Pointer arithmetic, `ADR()` / `^` deref (CODESYS extension, non-portable).
- Inline assembly, `__VARINFO`, attribute pragmas.
- Empty `VAR ... END_VAR` blocks.

---

## 10. File-by-File Change List

### New files
- `internal/sandbox/plc/backend.go` — factory + `Invoke` dispatch
- `internal/sandbox/plc/pou.go` — POU/GVL assembly
- `internal/sandbox/plc/paths.go` — sandboxed path resolution
- `internal/sandbox/plc/constants.go` — kind keywords, suffixes, limits
- `internal/agent/plc/prompt.go` — `//go:embed prompt.md`
- `internal/agent/plc/prompt.md` — the system prompt from section 7
- `internal/agent/plc/tools.go` — tool schema definitions (Anthropic tool format)
- `internal/store/sql/conversations/0002_add_language.sql` — migration adding `language` column to `code_executions` (default `python`)

### Modified files
- `internal/domain/enums.go` — add `SandboxPLC`, `CodeLanguage` enum
- `internal/domain/types.go` — add `Language` field to `CodeExecution`
- `internal/sandbox/sandbox.go` — change `Sandbox.Execute(ctx, code)` to `Sandbox.Invoke(ctx, ToolCall)`; add `ToolCall` type
- `internal/sandbox/anthropic/backend.go` — implement `Invoke` as a shim over the old `Execute`
- `internal/sandbox/docker/backend.go` — same
- `internal/sandbox/e2b/backend.go` — same
- `internal/sandbox/daytona/backend.go` — same
- `internal/agent/agent.go` — route to PLC tool list + prompt when `sandbox_type == "plc"`
- `internal/agent/loop.go` — generalize tool dispatch to use `ToolCall` instead of a hardcoded `code` arg
- `internal/store/turn.go` + relevant SQL — read/write the new `language` column
- `internal/config/constants.go` — POU size limits, PLC working-dir name (`"pous"`)

### Migration risk
The `Sandbox` interface change is the only breaking refactor. It touches four backend files in a mechanical way (rename + wrap). Tests for Python backends should be re-run; no behavior change is intended.

---

## 11. Open Questions

1. **POU file extension.** `.st` is conventional but CODESYS exports use `.EXP` for POU exports and `.xml` for PLCopen XML. Going with `.st` keeps files readable in any editor; if you plan to import into CODESYS via PLCopen XML later (Phase 1), we'll add a parallel `.xml` emitter at that time.
2. **Multiple POUs per turn.** Allowed — the agent can call `write_pou` repeatedly within one Claude response. Should we cap it (e.g., max 5 per turn) to keep turns reviewable? Default: no cap, revisit if turns get noisy.
3. **Conversation reset.** If a user wants to "throw away" the current POU set and start over, do we expose a `reset_pous` tool, or do they create a new conversation? Default: new conversation; keeps history immutable.
4. **POU naming collisions.** If the agent calls `write_pou` with an existing name, overwrite silently or require an explicit `overwrite=true` flag? Default: overwrite silently — the git history is the safety net, and forcing flags slows the loop.

These are Phase 0 finalization questions; none block the design above.

---

## 12. Phase 1 Hooks (so we don't paint into a corner)

Decisions in this plan that exist specifically to make Phase 1 painless:
- **`Sandbox.Invoke(ToolCall)`** generalization — CODESYS backend will register `build_project`, `read_pou_from_project`, etc.
- **Per-conversation working tree on disk** — CODESYS Scripting Engine reads from disk; nothing about Phase 0's persistence needs to change.
- **`SandboxType` enum, not boolean** — adding `SandboxCODESYS` is one line.
- **`CodeLanguage` enum** — same story for ST vs (eventually) IL or SFC text export.

What we are deliberately NOT building yet:
- A project-level abstraction (`Project`, `Library`, `Device`) — premature without CODESYS in the loop to validate the shape.
- A diagnostic event type in SSE (`compile_diagnostics`) — no diagnostics in Phase 0.
- Read access to a user-supplied CODESYS project on disk — requires path/permission model we can design once we know how CODESYS exposes projects.
