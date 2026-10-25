# Specs History

- Version: 1
- Updated: 2026-10-25T18:30:00+08:00

## HIST-0001 — 2026-10-23T10:30:00Z
- Action: created
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.0
- Requirements: FR-FUTURE-001, FR-FUTURE-002, FR-FUTURE-003, FR-FUTURE-004, FR-FUTURE-005, FR-FUTURE-006, FR-FUTURE-007, FR-FUTURE-008, FR-FUTURE-009, FR-FUTURE-010, FR-FUTURE-011, FR-FUTURE-012, FR-FUTURE-013, NFR-FUTURE-001, NFR-FUTURE-002, AC-FUTURE-001, AC-FUTURE-002, AC-FUTURE-003, AC-FUTURE-004, AC-FUTURE-005, AC-FUTURE-006, AC-FUTURE-007, AC-FUTURE-008, AC-FUTURE-009, AC-FUTURE-010, AC-FUTURE-011, AC-FUTURE-012, AC-FUTURE-013, AC-FUTURE-014, AC-FUTURE-015, AC-FUTURE-016, AC-FUTURE-017, AC-FUTURE-018, AC-FUTURE-019, AC-FUTURE-020, AC-FUTURE-021
- Tests: TC-FUTURE-001, TC-FUTURE-002, TC-FUTURE-003, TC-FUTURE-004, TC-FUTURE-005, TC-FUTURE-006, TC-FUTURE-007, TC-FUTURE-008, TC-FUTURE-009, TC-FUTURE-010, TC-FUTURE-011, TC-FUTURE-012, TC-FUTURE-013
- Actor: specs-creator
- Source: code
- Summary: Created the future spec from future.go, future_test.go, README.md, and git log; baseline of 13 functional and 2 non-functional requirements with 21 acceptance criteria, 3 covered and 10 missing test cases, and 8 open questions.
- Details: This is the first spec in a new store (no prior specs-index.md or specs-history.md). Highest-risk findings: Race self-deadlocks on its first settle callback, finally observes FinalUnknown for asynchronous execution, a promise with no success handler panics on resolve, README documents a SetCatch signature the code does not accept, and the repository has no go.mod so the package cannot be built or tested in this environment.

## HIST-0002 — 2026-10-23T11:08:12Z
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.1
- Requirements: NFR-FUTURE-003, AC-FUTURE-022, AC-FUTURE-023, AC-FUTURE-024
- Tests: TC-FUTURE-014, TC-FUTURE-015, TC-FUTURE-016
- Actor: specs-creator
- Source: instruction
- Summary: Added a build requirement for the module manifest and a make test entry point, at the user's instruction, to settle Q-FUTURE-004.
- Details: NFR-FUTURE-003 requires go.mod with module path github.com/noypi/future, go.sum pinning the two already-imported dependencies and their transitive requirements, and a Makefile test target that runs the suite. No existing requirement, acceptance criterion, or ID was changed; the only edits to existing text were the build scope, the verification table, the code map, and the Q-FUTURE-004 status (needs-decision to needs-evidence, with the decision recorded).

## HIST-0003 — 2026-10-24T06:33:15Z
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.2
- Requirements: NFR-FUTURE-001, NFR-FUTURE-002, NFR-FUTURE-003, AC-FUTURE-019, AC-FUTURE-020, AC-FUTURE-021, AC-FUTURE-022, AC-FUTURE-023, AC-FUTURE-024
- Tests: TC-FUTURE-002, TC-FUTURE-013, TC-FUTURE-014, TC-FUTURE-015, TC-FUTURE-016
- Actor: specs-creator
- Source: instruction
- Summary: Recorded the /verify-nfr run at 2026-10-24T06:33:15Z: added Method and Check to the five NFR verification rows (TC-FUTURE-002, TC-FUTURE-013 to TC-FUTURE-016), set TC-FUTURE-002 to missing for an environmental read-only GOTMPDIR failure, and corrected TC-FUTURE-016 Method from build to test per the approved decision.
- Details: Spec version 1.1 to 1.2; index Verified 5 to 4. The audit evidence is workflow.local/nfr-audit.json (exit code 1): 3 NFRs, 0 pass, 2 fail, 1 manual; 5 rows, 4 fail, 1 manual. TC-FUTURE-002 is missing although TestFuture_resolveFuncSignatures asserts its criteria, because go test cannot create its build dir under GOTMPDIR=/media/saif/saif_dev/tmp/go/ (read-only filesystem); TC-FUTURE-014 fails the same way. TC-FUTURE-015 and TC-FUTURE-016 fail for the genuinely missing go.sum and Makefile. TC-FUTURE-016 keeps Check: make test. No requirement statement, acceptance criterion, or ID was changed, and no other Status cell moved.

