---
id: requirement_customer_user_account
status: STABLE
version: 1.0
parents: []
human_name: User Account Identity Management
priority: 3
tags: [auth, identity, account]
dependents:
  - [[battleui:ui_modal_box]]
  - [[upsilonapi:api_auth_user]]
  - [[upsilonapi:rule_gdpr_compliance]]
  - [[upsilonapi:rule_password_policy]]
type: REQUIREMENT
layer: BUSINESS
---

# New Atom

## INTENT
To define the requirements for managing user authentication and personal identity data (account name, email, credentials).

## THE RULE / LOGIC
- Users must be able to update their nickname (`account_name`), email, birth date, and residential address.
- Account names must remain unique across the system.
- Password changes require secure validation and conform to the password policy.
- All sensitive updates require a valid session.
- **Account Deletion:** Users must have the right to request full account deletion, which must be handled via soft-deletion and anonymization as per [[rule_gdpr_compliance]].

## TECHNICAL INTERFACE
- **Controller:** `AuthController`
- **Code Tag:** `@spec-link [[customer_user_account]]`
- **Endpoints:** `POST /api/v1/auth/update`, `POST /api/v1/auth/password`

## EXPECTATION
