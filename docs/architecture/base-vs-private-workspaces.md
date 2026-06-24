# BusinessOS Base And Workspace Data

## Purpose

BusinessOS is one reusable application base.

The smart default is not one fork per company. The smart default is one BusinessOS codebase with many workspaces inside it.

Agency MIOSA should be a workspace inside BusinessOS unless there is a specific reason to deploy a separate private fork.

## Public Base Repo

Repo: `Miosa-osa/BusinessOS`

This repo is the reusable product code.

It should stay safe to publish as an open-source base:

- authentication shell,
- workspace shell,
- Knowledge module,
- generic backend services,
- generic docs,
- generic examples,
- no seeded company records,
- no Agency MIOSA operating data,
- no private Google, Gmail, calendar, CRM, payment, or client context.

Current base shape:

- sign in / sign up,
- workspace switcher,
- Knowledge as the primary visible module,
- Settings/Profile for account configuration,
- no seeded company data.

The codebase can still contain dormant implementation modules while the product surface matures, but the default user-facing shell should not expose CRM, Projects, Apps, Agents, Terminal, Computer, or other modules until those modules are deliberately productized.

## Workspace Model

Agency MIOSA should be created as a workspace inside BusinessOS:

```text
BusinessOS public base
  -> reusable code
  -> auth shell
  -> workspace shell
  -> Knowledge module
  -> no private workspace data

BusinessOS runtime
  -> workspace: Agency MIOSA
     -> company knowledge
     -> offers
     -> Robert/Roberto operating map
     -> client delivery process
     -> content strategy
     -> calendar rhythm
     -> tasks and decisions
```

## Rule

Code belongs in the BusinessOS repo.

Workspace data belongs in workspace storage.

Private workspace data never flows back into the public base repo.

## When To Use A Separate Private Repo

A separate private repo is an exception, not the default.

Use one only when the workspace needs:

- custom client-specific source code,
- a hard-isolated deployment,
- contractual restrictions that forbid shared code/runtime,
- a custom module that should not ship in the public base,
- regulated data workflows that require a separate deploy boundary.

Agency MIOSA does not need that by default. It should start as a workspace.

## Agency MIOSA Workspace

The Agency MIOSA workspace should contain:

- Agency MIOSA operating documents,
- company pillars,
- offers,
- Robert/Roberto role map,
- client delivery notes,
- pipeline and follow-up process,
- content strategy,
- calendar rhythm,
- internal tasks,
- private meeting/email context.

This data should be loaded through the app, a seed/import path, or a private workspace export. It should not be committed to the public BusinessOS base.

## First Agency MIOSA Build Target

The first Agency MIOSA workspace should focus on one workflow:

1. Robert and Roberto review the Agency MIOSA alignment package.
2. Decisions are recorded in Knowledge.
3. Next actions become tasks.
4. Offer language, content direction, and owner map are updated.
5. The company operating rhythm becomes visible in the workspace.

Do not build every BusinessOS module first.

Start with Knowledge, then add modules only when the company workflow demands them.
