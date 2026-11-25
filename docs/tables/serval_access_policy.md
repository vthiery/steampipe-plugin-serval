# Table: serval_access_policy

Access policies: approval procedure, maximum duration and justification requirements for access requests.

## Examples

**Policies that grant standing access without approval:**

```sql
select name, team_id
from serval_access_policy
where approval_procedure_id is null
  and max_access_minutes is null;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Policy ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `name` | text | Name. |
| `description` | text | Description. |
| `approval_procedure_id` | text | Approval procedure that must approve requests. Null means requests are granted without approval. |
| `max_access_minutes` | bigint | Maximum access duration in minutes. Null means access does not expire. |
| `recommended_access_minutes` | bigint | Recommended access duration in minutes. |
| `require_business_justification` | boolean | Whether a business justification is required. |
| `require_on_behalf_of_confirmation` | boolean | Whether requests made for someone else need their confirmation. |
| `require_form_confirmation` | boolean | Whether requests are collected with a form. |
| `scoped_id` | text | Stable scoped ID (apo_...). |
