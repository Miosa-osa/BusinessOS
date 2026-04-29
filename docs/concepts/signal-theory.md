# Signal Theory

> Adapted from Alice H. Luna, *Signal Theory: The Architecture of Optimal Intent
> Encoding* (MIOSA Research, February 2026). This is the theoretical foundation
> the Optimal Engine is built to instantiate.

## Root objective

**Maximize S/N** — the ratio of actionable intent to noise in every output. S/N is
not one metric among many; it is the measure of communication quality.

## The Signal

Every output is a **Signal**, modeled as:

```
S = (M, G, T, F, W)
```

| Dim | Name        | Question it answers                              | Examples                                          |
|-----|-------------|--------------------------------------------------|---------------------------------------------------|
| M   | **Mode**    | How is it perceived?                             | linguistic, visual, code, data, mixed             |
| G   | **Genre**   | What conventionalized form?                      | spec, brief, plan, transcript, report, ADR, note  |
| T   | **Type**    | What does it DO?                                 | direct, inform, commit, decide, express           |
| F   | **Format**  | What container?                                  | markdown, code, JSON, CLI output, diff            |
| W   | **Structure** | Internal skeleton                              | genre-specific template                            |

Before producing any non-trivial output, resolve all five dimensions. Unresolved
dimensions create noise.

## Four governing constraints

Any Signal that violates one of these fails regardless of its content.

1. **Shannon — the ceiling.** Every channel has finite capacity. Don't exceed the
   receiver's bandwidth. A 500-line explanation when 20 lines suffice is a
   Shannon violation.
2. **Ashby — the repertoire.** Have enough Signal variety (genres, modes,
   structures) to handle every situation. Prose when a table is needed is an
   Ashby violation.
3. **Beer — the architecture.** Maintain viable structure at every scale. A
   response, a file, a system — each must be coherently structured. Orphaned
   logic is a Beer violation.
4. **Wiener — the feedback loop.** Never broadcast without confirmation. Close
   the loop: verify the receiver decoded correctly. Ask when ambiguous. Check
   that the action happened.

## Six encoding principles

Apply on every non-trivial output:

1. **Mode-message alignment** — sequential logic goes to text/code; relational
   logic goes to diagrams/tables. Don't use the wrong logic for the message.
2. **Genre-receiver alignment** — match genre to receiver. Developers decode
   specs. Executives decode briefs. Wrong genre = failed Signal.
3. **Structure imposition** — raw information is noise. Always impose structure.
   Headers, sections, genre-specific skeletons.
4. **Redundancy proportional to noise** — high-stakes contexts get more
   structure and explicit intent. Simple contexts get minimal framing.
5. **Entropy preservation** — maximum meaning per unit of output. No filler, no
   hedging, no padding. An SOP that buries the action in paragraphs has lost
   entropy to noise.
6. **Bandwidth matching** — match output density to receiver capacity. Three
   bullet points when that's what's needed. Full spec when that's what's needed.

## Eleven failure modes

```
SHANNON VIOLATIONS
  Routing failure       wrong recipient. Re-route.
  Bandwidth overload    too much output. Reduce, prioritize, batch.
  Fidelity failure      meaning lost. Re-encode with clearer structure.

ASHBY VIOLATIONS
  Genre mismatch        wrong form. Re-encode in correct genre.
  Variety failure       no genre exists. Create one.
  Structure failure     no internal skeleton. Impose genre structure.

BEER VIOLATIONS
  Bridge failure        no shared context. Add preamble/conventions.
  Herniation failure    incoherence across layers. Re-encode with proper traversal.
  Decay failure         outdated Signal. Audit, version, or sunset.

WIENER VIOLATIONS
  Feedback failure      no confirmation loop. Close it.

CROSS-CUTTING
  Adversarial noise     deliberate degradation. Make visible, escalate.
```

The engine's `mix optimal.health` and `OptimalEngine.Signal.FailureModes`
modules detect these at the data layer.

## Path of least resistance

Optimal encoding converges on minimum decoding effort at the receiver. The best
Signal:

- Uses the **mode** the receiver perceives most naturally
- Takes the **genre** the receiver has competence to decode
- Performs the **type** (speech act) that matches intended outcome
- Arrives in the **format** appropriate to the channel
- Has a **structure** that makes the internal skeleton immediately decodable
- Carries enough **redundancy** to survive noise
- Minimizes **decoding effort** at the receiver
- Produces **action** with highest probability

What the receiver experiences as clarity, elegance, or "just right" — that is
the path of least resistance. Maximum meaning per unit of signal, minimum noise
per unit of meaning.

## How the engine instantiates this

| Theory element              | Engine artifact                                                     |
|----------------------------|---------------------------------------------------------------------|
| Signal `S=(M,G,T,F,W)`     | `OptimalEngine.Signal.Envelope` struct — CloudEvents + classification |
| Classification             | `OptimalEngine.Signal.Classifier` — auto-infers dimensions           |
| Genre routing              | `OptimalEngine.Router` — trie-based pattern matching                 |
| Failure-mode detection     | `OptimalEngine.Signal.Classifier.FailureModes`                       |
| Bandwidth matching (tiers) | L0 (~100 tok) / L1 (~2K tok) / full — `ContextAssembler`             |
| Entropy preservation       | Redundancy budget enforced in `Indexer`; duplicates surfaced by `optimal.health` |
| Feedback loop              | `OptimalEngine.Memory.Learning` (SICA) — corrections feed back in    |

## Recommended reading order

1. This document.
2. [`three-spaces.md`](three-spaces.md) — input, signal, persistence separation.
3. [`failure-modes.md`](failure-modes.md) — deep dive on the 11 failure modes.
4. [`infinite-context-framework.md`](infinite-context-framework.md) — tiered
   loading as bandwidth matching at scale.
5. [`../architecture/system-overview.md`](../architecture/system-overview.md) —
   how the theory maps onto the running system.
