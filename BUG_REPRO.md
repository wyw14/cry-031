# Bug reproduction

- Bug: A disabled administrator can still read another member's profile.
- Trigger: Disable an administrator account, then request a different member profile with that account.
- Error: The request succeeds instead of returning the protected-profile denial.
