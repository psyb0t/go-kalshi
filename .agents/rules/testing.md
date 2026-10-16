# Testing

Refines `~/.claude/rule-details/go/writing-tests.md`.

- **TST1** Tests are black-box, in `_test` packages, and use only the exported API.
- **TST2** Tests never call Kalshi. They decode literal frames (fixtures in `ws/testdata/`), write through a fake transport, and check signatures against generated test keys.
- **TST3** The README examples are compiled by `readme_test.go` and `ws/readme_test.go`. A README example change goes into the matching test in the same commit.
- **TST4** Every spec patch or overlay change that alters wire behavior gets a test with a frame shaped like what Kalshi sends.
- **TST5** Table tests use `testCases` and `tc`, call `t.Parallel()` in the test and in every subtest, and assert errors with `require.ErrorIs` against a sentinel.
- **TST6** `make test` runs `go test -race` across every package and must pass.
