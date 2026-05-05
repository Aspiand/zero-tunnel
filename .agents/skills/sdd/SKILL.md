---
name: sdd
description: Spec Driven Development (SDD) workflow. Use this skill to ensure all confirmed design decisions, feature requirements, and architectural changes are documented in SPECS.md before implementation.
---

# Spec Driven Development (SDD)

This skill enforces a "Spec First" workflow. Every major decision or feature must be discussed, confirmed by the user, and then recorded in `SPECS.md` before any code is written.

## Workflow

1.  **Analyze & Discuss**: Identify the impact of the request on the system architecture.
2.  **Confirm**: Ask for explicit confirmation of the proposed strategy.
3.  **Update SPECS.md**: `SPECS.md` remains the "Single Source of Truth".
4.  **Implement**: Write code that adheres strictly to the specification.
5.  **Validate**: Verify that the implementation matches the specification document.

## Core Rules

-   **Spec-First**: Never write code for new features without updating `SPECS.md` first.
-   **Consistency**: Maintain a clean, structured, and professional format in `SPECS.md`.
-   **Traceability**: Every feature must be traceable from the specification to the code implementation.

## References
- [Detailed Workflow](./references/workflow.md)

## Triggering

Use this skill whenever:
- A new feature is proposed.
- Architectural decisions are being made (e.g., choosing between API vs. File-based management).
- The user says "update the specs" or "add this to the spec".
