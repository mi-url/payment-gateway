# Engineering Standards — Payment Gateway

> **These are NON-NEGOTIABLE rules for every line of code in this project.**
> This is a financial system. An auditor can inspect this codebase at any time.

---

## 1. Documentation is Mandatory

- **Every exported function** has a GoDoc comment explaining WHAT it does and WHY.
- **Every file** has a header comment explaining its role in the system.
- **Every non-obvious decision** has an inline comment with the rationale.
- **No comment is better than a wrong comment.** If code changes, the comment changes with it.
- **API contracts** are documented with OpenAPI/Swagger — auto-generated from code annotations.
- **Database migrations** include a description comment at the top explaining what changes and why.

## 2. Zero Dead Code

- No commented-out code. Ever. Git history exists for that.
- No unused functions, variables, imports, or types. `go vet` and `staticcheck` enforce this.
- No TODO/FIXME/HACK comments in production branches. If it's not done, it's not merged.
- No experimental code in production. Experiments live in feature branches and get deleted after merge.

## 3. Zero Patches / No Band-Aids

- Every fix addresses the root cause, not the symptom.
- No "temporary" workarounds that become permanent. If a workaround is needed, it gets a tracking issue and a deadline.
- Refactor as you go. If adding a feature makes existing code messy, clean it in the same PR.
- Code must read like it was written by one disciplined engineer, not assembled from Stack Overflow.

## 4. Simplicity is a Feature

- Prefer explicit over clever. A junior engineer should understand the flow.
- One way to do things. No duplicate patterns for the same operation.
- Flat is better than nested. If a function has 4+ levels of indentation, refactor it.
- Small files, small functions, clear names. A function longer than 50 lines needs justification.

## 5. Audit-Ready at All Times

- The `main` branch is always deployable and clean.
- Every PR passes linting, tests, and security checks before merge.
- Sensitive data (keys, credentials, PII) never appears in logs, error messages, or version control.
- Database migrations are sequential, numbered, and reversible.
- Git history is clean: meaningful commit messages, no "fix typo" chains.

## 6. Testing

- Critical paths (charge flow, encryption, state machine) have unit tests.
- Integration tests exist for every bank adapter.
- Tests are documentation — reading a test should explain the behavior.
- No test depends on external services (mock everything in unit tests).

## 7. Dependency Hygiene

- Minimal dependencies. Every `go get` must be justified.
- No abandoned or unmaintained libraries.
- `go.sum` is committed. Reproducible builds always.
- Security advisories checked on every build (via `govulncheck`).

---

## Enforcement

These standards are enforced via:
- **CI/CD pipeline:** `go vet`, `staticcheck`, `golangci-lint`, `govulncheck` on every PR
- **Code review discipline:** Every merge follows these rules
- **Pre-commit hooks:** Format, lint, and verify before push
