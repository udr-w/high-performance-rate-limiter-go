# Summary

Describe what changed and why.

## Type of Change

- [ ] Bug fix
- [ ] Feature
- [ ] Documentation
- [ ] Test or benchmark
- [ ] Refactor
- [ ] Security hardening

## Behavior and Compatibility

- Public API impact:
- Algorithm or rate-limit semantics impact:
- Redis behavior impact:
- Backward compatibility notes:

## Verification

- [ ] `gofmt` run on changed Go files
- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] Redis integration tests run when Redis behavior changed
- [ ] Benchmarks run when hot paths changed

## Documentation

- [ ] README updated if public behavior changed
- [ ] CONTRIBUTING updated if development workflow changed
- [ ] SECURITY updated if vulnerability handling or security posture changed

## Notes for Reviewers

Call out areas that need careful review, such as concurrency, Redis Lua scripts, context handling, validation, or compatibility.
