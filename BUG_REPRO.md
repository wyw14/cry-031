# Bug reproduction

- Bug: A risk follow-up accepts a missing or disabled owner.
- Trigger: Create a follow-up using an unknown or inactive user as owner.
- Error: The record is saved and later notification handling has no valid recipient.
