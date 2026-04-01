#!/bin/bash
set -e

cd /app

# Initialize project
spektacular init claude

# Start spec workflow
spektacular spec new --data '{"name":"user-auth"}'

# Overview
cat > /tmp/overview.md << 'CONTENT'
Stateless user authentication system using JWT access and refresh tokens.
Replaces the current session-based auth to enable horizontal scaling across
multiple backend services. Benefits backend developers who consume the auth
API and end users who need reliable login across services.
CONTENT

spektacular spec goto --data '{"step":"requirements"}' --stdin spec_content < /tmp/overview.md

# Requirements
cat > /tmp/requirements.md << 'CONTENT'
- The system MUST issue a signed JWT access token on successful login
- The system MUST issue a refresh token alongside the access token
- Access tokens MUST expire after 15 minutes
- Refresh tokens MUST expire after 7 days
- The system MUST validate token signatures on every authenticated request
- The system MUST reject expired tokens with a 401 response
- The system MUST support token revocation via a deny list
- The system MUST hash refresh tokens before storing them
CONTENT

spektacular spec goto --data '{"step":"acceptance_criteria"}' --stdin spec_content < /tmp/requirements.md

# Acceptance criteria
cat > /tmp/acceptance.md << 'CONTENT'
- [ ] A user can log in with valid credentials and receive a JWT access token and refresh token
- [ ] A request with an expired access token returns HTTP 401
- [ ] A valid refresh token can be exchanged for a new access token
- [ ] A revoked token is rejected on the next request
- [ ] Tokens issued by one service instance are accepted by another instance
CONTENT

spektacular spec goto --data '{"step":"constraints"}' --stdin spec_content < /tmp/acceptance.md

# Constraints
cat > /tmp/constraints.md << 'CONTENT'
- Must use RS256 signing algorithm for JWT tokens
- Must not store access tokens server-side
- Refresh token storage must use the existing PostgreSQL database
- Must not break existing API endpoints during migration
- Token payload must not contain sensitive PII beyond user ID and role
CONTENT

spektacular spec goto --data '{"step":"technical_approach"}' --stdin spec_content < /tmp/constraints.md

# Technical approach
cat > /tmp/approach.md << 'CONTENT'
- Add a new `auth` package implementing token issuance and validation
- Use asymmetric RS256 keys stored in environment config
- Implement middleware that validates the Authorization header on protected routes
- Store refresh tokens in a `refresh_tokens` PostgreSQL table with hashed values
- Add a `/auth/login`, `/auth/refresh`, and `/auth/revoke` endpoint
- Use Redis for the token deny list with TTL matching token expiry
CONTENT

spektacular spec goto --data '{"step":"success_metrics"}' --stdin spec_content < /tmp/approach.md

# Success metrics
cat > /tmp/metrics.md << 'CONTENT'
- Token validation adds less than 5ms p99 latency to authenticated requests
- Auth service handles 1000 login requests per second
- Zero downtime during migration from session-based auth
- 100% of existing integration tests pass after migration
CONTENT

spektacular spec goto --data '{"step":"non_goals"}' --stdin spec_content < /tmp/metrics.md

# Non-goals
cat > /tmp/nongoals.md << 'CONTENT'
- OAuth2 / OpenID Connect provider support (future work)
- Multi-factor authentication (separate initiative)
- Social login (Google, GitHub, etc.)
- Fine-grained permission scoping beyond role-based access
CONTENT

spektacular spec goto --data '{"step":"verification"}' --stdin spec_content < /tmp/nongoals.md

# Verification
cat > /tmp/verification.md << 'CONTENT'
All sections reviewed and complete. Requirements are testable. Acceptance
criteria are binary pass/fail. Constraints are concrete. Technical approach
aligns with existing architecture. Success metrics are measurable. Non-goals
clearly scope the work.
CONTENT

spektacular spec goto --data '{"step":"finished"}' --stdin spec_content < /tmp/verification.md
