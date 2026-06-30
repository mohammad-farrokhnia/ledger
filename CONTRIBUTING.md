# Contributing to ledger

Thank you for your interest in contributing to `ledger`! As a high-consistency financial service, we mastertain strict code quality standards.

## 🌿 Branching Strategy

We follow a strict **Git Flow** workflow:

* **`master`**: Production-ready code. Locked. Use Pull Requests.
* **`develop`**: Integration branch. All features merge here first.
* **`feat/xxx`**: Feature branches. Branch off `develop`.
* **`fix/xxx`**: Bug fix branches. Branch off `develop`.

## 🛡️ Quality Gates

Your PR will only be merged if:

1.  **CI Passes**: All unit tests and linters (golangci-lint) must pass.
2.  **Coverage**: Code coverage must not decrease.
3.  **Review**: At least **1 approval** from a code owner is required for `master`.
4.  **Linear History**: We use "Squash and Merge" or "Rebase and Merge". No merge commits.

## 💻 Local Development

1.  **Setup**: `make up` (starts DB)
2.  **Lint**: `make lint` (checks your code against our strict rules)
3.  **Test**: `make test` (runs race detection)

## 📝 Commit Messages

Please follow [Conventional Commits](https://www.conventionalcommits.org/):
* `feat: add idempotency middleware`
* `fix: resolve race condition in wallet update`
* `docs: update swagger definition`