# Bug reproduction

- Bug: Activity discovery without an explicit sort has unstable ordering.
- Trigger: Query the same activities repeatedly without a sort parameter.
- Error: Results can arrive in different orders instead of ascending start time.
