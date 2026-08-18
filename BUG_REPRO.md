# Bug reproduction

- Bug: An idempotency key is shared across different activity creators.
- Trigger: Have two different operators create activities with the same key.
- Error: The second request returns the first operator's activity instead of creating its own.
