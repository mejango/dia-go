# Dia Go repository instructions

Read README.md and relevant dependency documentation before changes. Apply the
pinned [workflow/ponytail/SKILL.md](workflow/ponytail/SKILL.md),
[docs/PLAN_REFINEMENT.md](docs/PLAN_REFINEMENT.md) and [tasks/lessons.md](tasks/lessons.md).
Write a checkable plan in tasks/todo.md before implementation and run
`python3 tools/check_plan_refinement.py tasks/todo.md`. Carry these paths into
any delegated task and preserve existing changes.

Dia owns all application models, policy, services, HTTP routes and runtime.
Toolbox is a third-party framework dependency and must never depend on Dia.
Use only its public APIs; do not commit local replace directives or duplicate its
implementation. Verify both independent modules and actual container lifecycles
when changing their boundary. Cloud writes, publication,
merges and blockchain transactions require explicit user authority. Use no
middots in documentation.
