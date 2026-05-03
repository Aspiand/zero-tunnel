---
name: sdd
description: Spec Driven Development (SDD) workflow. Use this skill to ensure all confirmed design decisions, feature requirements, and architectural changes are documented in SPECS.md before implementation.
---

# Spec Driven Development (SDD)

This skill enforces a "Spec First" workflow. Every major decision or feature must be discussed, confirmed by the user, and then recorded in `SPECS.md` before any code is written.

## Workflow

1.  **Discuss**: When a new feature or change is requested, discuss the technical approach with the user.
2.  **Confirm**: Explicitly ask for confirmation of the proposed strategy.
3.  **Update Specs**: Once confirmed, update `SPECS.md` to reflect the decision. This ensures the specification remains the "single source of truth".
4.  **Implement**: Only after the spec is updated, proceed with the implementation.
5.  **Validate**: Ensure the implementation matches the updated specification.

## Rules for SPECS.md

-   **Single Source of Truth**: `SPECS.md` must always reflect the current intended state of the project.
-   **Incremental Updates**: Do not overwrite the entire spec unless necessary. Add or modify sections surgically.
-   **Clarity**: Use clear, concise language and structured formatting (tables, lists, headings).
-   **History**: If significant changes are made to previous decisions, note them or update the relevant sections to keep the document current.

## Triggering

Use this skill whenever:
-   A new feature is proposed.
-   Architectural decisions are being made (e.g., choosing between API vs. File-based management).
-   The user says "update the specs" or "add this to the spec".
