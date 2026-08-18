# Bug reproduction

- Bug: A valid private-team invite fails when copied with surrounding spaces.
- Trigger: Submit the invite code with one or more leading or trailing spaces.
- Error: The join request reports an invalid invite even though the code is correct.