## HIST-0004 — 2026-10-24T20:11:50+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.3
- Tests: TC-FUTURE-012
- Actor: specs-creator
- Source: instruction
- Summary: Reconciled the Race descriptive records with the implemented fix from run ac-future-018 (commit e5783cf): FR-FUTURE-013 Notes now name the buffered result channel plus sync.Once and race_test.go, API-FUTURE-009 Errors drops "blocks indefinitely in the current implementation", Q-FUTURE-003 is marked resolved with red/green evidence, and TC-FUTURE-012 records race_test.go with Method test, the green command, and Status covered.
- Details: Spec version 1.2 to 1.3; index Verified 4 to 5 (FR-FUTURE-013 gains its first covered row). The covered claim is the run checkout /media/saif/EXTRA/workspaces/future/.worktrees/ac-future-018/_run at commit e5783cf: go test -count=1 -v -timeout 60s -run '^TestRace_' . was 3/3 green (ok github.com/noypi/future 1.203s) and the full suite go test -count=1 -v ./... was 6/6 green, exit 0; the same tests failed 3/3 with "Race did not return within 3s (deadlock)" before the fix. The workspace root still has no go.mod until NFR-FUTURE-003 lands on master, so no root-level run is claimed, and scripts/nfr-audit does not evaluate TC-FUTURE-012. No FR/NFR/AC/rule requirement text, acceptance bullet, ID, Data Model row, or Code Map row was changed; the Verification intro gained only the TC-FUTURE-012 provenance note. NFR-FUTURE-002 remains unverified and every NFR row's Status is unchanged.


## HIST-0005 — 2026-10-24T20:33:04+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.4
- Requirements: NFR-FUTURE-003
- Tests: TC-FUTURE-014, TC-FUTURE-015, TC-FUTURE-016
- Actor: specs-creator
- Source: instruction
- Summary: Recorded the 2026-10-24T12:31:39Z nfr-audit pass for NFR-FUTURE-003 after it landed: TC-FUTURE-014, TC-FUTURE-015, and TC-FUTURE-016 moved from missing to covered, and the Verification intro was corrected to name that run.
- Details: Spec version 1.3 to 1.4; index Verified 5 to 6 (NFR-FUTURE-003 gains its first covered row). Evidence is workflow.local/nfr-audit.json, checked_at 2026-10-24T12:31:39Z, exit 0, summary "3 NFRs — 2 pass, 0 fail, 0 external, 1 manual, 0 unknown". Row evidence: TC-FUTURE-014 (go.mod, Method build, Check go build ./...) verdict pass exit=0; TC-FUTURE-015 (go.sum, Method build, Check go mod verify) verdict pass exit=0; TC-FUTURE-016 (Makefile, Method test, Check make test) verdict pass exit=0. TC-FUTURE-013 stays missing (manual, no procedure executed). TC-FUTURE-002 also passed in this run (exit=0) but stays missing because it covers NFR-FUTURE-001 and is outside the requested NFR-FUTURE-003 scope, for a separate full pass. No requirement statement, acceptance criterion, Method, Check, File, Covers, or other Status cell changed.

