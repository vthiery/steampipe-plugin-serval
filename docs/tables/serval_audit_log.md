# Table: serval_audit_log

The organisation's Serval audit log (last 90 days, newest first). Filter timestamp (>, >=, <, <=) to bound the scan. Needs the org:audit scope.

## Examples

**API keys created or revoked in the last 30 days:**

```sql
select timestamp, event_type, actor_email, event ->> 'name' as key_name
from serval_audit_log
where timestamp > now() - interval '30 days'
  and event_type in ('apiKeyCreated', 'apiKeyRevoked');
```

**Sign-ins by Serval staff impersonating users:**

```sql
select timestamp, actor_display_name, event
from serval_audit_log
where timestamp > now() - interval '90 days'
  and event_type in ('impersonationSessionStarted', 'impersonationSessionTerminated');
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Event ID. |
| `timestamp` | timestamp with time zone | When the event happened. |
| `event_type` | text | Event type, e.g. userLogin, apiKeyCreated, accessProvisioned. |
| `event` | jsonb | Event details (the payload of event_type). |
| `team_id` | text | Workspace the event belongs to, when it is workspace-specific. |
| `severity` | text | SEVERITY_INFO, SEVERITY_WARN, ... |
| `outcome_status` | text | STATUS_SUCCESS, STATUS_DENIED, STATUS_FAILURE, ... |
| `outcome_reason_code` | text | Machine-readable outcome reason. |
| `outcome_reason_detail` | text | Outcome detail. |
| `actor_id` | text | Acting user, API key or system ID. |
| `actor_type` | text | user, apiKey, worker or system. |
| `actor_email` | text | Acting user's email. |
| `actor_display_name` | text | Acting user's, API key's or system's display name. |
| `actor` | jsonb | Actor object. |
| `ip_address` | text | Source IP address. |
| `user_agent` | text | Source user agent. |
| `organization_id` | text | Organisation ID. |
