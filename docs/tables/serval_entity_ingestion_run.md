# Table: serval_entity_ingestion_run

Runs of each entity ingestion config, newest first, with status and item counts. Filter ingestion_config_id and started_at (>, >=) to bound the scan.

## Examples

**Failed runs in the last 7 days:**

```sql
select c.workflow_name, r.started_at, r.failure_reason, r.failure_message
from serval_entity_ingestion_config c
join serval_entity_ingestion_run r on r.ingestion_config_id = c.id
where r.started_at > now() - interval '7 days'
  and r.status = 'ENTITY_INGESTION_RUN_STATUS_FAILED';
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Run ID. |
| `ingestion_config_id` | text | Ingestion config. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `status` | text | ENTITY_INGESTION_RUN_STATUS_COMPLETED, ..._FAILED, ..._RUNNING, ... |
| `is_full_sync` | boolean | Whether the run was a full sync. |
| `started_at` | timestamp with time zone | Start time. |
| `completed_at` | timestamp with time zone | Completion time. |
| `items_added_count` | bigint | Records added. |
| `items_updated_count` | bigint | Records updated. |
| `items_deleted_count` | bigint | Records deleted. |
| `items_unchanged_count` | bigint | Records unchanged. |
| `items_errored_count` | bigint | Records that failed. |
| `items_publish_failed_count` | bigint | Publish targets reported as failed. |
| `publish_outcome` | text | published, partial, failed, not_ready or unknown. |
| `failure_reason` | text | Pipeline phase that failed. |
| `failure_code` | text | Machine-readable failure cause. |
| `failure_message` | text | Failure message. |