## HIST-0006 — 2026-10-24T20:36:38+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.5
- Requirements: NFR-FUTURE-001, NFR-FUTURE-003
- Tests: TC-FUTURE-002, TC-FUTURE-014, TC-FUTURE-015, TC-FUTURE-016
- Actor: specs-creator
- Source: instruction
- Summary: Recorded the full /verify-nfr all audit run at 2026-10-24T12:34:47Z: TC-FUTURE-002 moved from missing to covered, and the stale records were reconciled with the landed NFR-FUTURE-003.
- Details: Spec version 1.4 to 1.5; index Verified 6 to 7 (NFR-FUTURE-001 gains its first covered row through TC-FUTURE-002). Evidence is workflow.local/nfr-audit.json, checked_at 2026-10-24T12:34:47Z, exit 0, summary "3 NFRs — 2 pass, 0 fail, 0 external, 1 manual, 0 unknown". Row evidence: TC-FUTURE-002 (NFR-FUTURE-001, Method test, Check go test -run TestFuture_resolveFuncSignatures ./...) verdict pass exit=0; TC-FUTURE-014 (go.mod, Method build, Check go build ./...) verdict pass exit=0; TC-FUTURE-015 (go.sum, Method build, Check go mod verify) verdict pass exit=0; TC-FUTURE-016 (Makefile, Method test, Check make test) verdict pass exit=0; TC-FUTURE-013 stays missing (manual, no procedure executed). Stale records corrected: the Verification intro no longer claims a read-only GOTMPDIR failure or points at Q-FUTURE-004; the TC-FUTURE-012 provenance paragraph now records that NFR-FUTURE-003 landed (PR #1 merged as 095bc13) and that go.mod, go.sum, and the Makefile are at the repository root, while keeping its evidence in the ac-future-018 run checkout; the Code Map build rows state the files are present and what they carry; Q-FUTURE-004 moved from needs-evidence to resolved with the audit evidence. TC-FUTURE-012 remains covered and every other Status cell is unchanged. No requirement statement, acceptance criterion, Method, Check, File, Covers, ID, Data Model row, Rule, interface, or error was changed.

## HIST-0007 — 2026-10-24T21:30:36+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.6
- Requirements: FR-FUTURE-008, AC-FUTURE-013
- Tests: TC-FUTURE-008
- Actor: specs-creator
- Source: instruction
- Summary: Declared how TC-FUTURE-008 is proven after run ac-future-013 (PR #3, merged as 1f71353) added future_state_test.go: File is future_state_test.go, Method is test, and Check is the TestPromiseState_* command; Status deliberately stays missing pending an nfr-audit run --all-rows pass.
- Details: Spec version 1.5 to 1.6; index Version 1.5 to 1.6, with Verified still 7 and Requirements still 15 because the row is not covered. The test file is package future and holds TestPromiseState_UnknownUntilCallback, TestPromiseState_ResolvedBeforeSuccessHandlers, and TestPromiseState_RejectedBeforeFailureHandlers, which were 3/3 green and 9/9 with the full suite in run ac-future-013 (commit 76f5d47), but a green suite is not a row pass (N4.6), so Status stays missing. No requirement statement, acceptance criterion, Covers, or another row's File, Method, or Check changed; the Verification intro prose was left unchanged because it still describes covered correctly.

## HIST-0008 — 2026-10-24T21:32:13+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.7
- Requirements: FR-FUTURE-008, AC-FUTURE-013
- Tests: TC-FUTURE-008
- Actor: specs-creator
- Source: instruction
- Summary: Reconciled verification row TC-FUTURE-008 with the 2026-10-24T13:31:05Z nfr-audit run --all-rows: Status moved from missing to covered on verdict=pass, detail=exit=0; spec version 1.6 to 1.7 and index Verified 7 to 8.
- Details: Audit evidence is workflow.local/nfr-audit.json (checked_at 2026-10-24T13:31:05Z), workspace and code root /media/saif/EXTRA/workspaces/future on branch feat/test-branch at 1f71353, scope all, 16 rows, exit code 3 with no row failed (6 pass, 1 manual, 9 unknown). TC-FUTURE-008 (requirement FR-FUTURE-008, Method test, Check go test -count=1 -run '^TestPromiseState_' .) is verdict pass, detail exit=0. TC-FUTURE-002, TC-FUTURE-012, TC-FUTURE-014, TC-FUTURE-015, and TC-FUTURE-016 also passed exit=0; TC-FUTURE-013 stays manual with Status missing. The nine rows that declare no Method are unknown and were left untouched, including TC-FUTURE-001 and TC-FUTURE-003, which keep their recorded covered Status pending an owner Method decision. The Verification intro now cites this run and its counts, and the TC-FUTURE-012 exception note was corrected to say nfr-audit reports that FR row only for information while the recorded run at e5783cf remains its evidence. No requirement statement, acceptance criterion, Covers, File, Method, or Check value changed.

## HIST-0009 — 2026-10-24T22:39:01+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.8
- Requirements: FR-FUTURE-009
- Tests: TC-FUTURE-009
- Actor: specs-creator
- Source: instruction
- Summary: Declared how TC-FUTURE-009 (AC-FUTURE-014) is proven after run ac-future-014 (commit bb61245, PR #5 merged as 72fc530) added three tests to future_test.go: File names TestFuture_DefaultCatchLogsRecoveredValue, TestFuture_SetCatchReplacesDefault, and TestFuture_SetFinallyReplacesDefault; Method is test; Check is go test -count=1 -run '^TestFuture_' .; Status deliberately stays missing pending a passing nfr-audit run --all-rows.
- Details: Spec version 1.7 to 1.8; index Version 1.7 to 1.8, with Verified still 8, Requirements still 15, and Updated still 2026-10-24 because the row is not covered. The three cited tests sit at future_test.go lines 174, 207, and 238, and workflow.local/runs/ac-future-014/test-report.md records PASS: 3/3 on the targeted command (go test -count=1 -run '^TestFuture_' .) and 9/9 on the full suite, exit 0. The most recent nfr-audit run reported TC-FUTURE-009 as unknown with detail "unknown method: —" solely because the row declared no File/Method/Check; it has never executed. The Verification intro's no-Method row count moved from nine to eight (TC-FUTURE-001, TC-FUTURE-003 to TC-FUTURE-007, TC-FUTURE-010 to TC-FUTURE-011); the cited 2026-10-24T13:31:05Z run and all other counts and provenance are unchanged and no new run was invented. The Check deliberately runs every TestFuture_* in the package (not a pipe-separated regex) so it contains no |, $, ;, &, <, >, backtick, or newline, satisfying N4.2. No requirement statement, acceptance criterion, Covers value, other row's File/Method/Check/Status, or ID was changed.

## HIST-0010 — 2026-10-24T22:40:32+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.9
- Requirements: FR-FUTURE-009
- Tests: TC-FUTURE-009
- Actor: specs-creator
- Source: instruction
- Summary: Reconciled verification row TC-FUTURE-009 with the 2026-10-24T14:39:29Z nfr-audit run --all-rows: Status moved from missing to covered on verdict=pass, detail=exit=0; spec version 1.8 to 1.9 and index Verified 8 to 9.
- Details: Audit evidence is workflow.local/nfr-audit.json (checked_at 2026-10-24T14:39:29Z), workspace and code root /media/saif/EXTRA/workspaces/future at HEAD 13c22ea, scope all, 16 rows, exit code 3 because 8 rows are unknown and no row failed (7 pass, 0 fail, 0 external, 1 manual, 8 unknown). TC-FUTURE-009 (requirement FR-FUTURE-009, kind fr, Method test, Check go test -count=1 -run '^TestFuture_' .) is verdict pass, detail exit=0; its File, Method, Check, and Covers cells are byte-identical. The Verification intro now cites the 14:39:29Z run and its counts, naming the passing rows TC-FUTURE-002, TC-FUTURE-008, TC-FUTURE-009, TC-FUTURE-012, TC-FUTURE-014, TC-FUTURE-015, and TC-FUTURE-016; TC-FUTURE-013 stays manual, the eight rows that declare no Method (TC-FUTURE-001, TC-FUTURE-003 to TC-FUTURE-007, TC-FUTURE-010 to TC-FUTURE-011) keep their recorded Status, and the TC-FUTURE-012 provenance paragraph is unchanged. A 1.9 changelog entry was added. Index SPEC-FUTURE moved to Version 1.9 with Verified 9 (TC-FUTURE-001, -002, -003, -008, -009, -012, -014, -015, -016) and Requirements still 15; its Notes paragraph now cites the 14:39:29Z run and explains the 8 to 9 move through FR-FUTURE-009. No requirement statement, acceptance criterion, Covers, File, Method, Check, other row's Status, or FR/NFR lifecycle status changed, and workflow.local/, production code, tests, and config were not touched.

## HIST-0011 — 2026-10-24T23:16:39+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.10
- Requirements: FR-FUTURE-012
- Tests: TC-FUTURE-011
- Actor: specs-creator
- Source: instruction
- Summary: Declared how TC-FUTURE-011 (AC-FUTURE-017) is proven after run ac-future-017 published TestFuture_WaitBlocksForAsyncExec and TestFuture_WaitReturnsAfterSyncExec in future_test.go: File names those two tests, Method is test, and Check is go test -count=1 -run '^TestFuture_Wait' .; Status deliberately stays missing for the run's Close to reconcile from the audit.
- Details: Spec version 1.9 to 1.10; index SPEC-FUTURE Version 1.9 to 1.10 with Updated still 2026-10-24, Requirements still 15 and Verified still 9 because the row is not covered. Evidence: workflow.local/runs/ac-future-017/test-report.md (verdict PASS) records go test -count=1 -run '^TestFuture_Wait' . as 2/2 PASS, exit 0, and both tests live in future_test.go in the run worktree .worktrees/ac-future-017/_run (run status INTEGRATING); the workspace root future_test.go does not carry them yet, so this is a declaration of the package test file they land in, not a claim that the root file already holds them. The Verification intro's no-Method row count moved from eight to seven (TC-FUTURE-001, TC-FUTURE-003 to TC-FUTURE-007, TC-FUTURE-010); the cited 2026-10-24T14:39:29Z nfr-audit run and its counts (7 pass, 1 manual, 8 unknown) and provenance are unchanged and no new audit run was invented. No requirement statement, acceptance criterion, Covers value, Status cell, other row's File/Method/Check, or ID was changed, and production code, tests, plan.md, and workflow.local/ were not modified.

## HIST-0012 — 2026-10-24T23:19:10+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.11
- Requirements: FR-FUTURE-012
- Tests: TC-FUTURE-011
- Actor: specs-creator
- Source: instruction
- Summary: Reconciled TC-FUTURE-011 with the 2026-10-24T15:18:05Z nfr-audit run --all-rows: Status moved from missing to covered on verdict=pass, detail=exit=0, giving FR-FUTURE-012 its first covered row; spec Version 1.10 to 1.11 and index Verified 9 to 10.
- Details: Audit evidence is workflow.local/nfr-audit.json (checked_at 2026-10-24T15:18:05Z, code root /media/saif/EXTRA/workspaces/future/.worktrees/ac-future-017/_run, scope all, 16 rows, rows_pass: 8, rows_fail: 0, rows_external: 0, rows_manual: 1, rows_unknown: 7). TC-FUTURE-011 (requirement FR-FUTURE-012, Covers AC-FUTURE-017, Method test, Check go test -count=1 -run '^TestFuture_Wait' .) is verdict pass, detail exit=0, so TC-FUTURE-011 missing → covered and FR-FUTURE-012 gains its first covered row. TC-FUTURE-013 stays manual/missing. The seven unknown rows TC-FUTURE-001, TC-FUTURE-003, TC-FUTURE-004, TC-FUTURE-005, TC-FUTURE-006, TC-FUTURE-007, and TC-FUTURE-010 were left unchanged. The Verification intro now cites the 15:18:05Z run and its counts (8 pass, 1 manual, 7 unknown), naming TC-FUTURE-002, TC-FUTURE-008, TC-FUTURE-009, TC-FUTURE-011, TC-FUTURE-012, TC-FUTURE-014, TC-FUTURE-015, and TC-FUTURE-016 as passed; the no-Method sentence still names seven rows. A 1.11 changelog entry was added. Index SPEC-FUTURE moved to Version 1.11 with `Verified` 9 → 10 and Requirements still 15; its Notes now cite this run and explain the 9 to 10 move through FR-FUTURE-012. Only the TC-FUTURE-011 Status cell, the provenance prose, the version, the changelog, this history entry, and the index count changed: no requirement, acceptance criterion, Covers, File, Method, Check, or other row's Status was altered, and no requirement Status field was set to a delivery state.

## HIST-0013 — 2026-10-24T23:31:45+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.12
- Tests: TC-FUTURE-010
- Actor: specs-creator
- Source: instruction
- Summary: Declared how TC-FUTURE-010 (AC-FUTURE-015) is proven after run ac-future-015 published eleven TestRecovered_* tests across future_state_test.go and future_test.go: File names both files and the eleven tests, Method is test, and Check is go test -count=1 -run '^TestRecovered_' .; Status deliberately stays missing for the run's Close to derive. Q-FUTURE-002 moved from needs-evidence to resolved on that run's finally evidence; spec Version 1.11 to 1.12.
- Details: Evidence is workflow.local/runs/ac-future-015/unit-tests.md (verify verdict GREEN): the batch command reports 11 run, 11 passed, 0 failed, exit 0, and the two test-file sha256 values are published there. The File inventory is future_state_test.go (TestRecovered_StateExecutorPanic, TestRecovered_StateHandlerPanic, TestRecovered_AsyncStateExecutorPanic, TestRecovered_AsyncStateHandlerPanic) and future_test.go (TestRecovered_ErrorAndValueExecutorPanic, TestRecovered_ErrorAndValueSuccessHandlerPanic, TestRecovered_ErrorAndValueFailureHandlerPanic, TestRecovered_AsyncErrorAndValueExecutorPanic, TestRecovered_AsyncErrorAndValueSuccessHandlerPanic, TestRecovered_AsyncErrorAndValueFailureHandlerPanic, TestRecovered_FutureAsyncPanic); future_acceptance_test.go, added by the run's acceptance tester, is not part of this 11-test row. Check is a single command with no shell metacharacters (N4.2). Status stays missing because a green suite is not an executed row check (N4.4/N4.6); the run's Close audit derives it, and index Verified stays 10 with Requirements still 15. Q-FUTURE-002 evidence: post-fix the deferred settlement block (recover, catch, finally) runs inside the execution closure on the executor goroutine for exec(true), so finally runs after settlement and Wait returns after Done; the async TestRecovered_AsyncState* tests record "finally observation: calls=1 state=3 args=[]" after Wait, state=3 being FinalRecovered. FR-FUTURE-011, AC-FUTURE-016, and TC-FUTURE-007 were not touched. The Verification intro's no-Method count moved from seven rows to six (TC-FUTURE-001, TC-FUTURE-003 to TC-FUTURE-007); the cited 2026-10-24T15:18:05Z ac-future-017 run, its counts (8 pass, 1 manual, 7 unknown), and its provenance are unchanged, with a parenthetical noting that run predates this declaration. A 1.12 changelog entry was added and specs-index.md moved to Version 1.12 with Updated 2026-10-24. No requirement statement, acceptance criterion, Covers value, other row's File/Method/Check/Status, or ID was changed, no requirement Status field was set to a delivery state, and production code, tests, plan.md, and workflow.local/ were not modified.

## HIST-0014 — 2026-10-24T23:33:48+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.13
- Requirements: FR-FUTURE-010
- Tests: TC-FUTURE-010
- Actor: specs-creator
- Source: instruction
- Summary: Reconciled TC-FUTURE-010 with the 2026-10-24T15:32:29Z nfr-audit run --all-rows: Status moved from missing to covered on verdict=pass, detail=exit=0, giving FR-FUTURE-010 its first covered row; spec Version 1.12 to 1.13 and index Verified 10 to 11.
- Details: Audit evidence is workflow.local/nfr-audit.json (checked_at 2026-10-24T15:32:29Z, code root /media/saif/EXTRA/workspaces/future/.worktrees/ac-future-015/_run, scope all, 16 rows, rows_pass: 9, rows_fail: 0, rows_external: 0, rows_manual: 1, rows_unknown: 6; nfr-audit status reports stale: no, exit 0). TC-FUTURE-010 (requirement FR-FUTURE-010, Covers AC-FUTURE-015, Method test, Check go test -count=1 -run '^TestRecovered_' .) is verdict pass, detail exit=0, so TC-FUTURE-010 missing -> covered and FR-FUTURE-010 gains its first covered row. Its File, Method, Check, and Covers cells are byte-identical to the 1.12 declaration. TC-FUTURE-013 stays manual/missing; the six unknown rows TC-FUTURE-001, TC-FUTURE-003, TC-FUTURE-004, TC-FUTURE-005, TC-FUTURE-006, and TC-FUTURE-007 were left unchanged. The Verification intro now cites the 15:32:29Z ac-future-015 run checkout and its counts (9 pass, 1 manual, 6 unknown), naming the passing rows TC-FUTURE-002, TC-FUTURE-008, TC-FUTURE-009, TC-FUTURE-010, TC-FUTURE-011, TC-FUTURE-012, TC-FUTURE-014, TC-FUTURE-015, and TC-FUTURE-016; the obsolete parenthetical saying the cited run predated the row's Method declaration was dropped, and the six-row no-Method sentence (TC-FUTURE-001, TC-FUTURE-003 to TC-FUTURE-007) is unchanged. A 1.13 changelog entry was added naming the audit run, checkout, and counts and the single Status change. Index SPEC-FUTURE moved to Version 1.13 with Verified 10 to 11, Requirements still 15, and Updated still 2026-10-24; its Notes now cite this run and explain the 10 to 11 move through FR-FUTURE-010/TC-FUTURE-010. No requirement statement, acceptance criterion, Covers, File, Method, Check, or other row's Status was altered, no requirement Status field was set to a delivery state (N4.7), and workflow.local/, production code, tests, plan.md, and git state were not touched (N4.3-N4.6).

## HIST-0015 — 2026-10-24T23:57:56+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.14
- Requirements: FR-FUTURE-014, FR-FUTURE-015, AC-FUTURE-025, AC-FUTURE-026, AC-FUTURE-027, AC-FUTURE-028, AC-FUTURE-029, AC-FUTURE-030
- Tests: TC-FUTURE-017, TC-FUTURE-018, TC-FUTURE-019
- Actor: specs-creator
- Source: instruction
- Summary: Added the two requested improvement features: handler-less settlement with executor-typed wrappers (`FR-FUTURE-014`) and outcome accessors `State`/`Results`/`Err` (`FR-FUTURE-015`), plus untyped-`nil` handling (`AC-FUTURE-025`); added three pending test rows and non-normative `/build` delivery estimates. Feature 1 adds no requirement because `FR-FUTURE-010`/`-011`/`-012` and `NFR-FUTURE-002` already own it.
- Details: `AC-FUTURE-026` to `AC-FUTURE-030` allocated; `API-FUTURE-010` to `API-FUTURE-012` added; `RULE-FUTURE-004` replaced with the executor-typed-wrapper rule; `Q-FUTURE-005` and `Q-FUTURE-008` resolved on the user's instruction; `CONFLICT-FUTURE-001` and `CONFLICT-FUTURE-002` recorded. Delivery Estimates: Feature 1 45–65 min, Feature 2 70–100 min under the `/build` stages; neither full feature fits 30 min, and the model "DeepSeek 4.1 (low)" is not measured here, so the section is explicitly non-normative. Index Version 1.14, Requirements 15 -> 17 (13 -> 15 FR), Verified unchanged at 11. No existing requirement, acceptance criterion, `covered` row, or `Status` changed; production code, tests, and `workflow.local/` were not modified.

## HIST-0016 — 2026-10-25T00:41:57+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.15
- Requirements: AC-FUTURE-026, AC-FUTURE-027, AC-FUTURE-028, AC-FUTURE-029, AC-FUTURE-030
- Tests: TC-FUTURE-017, TC-FUTURE-018, TC-FUTURE-020, TC-FUTURE-021, TC-FUTURE-022
- Actor: specs-creator
- Source: instruction
- Summary: Split the combined verification rows into one row per acceptance criterion, each in its own test file, so single-AC `/build` selector runs (`AC-FUTURE-025`, `-026`, `-028`) can close their own rows; spec 1.14 to 1.15, index 1.15.
- Details: `TC-FUTURE-017` now covers only `AC-FUTURE-026` in `future_settle_test.go`; `TC-FUTURE-018` covers only `AC-FUTURE-028` in `future_accessor_test.go`; `TC-FUTURE-019` (`AC-FUTURE-025`, `future_test.go`) is unchanged; new `TC-FUTURE-020` covers `AC-FUTURE-027` (`future_reject_test.go`), `TC-FUTURE-021` covers `AC-FUTURE-029` (`future_results_test.go`), and `TC-FUTURE-022` covers `AC-FUTURE-030` (`future_err_test.go`). All six stay `missing` and each `Check` targets one exact test command with no shell metacharacters (N4.2). The Verification intro and the Delivery Estimates note were updated; the `Covers`, `File`, and `Check` cells of `TC-FUTURE-017` and `TC-FUTURE-018` and the three new rows were rewritten as described. No requirement statement, acceptance criterion, `covered` row, or requirement `Status` changed; Requirements stays 17 and Verified stays 11. No production code, test file, or `workflow.local/` artifact was modified, and `nfr-audit` was not run.

## HIST-0017 — 2026-10-25T12:12:17+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.16
- Requirements: FR-FUTURE-015, AC-FUTURE-030
- Tests: TC-FUTURE-022
- Actor: specs-creator
- Source: instruction
- Summary: Close reconciliation of `TC-FUTURE-022` (`AC-FUTURE-030`, parent `FR-FUTURE-015`) from `missing` to `covered` on the 2026-10-25T04:10:40Z `nfr-audit run --all-rows` (22 rows: 15 pass, 6 unknown, 1 manual, 0 fail) over the `ac-future-030` run checkout at commit `ccaacd0`; `FR-FUTURE-015` gains its first covered row, spec 1.15 to 1.16, index Verified 11 to 12.
- Details: Run `ac-future-030`, branch `run/ac-future-030`, PR #11 (base `feat/test-plan-with-interfaces`). Audit evidence is `workflow.local/nfr-audit.json` (`nfr-audit status` reports `stale: no`, `uncovered_acs: 0`); code root `/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-030/_run`, scope `all`. The row's `Covers` (`AC-FUTURE-030`), `File` (`future_err_test.go` (`TestOutcome_Err`)), `Method` (`test`), and `Check` (`go test -count=1 -run '^TestOutcome_Err' .`) cells are byte-identical to 1.15; only its `Status` cell changed. `TestOutcome_Err` exists in the run checkout and `go test -count=1 -run '^TestOutcome_Err' .` exits 0 there; the run added exactly `future.go` (unexported `err error` field plus `func (q *Promise) Err() error`) and `future_err_test.go`. The five sibling rows `TC-FUTURE-017` to `TC-FUTURE-021` also report `verdict=pass`, `detail=exit=0`, but those passes are vacuous and all five deliberately stay `missing`: `future_settle_test.go`, `future_accessor_test.go`, `future_reject_test.go`, and `future_results_test.go` do not exist in that checkout, `future_test.go` contains no `TestFuture_UntypedNil*` function, and `go test -run '<regex>' .` exits 0 when nothing matches, so promoting them would claim coverage for tests that do not exist. The Verification intro now cites the 2026-10-25T04:10:40Z run, its counts, and the vacuous-pass reason. No requirement statement, acceptance criterion, `Covers`, `File`, `Method`, `Check`, or other row's `Status` was altered; the Data Model table is untouched (the new unexported `err` field is a known follow-up outside this Close); no production code, test file, or `workflow.local/` artifact was modified.

## HIST-0018 — 2026-10-25T13:28:11+08:00
- Action: created
- Spec: specs-unit-tests.md
- Spec ID: SPEC-UNIT-TESTS
- Version: 1.0
- Requirements: FR-UNIT-TESTS-001, AC-UNIT-TESTS-001, AC-UNIT-TESTS-002
- Tests: TC-UNIT-TESTS-001, TC-UNIT-TESTS-002
- Actor: specs-creator
- Source: instruction
- Summary: Created the table-driven unit-test spec at the user's instruction: one table per behavior, every case a named `t.Run` subtest, single-case and non-unit tests exempt; index 1.16 to 1.17 with SPEC-UNIT-TESTS (Requirements 1, Verified 0).
- Details: New spec `SPEC-UNIT-TESTS`, purpose `unit-tests`. `AC-UNIT-TESTS-001` requires a multi-case unit test to declare its cases in one table carrying `name`, inputs, and expected result and to run each row through `t.Run(tt.name, ...)`; `AC-UNIT-TESTS-002` exempts single-case tests and non-unit tests and excludes benchmarks, examples, and fuzz targets. Two manual verification rows (`TC-UNIT-TESTS-001`, `TC-UNIT-TESTS-002`) with numbered review procedures, both `Status: missing` because no human has run them (N4.3). Scope limits the rule to unit tests authored or modified from 2026-10-25 onward and migrates existing tests only when they are next modified; the migration deadline is recorded as `Q-UNIT-TESTS-001` (`unverified`). No existing requirement, acceptance criterion, verification row, or `Status` in `specs-future.md` changed; no production code, test file, or `workflow.local/` artifact was modified.

## HIST-0019 — 2026-10-25T14:08:08+08:00
- Action: updated
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.17
- Requirements: FR-FUTURE-015, AC-FUTURE-029
- Tests: TC-FUTURE-021
- Actor: specs-creator
- Source: instruction
- Summary: Close reconciliation of `TC-FUTURE-021` (`AC-FUTURE-029`, parent `FR-FUTURE-015`) from `missing` to `covered` on the 2026-10-25T06:07:02Z `nfr-audit run --all-rows` (22 rows: 15 pass, 6 unknown, 1 manual, 0 fail) over the `ac-future-029` run checkout at head `2b4706f`; spec 1.16 to 1.17, index Verified 12 to 13.
- Details: Run `ac-future-029`, branch `run/ac-future-029`, PR #13 (base `feat/test-plan-with-prs`). Audit evidence is `workflow.local/nfr-audit.json` (`nfr-audit status` reports `stale: no`); code root `/media/saif/EXTRA/workspaces/future/.worktrees/ac-future-029/_run`, scope `all`. The row's `Covers` (`AC-FUTURE-029`), `File` (`future_results_test.go` (`TestOutcome_Results`)), `Method` (`test`), and `Check` (`go test -count=1 -run '^TestOutcome_Results' .`) cells are byte-identical to 1.16; only its `Status` cell changed. `TestOutcome_Results` exists in the run checkout, `go test -count=1 -run '^TestOutcome_Results' .` exits 0 there, and the verbose run reports `--- PASS: TestOutcome_Results` with eight named subtests; the acceptance tester independently reproduced the behaviour with a 12-test external probe and confirmed the row non-vacuous via first-value-only and reversed mutants. The run added `future.go` (`Results()` accessor, +17) and `future_results_test.go`. The four sibling rows `TC-FUTURE-017` to `TC-FUTURE-020` also report `verdict=pass`, `detail=exit=0`, but those passes are vacuous and all four deliberately stay `missing`: `future_settle_test.go`, `future_accessor_test.go`, and `future_reject_test.go` do not exist in the run checkout, `future_test.go` has no `TestFuture_UntypedNil*` function, and `go test -run '<regex>' .` exits 0 when nothing matches, so promoting them would claim coverage for tests that do not exist. `TC-FUTURE-022` was already `covered` and is unchanged. The Verification intro now cites the 06:07:02Z run, its counts, and the vacuous-pass reason. No requirement statement, acceptance criterion, `Covers`, `File`, `Method`, `Check`, or other row's `Status` was altered; the Data Model table and `specs-unit-tests.md` are untouched; no production code, test file, or `workflow.local/` artifact was modified.

## HIST-0020 — 2026-10-25T18:20:00+08:00
- Action: added
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.18
- Requirements: FR-FUTURE-016, AC-FUTURE-031, AC-FUTURE-032, AC-FUTURE-033
- Tests: TC-FUTURE-023, TC-FUTURE-024, TC-FUTURE-025
- Actor: specs-creator
- Source: instruction
- Summary: Added `FR-FUTURE-016` (a promise settles at most once; the first settlement wins) with three acceptance criteria and three pending verification rows in `future_settle_once_test.go`, rewrote `RULE-FUTURE-001`, noted the per-promise latch in `RULE-FUTURE-003`/`Q-FUTURE-007`, extended the `NFR-FUTURE-002` note, and added the Feature 3 delivery estimate; spec 1.17 to 1.18, index Requirements 17 to 18, Verified stays 13.
- Details: The requirements were written at the user's instruction to carry a small, concurrency-shaped enhancement that a weak model is unlikely to complete in four attempts and that `/build` can deliver as one batch. `AC-FUTURE-031` covers the sequential second settle (normal return, no outcome change, second chain not invoked); `AC-FUTURE-032` covers two callbacks invoked concurrently from different goroutines settling exactly once without a data race on the settlement fields; `AC-FUTURE-033` requires a handler that reads `Results()`/`Err()` to neither deadlock nor panic, which rules out holding the settlement lock across handler invocation. The notes record that the wrappers are still created lazily until `FR-FUTURE-014` lands, so a settlement test registers both a success and a failure handler, and that `State()` is not required by these criteria. The three rows name `future_settle_once_test.go`, which does not exist yet, and all three are `missing` — a `go test -run` that matches nothing exits 0 and must not promote them (the vacuous-pass hazard recorded in 1.17). No production code, test file, or `workflow.local/` artifact was modified; the Data Model table is untouched because the latch field is unexported and named at implementation.

## HIST-0021 — 2026-10-25T18:30:00+08:00
- Action: added
- Spec: specs-future.md
- Spec ID: SPEC-FUTURE
- Version: 1.19
- Requirements: FR-FUTURE-017, FR-FUTURE-018, AC-FUTURE-034, AC-FUTURE-035, AC-FUTURE-036, AC-FUTURE-037
- Tests: TC-FUTURE-026, TC-FUTURE-027, TC-FUTURE-028, TC-FUTURE-029
- Actor: specs-creator
- Source: instruction
- Summary: Added `FR-FUTURE-017` (a variadic handler receives the settled values packed into its variadic parameter) and `FR-FUTURE-018` (a registration call is atomic across the handlers it names) with four acceptance criteria and four pending verification rows; spec 1.18 to 1.19, index Requirements 18 to 20, Verified stays 13.
- Details: Written at the user's instruction to add two enhancements that run concurrently with `FR-FUTURE-016` and each fit a 38–48 minute `/build` run. `FR-FUTURE-017` records the reflection defect in `fitFnArgs`: it builds one argument per declared position and never packs a variadic tail, so `func(xs ...string)` receives a zero slice and loses every settled value; the fix is scoped to `fitFnArgs` so `wrappedFunc` keeps a plain `Call`. `FR-FUTURE-018` records that `OnSuccess`, `OnFail`, and `Then` append as they go, so an invalid argument leaves earlier handlers registered and the wrapper type initialised from a partial call; atomicity is required for `OnSuccess`, `OnFail`, and across both sides of `Then`, while nil stays skipped per `AC-FUTURE-025` and the tests use `42` until that criterion is implemented. `## Delivery Estimates` gained Feature 4 (40–48 min) and Feature 5 (38–45 min) plus a note recording that Features 3 to 5 edit disjoint regions of `future.go` (struct and wrappers; `fitFnArgs`; `Then`/`On`*), so they may run as three concurrent `/build` runs with merge order 3, 4, 5. No existing requirement, acceptance criterion, verification row, or `Status` changed; the four new rows are `missing` and a `go test -run` that matches nothing must not promote them. No production code, test file, or `workflow.local/` artifact was modified.
