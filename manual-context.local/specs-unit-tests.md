# SPEC-UNIT-TESTS — Table-Driven Unit Tests

- Version: 1.0
- Status: active
- Purpose: unit-tests
- Updated: 2026-10-25

## Summary

Unit tests in this repository are authored in the table-driven style: the cases for one behavior
live in a single table, and every case runs as its own named subtest. The rule fixes the structure
of a unit test, not what it covers — coverage requirements stay in the per-feature specs.

## Scope

- Includes: unit tests authored or modified from 2026-10-25 onward, and the `code-unit-tester`'s
  authoring standard for them.
- Excludes: acceptance, integration, and end-to-end tests; benchmarks, examples, and fuzz targets;
  test coverage requirements and the verification rows of other specs.
- Existing unit tests are migrated to the pattern when they are next modified; a retroactive rewrite
  of tests that a change does not touch is out of scope.

## Actors

| ID | Name |
|----|------|
| ACTOR-UNIT-TEST-AUTHOR | `code-unit-tester` |
| ACTOR-UNIT-TEST-REVIEWER | `code-reviewer` |

## Requirements

### FR-UNIT-TESTS-001 — A unit test that covers more than one case expresses those cases as one table and runs each row as a named subtest.

- Kind: functional
- Status: active
- Priority: must
- Source: instruction
- Rationale: A single table keeps the covered inputs and expected results side by side, makes the
  boundary and error cases visible in review, and reports a failing case under its own name instead
  of as an anonymous assertion. It also stops a new assertion from becoming a second, differently
  shaped test function for the same behavior.
- Supersedes: none
- Acceptance:
  - **AC-UNIT-TESTS-001** — A unit test that covers two or more cases declares them in one table
    whose entries carry a case `name` together with that case's inputs and expected result, and
    executes every entry through `t.Run(tt.name, ...)` so a failing case is identified by name.
  - **AC-UNIT-TESTS-002** — A unit test that covers exactly one case, and a test that is not a unit
    test (acceptance, integration, end-to-end), may be written without a table; a benchmark,
    example, or fuzz target is out of scope of this requirement.
- Notes: This is a structure rule. It does not require a table where a single case is the whole
  behavior, and it does not change which behaviors must be covered — that remains the job of the
  acceptance criteria in the owning feature spec. `future_err_test.go` (`TestOutcome_Err`) is the
  current example of a multi-case unit test.

## Rules

- **RULE-UNIT-TESTS-001** — One table exercises one behavior. Rows of a table differ in their
  inputs, preconditions, and expected results for that behavior; unrelated behaviors get their own
  test function and their own table.
- **RULE-UNIT-TESTS-002** — A table row is independent: it sets up its own inputs and does not rely
  on another row having run first, so the row set stays valid under `-run`, `-shuffle`, and when a
  single row is selected.

## Verification

| Test | Covers | File | Method | Check | Status |
|------|--------|------|--------|-------|--------|
| TC-UNIT-TESTS-001 | AC-UNIT-TESTS-001 | — | manual | 1. Open the unit-test file changed by the change set.<br>2. Find each test function that covers more than one case.<br>3. Confirm the cases are rows of one table carrying `name`, inputs, and expected result, and that every row runs through `t.Run(tt.name, ...)`.<br>4. Record the file and function in the review. | missing |
| TC-UNIT-TESTS-002 | AC-UNIT-TESTS-002 | — | manual | 1. List the test functions changed by the change set.<br>2. Confirm any single-case or non-unit test omits a table without being rewritten, and that no benchmark, example, or fuzz target was restructured for this rule.<br>3. Record the file and function in the review. | missing |

## Code Map

| File | Role |
|------|------|
| `future_err_test.go` | unit-test — reference example: `TestOutcome_Err` |
| `future_test.go` | unit-test — existing single-case tests, migrated when next modified |

## Open Questions

- **Q-UNIT-TESTS-001** (`unverified`) — The instruction fixes the pattern but not the migration
  deadline. This spec lets existing tests migrate when they are next modified; a repo-wide rewrite
  in one change would settle it the other way.

## Changelog

- **1.0** (2026-10-25, created) — Established the table-driven unit-test standard at the user's
  instruction: one table per behavior, every case a named `t.Run` subtest, with single-case and
  non-unit tests exempt and existing tests migrated when next modified.
