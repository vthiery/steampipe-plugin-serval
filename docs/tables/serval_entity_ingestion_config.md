# Table: serval_entity_ingestion_config

Entity ingestion configs: scheduled imports (devices, SaaS apps, ...) into Serval databases.

## Examples

**Disabled imports:**

```sql
select workflow_name, service_name, team_id
from serval_entity_ingestion_config
where not enabled;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Ingestion config ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `app_instance_id` | text | App instance the data comes from. |
| `service_name` | text | Source service name. |
| `workflow_id` | text | Ingestion workflow. |
| `workflow_name` | text | Ingestion workflow name. |
| `is_workflow_published` | boolean | Whether the workflow is published. |
| `enabled` | boolean | Whether scheduled syncs run. |
| `sync_interval_seconds` | bigint | Minimum interval between scheduled syncs. Null means the platform default. |
| `sync_deletions` | boolean | Whether records removed at the source are removed. |
| `internal_entity_type_id` | text | Entity type (database) the records land in. |
| `allow_create_new_entities` | boolean | Whether a sync may create new records. |
| `has_ingestion_context` | boolean | Whether the workflow uses the ingestion context. |
| `created_at` | timestamp with time zone | Creation time. |
| `updated_at` | timestamp with time zone | Last update time. |
