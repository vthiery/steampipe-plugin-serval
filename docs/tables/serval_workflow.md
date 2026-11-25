# Table: serval_workflow

Workflows (automation code) in each workspace.

## Examples

**Workflows with unpublished changes:**

```sql
select name, team_id
from serval_workflow
where has_unpublished_changes;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Workflow ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `name` | text | Name. |
| `description` | text | Description. |
| `is_published` | boolean | Whether the workflow is published. |
| `has_unpublished_changes` | boolean | Whether there are unpublished changes. |
| `approval_procedure_id` | text | Approval procedure every run needs. Null means runs need no approval. |
| `show_in_service_catalog` | boolean | Whether users can run it from the service catalog. |
| `require_form_confirmation` | boolean | Whether runs are collected with a form. |
| `package` | text | Access package whose connections the workflow uses. |
| `folder_id` | text | Folder. |
| `scoped_id` | text | Stable scoped ID (wf_...). |
| `tag_ids` | jsonb | Tag IDs. |
| `content` | text | Workflow source code. |
