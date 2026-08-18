# Bug reproduction

- Bug: The public team discovery response exposes a private team's invite code.
- Trigger: Query team discovery as a visitor who is not a team member.
- Error: The response includes the invite credential alongside public team fields.
