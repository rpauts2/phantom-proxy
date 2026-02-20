---
name: 📋 Release Checklist
about: Track release preparation and deployment
title: '[RELEASE] v'
labels: ['release']
assignees: ''
---

# Release Checklist

## Release Information

**Version:** `v`
**Release Date:** 
**Release Manager:** 
**Branch:** 

---

## Pre-Release Tasks

### Code Quality
- [ ] All tests passing
- [ ] Code coverage meets threshold (>80%)
- [ ] No critical security vulnerabilities
- [ ] Linting passes (black, flake8, mypy)
- [ ] Go tests passing (`go test`)
- [ ] No TODO/FIXME in critical paths

### Documentation
- [ ] CHANGELOG.md updated
- [ ] README.md reflects new features
- [ ] API documentation updated
- [ ] Migration guide created (if breaking changes)
- [ ] Release notes drafted

### Version Updates
- [ ] Version number updated in code
- [ ] Version number updated in config files
- [ ] go.mod version updated (if applicable)
- [ ] Python setup.py/_version.py updated

---

## Build Verification

### Platforms
- [ ] Linux build successful
- [ ] Windows build successful
- [ ] macOS build successful
- [ ] Docker image built and tested

### Artifacts
- [ ] Python executable generated
- [ ] Go binary generated
- [ ] All artifacts uploaded to release
- [ ] Checksums generated (SHA256)

---

## Testing

### Automated Tests
- [ ] Unit tests: PASSED
- [ ] Integration tests: PASSED
- [ ] E2E tests: PASSED
- [ ] Performance tests: PASSED

### Manual Testing
- [ ] Core functionality verified
- [ ] API endpoints tested
- [ ] UI tested (if applicable)
- [ ] Backward compatibility verified

---

## Security

- [ ] Security scan completed
- [ ] Dependencies updated
- [ ] No known CVEs in dependencies
- [ ] Secrets scanning passed
- [ ] SAST scan completed

---

## Deployment

### Staging
- [ ] Deployed to staging
- [ ] Smoke tests passed
- [ ] Performance baseline met

### Production
- [ ] Deployment plan reviewed
- [ ] Rollback plan prepared
- [ ] Monitoring configured
- [ ] Alert rules updated
- [ ] Team notified

---

## Post-Release

- [ ] GitHub release published
- [ ] Release announcement sent
- [ ] Documentation published
- [ ] Social media posts (if applicable)
- [ ] Internal wiki updated

---

## Release Notes

### New Features

### Bug Fixes

### Breaking Changes

### Known Issues

---

## Approvers

- [ ] Code Review: 
- [ ] QA Approval: 
- [ ] Security Approval: 
- [ ] Product Approval: 

---

**Release Status:** 
- [ ] Ready to Release
- [ ] Blocked (see comments)

---

## Comments

