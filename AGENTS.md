# Dia Go repository instructions

Read README.md and relevant dependency documentation before changes. Apply the
pinned [workflow/ponytail/SKILL.md](workflow/ponytail/SKILL.md),
[docs/PLAN_REFINEMENT.md](docs/PLAN_REFINEMENT.md) and [tasks/lessons.md](tasks/lessons.md).
Write a checkable plan in tasks/todo.md before implementation and run
`python3 tools/check_plan_refinement.py tasks/todo.md`. Carry these paths into
any delegated task and preserve existing changes.

The pinned Toolbox dependency owns domain and runtime implementation. Do not
copy it here or silently override it with a local replace directive. Keep one
implementation per rule. Verify the consumer build, tests and actual container
lifecycle when changing the dependency or entrypoint. Cloud writes, publication,
merges and blockchain transactions require explicit user authority. Use no
middots in documentation.
