# Runtime ownership lessons

## Outbox ownership follows the runtime, not shared database access

A document process needs identity state to authenticate users but must not thereby
own identity publication. Remote document workers drain only the document outbox;
the Dia runtime owns identity and wallet notices. The inline demo explicitly owns
both and uses direct SessionProjection. Keep this rule at publicationOwners and
enforce it with the mode matrix and strict router tests, rejecting foreign notices
rather than acknowledging unsupported work.
