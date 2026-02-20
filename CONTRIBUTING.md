# 🤝 CONTRIBUTING GUIDELINES

Thank you for your interest in contributing to PhantomProxy Pro!

## ⚠️ IMPORTANT NOTICE

**PhantomProxy Pro is a proprietary security tool.** By contributing, you agree to:

1. **Legal Use Only** — Your contributions will be used only for legitimate security testing
2. **No Malicious Features** — Do not submit code designed for malicious purposes
3. **Ethical Standards** — Follow our [Ethical Boundaries](docs/ETHICAL_BOUNDARIES.md)

## 📋 HOW TO CONTRIBUTE

### 1. Reporting Issues

**Found a bug?** Open an issue with:
- Clear description
- Steps to reproduce
- Expected vs actual behavior
- Screenshots/logs (if applicable)
- Environment details (OS, Python version, etc.)

### 2. Feature Requests

**Have an idea?** Open an issue with:
- Feature description
- Use case
- Benefits
- Potential implementation approach

### 3. Code Contributions

**Want to submit code?**

1. **Fork** the repository
2. **Create a branch** (`git checkout -b feature/AmazingFeature`)
3. **Make your changes**
4. **Test thoroughly**
5. **Commit** (`git commit -m 'Add AmazingFeature'`)
6. **Push** (`git push origin feature/AmazingFeature`)
7. **Open a Pull Request**

### 4. Documentation

**Improve docs?** We welcome:
- Typos fixes
- Clarifications
- New guides
- Translations

## 📝 CODE STYLE

### Python

```python
# Follow PEP 8
# Use type hints
# Add docstrings

def example_function(param: str) -> bool:
    """Example docstring."""
    return True
```

### Testing

```bash
# Run tests before submitting
pytest tests/

# Check code style
black modules/
flake8 modules/

# Type checking
mypy modules/
```

## 🔍 PULL REQUEST PROCESS

1. **Review** — Maintainers review your PR
2. **Tests** — Automated tests run
3. **Feedback** — You may be asked for changes
4. **Approval** — PR is approved
5. **Merge** — Changes are merged

## 📜 CODE OF CONDUCT

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone.

### Expected Behavior

- Be respectful
- Be constructive
- Be inclusive
- Accept constructive criticism

### Unacceptable Behavior

- Harassment
- Discrimination
- Personal attacks
- Trolling

### Enforcement

Violations will result in:
1. Warning
2. Temporary ban
3. Permanent ban

Report violations to: conduct@phantomseclabs.com

## 📚 RESOURCES

- [Project Overview](docs/PROJECT_OVERVIEW.md)
- [Ethical Boundaries](docs/ETHICAL_BOUNDARIES.md)
- [Security Policy](SECURITY.md)
- [License](LICENSE)

## 🎯 AREAS WE NEED HELP

### High Priority

- [ ] API Documentation (Swagger/OpenAPI)
- [ ] Unit Tests
- [ ] Integration Tests
- [ ] CI/CD Pipeline
- [ ] Mobile App (React Native)
- [ ] Desktop App (Electron)

### Medium Priority

- [ ] Multi-language Support (i18n)
- [ ] Template Library
- [ ] Knowledge Base
- [ ] Video Tutorials

### Low Priority

- [ ] UI/UX Improvements
- [ ] Performance Optimizations
- [ ] Additional Integrations

## 💬 QUESTIONS?

**Get in touch:**
- Email: dev@phantomseclabs.com
- GitHub Issues
- Discussions tab

---

**Thank you for contributing to PhantomProxy Pro!** 🚀

**© 2026 PhantomSec Labs**
