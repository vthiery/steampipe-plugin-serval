# Table: serval_access_request

Access requests for roles, with their status and time allocations (initial grant and extensions).

## Examples

**Access requests in the last 30 days:**

```sql
select id, status, target_user_id, requested_role_id, created_at
from serval_access_request
where created_at > now() - interval '30 days';
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Request ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `status` | text | Overall status, e.g. ACCESS_REQUEST_STATUS_APPROVED. |
| `target_user_id` | text | User the access is for. |
| `requested_role_id` | text | Requested role. |
| `linked_ticket_id` | text | Ticket the request came from. |
| `time_allocations` | jsonb | Time allocations: each grant or extension, with requester, justification, status and minutes. |
| `expires_at` | timestamp with time zone | When the active time allocation expires. |
| `created_at` | timestamp with time zone | Creation time. |
