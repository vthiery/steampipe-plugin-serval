# Table: serval_app_resource

Resources of each app instance (the app itself, groups, projects, ...) that roles grant access to.

## Examples

**Resources per app:**

```sql
select a.name as app, count(r.id) as resources
from serval_app_instance a
join serval_app_resource r on r.app_instance_id = a.id
group by a.name
order by resources desc;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Resource ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `app_instance_id` | text | App instance the resource belongs to. |
| `resource_type` | text | Resource type, e.g. access or group. |
| `name` | text | Name. |
| `description` | text | Description. |
| `external_id` | text | ID in the external system. |
| `scoped_id` | text | Stable scoped ID (ar_...). |
| `tag_ids` | jsonb | Tag IDs. |
