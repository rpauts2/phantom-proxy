# Contributing to PhantomProxy v14.0

Thank you for contributing to PhantomProxy!

## Code of Conduct

- Be respectful
- Follow ethical guidelines
- Only test systems you own or have permission to test

## Development Setup

### 1. Fork and Clone

```bash
git clone https://github.com/YOUR_USERNAME/phantom-proxy.git
cd phantom-proxy
```

### 2. Install Dependencies

```bash
# Go
go mod download

# Python
pip install -r requirements.txt

# Node.js
cd frontend && npm install
```

### 3. Run Tests

```bash
# All tests
make test

# Go tests
make test-go

# Python tests
make test-python
```

### 4. Code Style

**Go:**
```bash
go fmt ./...
golangci-lint run
```

**Python:**
```bash
black .
flake8
```

**TypeScript:**
```bash
npm run lint
```

## Pull Request Process

1. Create feature branch
2. Make changes
3. Run tests
4. Update documentation
5. Submit PR

## Security

- Report vulnerabilities privately
- Do not expose sensitive information
- Follow responsible disclosure

## License

By contributing, you agree to license your work under MIT License.
