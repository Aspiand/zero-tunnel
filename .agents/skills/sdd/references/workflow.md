# SDD Detailed Workflow

Follow these steps strictly for every non-trivial change.

## Phase 1: Exploration & Discussion
1. Analyze the user's request.
2. Identify the parts of `SPECS.md` affected.
3. Propose a technical solution in the chat.

## Phase 2: Confirmation & Documentation
1. Wait for user approval of the proposed solution.
2. **CRITICAL**: Update `SPECS.md` first.
   - Use the `replace` tool for surgical edits.
   - Maintain consistency with the existing style.
   - Use templates from `templates.md` where appropriate.

## Phase 3: Implementation
1. Write code based on the newly updated specification.
2. Do not deviate from the spec without further discussion.

## Phase 4: Validation
1. Run tests.
2. Compare the final implementation with the specification.
3. If unexpected changes occur during implementation, update `SPECS.md` to reflect the code reality (with user approval).
