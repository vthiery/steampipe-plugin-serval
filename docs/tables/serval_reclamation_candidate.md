# Table: serval_reclamation_candidate

Reclamation candidates: access grants flagged for review or reclaim because the user stopped signing in to the app.

## Examples

**Open reviews of unused access older than 14 days:**

```sql
select user_id, app_instance_id, entitlement_id, last_login_at, created_at
from serval_reclamation_candidate
where status = 'RECLAMATION_CANDIDATE_STATUS_IN_REVIEW'
  and created_at < now() - interval '14 days';
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Reclamation ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `status` | text | Lifecycle state: in review, reclaiming, reclaimed, dismissed, reclaim failed, ... |
| `user_id` | text | User whose access is flagged. |
| `entitlement_id` | text | Role (entitlement) the grant is on. |
| `app_instance_id` | text | App instance. |
| `resource_type` | text | Entitlement's resource type. |
| `last_login_at` | timestamp with time zone | User's last sign-in to the app, the inactivity signal. |
| `decided_by_user_id` | text | Reviewer who concluded the review. Null for automatic reclaims and open reviews. |
| `created_at` | timestamp with time zone | When the grant was flagged. |
| `concluded_at` | timestamp with time zone | When the reclamation concluded. |
