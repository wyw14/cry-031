# Bug reproduction

- Bug: A completed handoff can be acknowledged again.
- Trigger: Confirm the same risk handoff after it is already completed.
- Error: The request succeeds and emits another confirmation side effect.
