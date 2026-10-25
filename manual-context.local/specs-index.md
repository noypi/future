# Specs Index

- Version: 1.19
- Updated: 2026-10-25
- Standard: manual-context/std-naming.md

| Spec ID | File | Purpose | Status | Version | Requirements | Verified | Updated | Title | Code roots |
|---------|------|---------|--------|---------|--------------|----------|---------|-------|------------|
| SPEC-FUTURE | [specs-future.md](specs-future.md) | future | active | 1.19 | 20 | 13 | 2026-10-25 | Async Futures and Promise Chaining | `future.go`, `future_test.go`, `race_test.go`, `go.mod`, `go.sum`, `Makefile` |
| SPEC-UNIT-TESTS | [specs-unit-tests.md](specs-unit-tests.md) | unit-tests | active | 1.0 | 1 | 0 | 2026-10-25 | Table-Driven Unit Tests | — |

## Notes

- 1.19 adds `FR-FUTURE-017` (`AC-FUTURE-034`/`AC-FUTURE-035`, rows `TC-FUTURE-026`/`TC-FUTURE-027`)
  and `FR-FUTURE-018` (`AC-FUTURE-036`/`AC-FUTURE-037`, rows `TC-FUTURE-028`/`TC-FUTURE-029`), all
  `missing`. That is 18 `FR` + 2 active `NFR`, so `Requirements` moves 18 to 20; `Verified` stays 13
  until a new row passes. Rows `TC-FUTURE-023` to `TC-FUTURE-029` name test files that do not exist
  yet, so a `go test -run` that matches nothing must not promote them.
- `Requirements` counts active functional and non-functional requirements (18 `FR` + 2 active
  `NFR`). `NFR-FUTURE-002` is `unverified` and therefore excluded. The 18 `FR` include the two added
  in 1.14 (`FR-FUTURE-014`, `FR-FUTURE-015`), whose six verification rows `TC-FUTURE-017` to
  `TC-FUTURE-022` are one per acceptance criterion, plus `FR-FUTURE-016` to `FR-FUTURE-018` added in
  1.18/1.19 with `TC-FUTURE-023` to `TC-FUTURE-029`; as of 1.17 `TC-FUTURE-021` (`FR-FUTURE-015`,
  `AC-FUTURE-029`) and `TC-FUTURE-022` (`FR-FUTURE-015`, `AC-FUTURE-030`) are `covered`; the other
  four of that 1.14 set and the seven newer rows stay `missing`. No requirement count changed for
  Feature 1 because it adds no new requirement.
- `Verified` counts requirements with at least one `covered` row. "Covered" now means the declared
  `Method`/`Check` last passed; the latest `nfr-audit run --all-rows` (2026-10-25T06:07:02Z, scope
  `all`, 22 rows: 15 pass, 6 unknown, 1 manual, 0 fail) over the `ac-future-029` run checkout
  (`/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-029/_run`, commit `2b4706f`) reports 2
  NFRs passing — `NFR-FUTURE-001` via `TC-FUTURE-002` and `NFR-FUTURE-003` via `TC-FUTURE-014` to
  `TC-FUTURE-016` — with `TC-FUTURE-013` still `manual` for `NFR-FUTURE-002`, which stays
  manual/unverified. The count moved from 12 to 13 because `FR-FUTURE-015` gained another covered
  row through `TC-FUTURE-021` (`exit=0`); it was already verified through `TC-FUTURE-022`.
  `TC-FUTURE-017` to `TC-FUTURE-020` also report `verdict=pass`, but those passes are vacuous —
  their named test files (`future_settle_test.go`, `future_accessor_test.go`,
  `future_reject_test.go`) are absent from that checkout, `future_test.go` has no
  `TestFuture_UntypedNil*` function, and `go test -run '<regex>' .` exits 0 when nothing matches — so
  they stay `missing` and are not counted. `TC-FUTURE-012` remains `covered` from a recorded test run
  at commit `e5783cf` in the `ac-future-018` run checkout — an FR row, so that provenance, not an NFR
  pass, is its evidence, although the latest all-rows run also lists it `exit=0`. Rows that declare
  no `Method` stay `unknown` to the audit and keep their recorded `Status`.
- Three of the four test functions in `future_test.go` map to criteria; `BenchmarkFuture` asserts
  nothing and is not mapped. `race_test.go` supplies the three `TestRace_*` tests behind
  `TC-FUTURE-012`.
