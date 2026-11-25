# Table: serval_approval_request

Approval requests raised for access requests and workflow runs. The steps column fetches each request's steps and approver decisions (one API call per row).

## Examples

**Approvers of each approved request:**

```sql
select
  a.id,
  a.result,
  res ->> 'approverUserId' as approver_user_id,
  (res ->> 'autoApproved')::boolean as auto_approved
from serval_approval_request a,
  jsonb_array_elements(a.steps) st,
  jsonb_array_elements(st -> 'resolutions') res
where a.result = 'APPROVED';
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Approval request ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `result` | text | Overall result: PENDING, APPROVED, DENIED, ... |
| `access_request_time_allocation_id` | text | Access request time allocation this approval is for. |
| `workflow_run_id` | text | Workflow run this approval is for. |
| `approver_justification_mode` | text | Whether approvers must justify their decision. |
| `steps` | jsonb | Steps, each with its result and approver resolutions (approverUserId, result, autoApproved). |
| `created_at` | timestamp with time zone | Creation time. |
| `updated_at` | timestamp with time zone | Last update time. |
