# Pull Request Template

## Description

<!-- Provide a detailed description of your changes -->

**Related Issue:** Fixes #

---

## Type of Change

<!-- Select one and delete others -->

- [ ] 🐛 Bug fix (non-breaking change that fixes an issue)
- [ ] ✨ New feature (non-breaking change that adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to change)
- [ ] 📝 Documentation update
- [ ] 🎨 Style/formatting changes
- [ ] ♻️ Code refactoring
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test addition/update
- [ ] 🔒 Security fix
- [ ] 🚀 Deployment/CI/CD update
- [ ] 🔧 Configuration change
- [ ] 📦 Dependency update
- [ ] 🌐 Internationalization
- [ ] 📈 Analytics/telemetry

---

## Testing

### Tests Performed

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] All existing tests pass

### Test Evidence

<!-- Provide screenshots, logs, or other evidence of testing -->

```
Paste logs or describe test results here
```

---

## Code Quality

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Code is properly commented
- [ ] No new warnings introduced
- [ ] No sensitive data exposed
- [ ] No hardcoded values (using config/env)

### Linting

```bash
# Python
black .
flake8 .
mypy .

# Go
go fmt ./...
go vet ./...
```

- [ ] Code passes linters

---

## Documentation

- [ ] README.md updated (if applicable)
- [ ] CHANGELOG.md updated
- [ ] Code comments added where needed
- [ ] API documentation updated (if applicable)
- [ ] User documentation updated (if applicable)

---

## Security Checklist

- [ ] No new security vulnerabilities introduced
- [ ] Input validation implemented
- [ ] Authentication/authorization checked
- [ ] No secrets or credentials in code
- [ ] SQL injection prevention (if applicable)
- [ ] XSS prevention (if applicable)
- [ ] CSRF protection (if applicable)

---

## Performance

- [ ] No performance regression
- [ ] Memory usage optimized
- [ ] Database queries optimized
- [ ] Caching implemented where appropriate

### Benchmarks

<!-- If applicable, provide before/after benchmarks -->

```
Before: 
After: 
```

---

## Screenshots

<!-- If UI changes, include before/after screenshots -->

| Before | After |
|--------|-------|
|        |       |

---

## Deployment Notes

### Migration Required

- [ ] Database migration needed
- [ ] Configuration changes required
- [ ] Environment variables to add
- [ ] Service restart required

### Rollback Plan

<!-- Describe how to rollback if needed -->



---

## Checklist

- [ ] I have read the [CONTRIBUTING.md](CONTRIBUTING.md) document
- [ ] My code follows the project's coding guidelines
- [ ] I have performed a self-review
- [ ] I have commented my code where appropriate
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix/feature works
- [ ] All new and existing tests pass
- [ ] Any dependent changes are merged and published

---

## Additional Context

<!-- Add any other context about the PR here -->



---

## Reviewers

<!-- Tag reviewers -->

@rpauts2

---

**Thank you for your contribution!** 🚀
