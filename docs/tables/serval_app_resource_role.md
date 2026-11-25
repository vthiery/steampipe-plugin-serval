# Table: serval_app_resource_role

Roles (entitlements) on app resources, with their access policy and provisioning method.

## Examples

**Requestable roles and their access policy:**

```sql
select r.name as role, p.name as policy, p.max_access_minutes
from serval_app_resource_role r
left join serval_access_policy p on p.id = r.access_policy_id
where r.requests_enabled;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Role ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `resource_id` | text | Resource the role belongs to. |
| `name` | text | Name. |
| `description` | text | Description. |
| `requests_enabled` | boolean | Whether users can request the role. |
| `access_policy_id` | text | Access policy governing requests for the role. |
| `provisioning_method` | jsonb | How access is provisioned (builtin, custom workflow, linked roles, manual). |
| `is_ingested` | boolean | Whether the role was imported from the external system. |
| `external_id` | text | ID in the external system. |
| `require_form_confirmation` | boolean | Whether requests are collected with a form. |
| `scoped_id` | text | Stable scoped ID (role_...). |
| `tag_ids` | jsonb | Tag IDs. |
