# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 14.0.x  | :white_check_mark: |
| 13.0.x  | :warning: Legacy   |
| < 13.0  | :x: Unsupported    |

## Reporting a Vulnerability

**DO NOT** create public issues for security vulnerabilities.

### Report Privately:

1. **Email:** security@phantomseclabs.com
2. **GitHub:** Use private vulnerability reporting
3. **Include:**
   - Description
   - Steps to reproduce
   - Impact assessment
   - Suggested fix (if any)

### Response Time:

- **Critical:** 24 hours
- **High:** 48 hours
- **Medium:** 1 week
- **Low:** 2 weeks

## Security Best Practices

### For Users:

1. **Authorization:**
   - Only test systems you own
   - Get written permission
   - Follow Rules of Engagement

2. **Configuration:**
   - Change default passwords
   - Use HTTPS/TLS
   - Enable authentication
   - Restrict network access

3. **Monitoring:**
   - Enable audit logging
   - Monitor for abuse
   - Regular security updates

### For Developers:

1. **Code:**
   - Input validation
   - Output encoding
   - Secure defaults
   - Error handling

2. **Dependencies:**
   - Regular updates
   - Vulnerability scanning
   - Pin versions

3. **Testing:**
   - Security tests
   - Penetration testing
   - Code review

## Security Features

- ✅ TLS 1.3 encryption
- ✅ mTLS support
- ✅ GOST encryption (FSTEC)
- ✅ Session isolation
- ✅ Audit logging
- ✅ Rate limiting
- ✅ Input validation

## Known Limitations

- Not for production use without additional hardening
- Requires proper network segmentation
- Needs external authentication for multi-user

## Compliance

- FSTEC УЗ-1/УЗ-2 ready
- GDPR considerations
- Ethical use only

---

**Last Updated:** February 20, 2026
