# Table: serval_ticket

Tickets in each workspace. Without a created_at filter the API returns the last year; filter created_at (>, >=, <, <=) to change the window.

## Examples

**Overdue open tickets created in the last 90 days:**

```sql
select friendly_identifier, name, status_name, priority_name, due_date
from serval_ticket
where created_at > now() - interval '90 days'
  and is_overdue
  and completed_at is null;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | Ticket ID. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `friendly_identifier` | text | Human-readable identifier. |
| `type` | text | Ticket type, e.g. request, incident or change. |
| `name` | text | Title. |
| `description` | text | Description. |
| `status_id` | text | Status option ID. |
| `status_name` | text | Status, e.g. In Progress. |
| `priority_id` | text | Priority option ID. |
| `priority_name` | text | Priority, e.g. High. |
| `escalation_level` | text | Escalation level. |
| `is_overdue` | boolean | Whether the due date has passed. |
| `due_date` | text | Due date (YYYY-MM-DD). |
| `due_at` | timestamp with time zone | Planned end. |
| `created_by_user_id` | text | Creator. |
| `requester_user_id` | text | Requester. |
| `assigned_to_user_id` | text | Assignee. |
| `assigned_to_group_id` | text | Assignment group. |
| `category_option_id` | text | Category option. |
| `parent_ticket_id` | text | Parent ticket, for tasks. |
| `label_ids` | jsonb | Label IDs. |
| `url` | text | Ticket URL. |
| `created_at` | timestamp with time zone | Creation time. |
| `escalated_at` | timestamp with time zone | Escalation time. |
| `completed_at` | timestamp with time zone | Completion time. |
