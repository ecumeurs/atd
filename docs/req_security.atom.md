---
id: req_security
human_name: Sanctum Token Security Requirement
type: MODULE
layer: ARCHITECTURE
version: 1.1
status: STABLE
priority: 5
tags: [auth, sanctum]
parents: []
dependents:
  - [[req_security_authorization]]
  - [[req_security_public_access]]
  - [[req_security_token_exchange]]
  - [[req_security_token_ttl]]
  - [[uc_auth_logout]]
  - [[upsilonapi:rule_admin_access_restriction]]
  - [[upsilonapi:rule_gdpr_compliance]]
  - [[upsilonapi:rule_password_policy]]
---
# Sanctum Token Security Requirement

## INTENT
To aggregate the constituent rules of Sanctum Token Security.

## THE RULE / LOGIC
Ensures that all non-public requests to the application are authenticated and secure.
- **Encryption:** All traffic MUST use HTTPS (support for self-signed certificates for development/light environments).
- **Authentication:** Laravel Sanctum Personal Access Tokens sent in the `Authorization` header as a Bearer token.
- **Password Policy:** Enforces complexity as defined in [[rule_password_policy]].
- **Privacy Core:** Enforces GDPR compliance as defined in [[rule_gdpr_compliance]].

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[req_security]]`
