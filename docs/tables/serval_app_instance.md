# Table: serval_app_instance

App instances (connected applications) in each workspace, with their owners.

## Examples

**Apps with no owner:**

```sql
select name, coalesce(service, custom_service_name) as service, team_id
from serval_app_instance
where jsonb_array_length(owner_user_ids) = 0
  and jsonb_array_length(owner_group_ids) = 0;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | App instance ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `name` | text | Name. |
| `service` | text | Built-in service slug, e.g. okta or github. |
| `custom_service_id` | text | Custom service ID, for apps without a built-in integration. |
| `custom_service_name` | text | Custom service name. |
| `external_service_instance_id` | text | Identifier of the instance in the external service, e.g. its domain. |
| `access_requests_enabled` | boolean | Whether users can request access to the app. |
| `default_access_policy_id` | text | Default access policy for the app's roles. |
| `owners` | jsonb | Owners: userIds and groupIds. |
| `owner_user_ids` | jsonb | IDs of the users who own the app. |
| `owner_group_ids` | jsonb | IDs of the groups that own the app. |
| `integration_ids` | jsonb | Integrations connected to the app. |
| `tag_ids` | jsonb | Tag IDs. |
