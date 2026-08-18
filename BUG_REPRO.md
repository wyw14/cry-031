# Bug reproduction

- Bug: A service correction accepts a future check-in or check-out time.
- Trigger: Submit a correction whose time is after the current clock.
- Error: A service record for an event that has not happened is accepted.
