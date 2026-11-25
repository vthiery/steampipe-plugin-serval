# Table: serval_access_membership

Role memberships recorded by Serval access management (provisioned, synced or linked-group). The API answers one user at a time: filter user_id to query one user, otherwise every user (including deactivated ones) is queried in every workspace, one call each.

## Examples

**Roles held by one user:**

```sql
select m.app_instance_name, m.resource_name, m.role_name
from serval_user u
join serval_access_membership m on m.user_id = u.id
where u.email = 'jane@example.com';
```

**Access still held by deactivated users:**

```sql
select u.email, m.app_instance_name, m.resource_name, m.role_name
from serval_user u
join lateral (
  select * from serval_access_membership where user_id = u.id
) m on true
where u.deactivated;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `user_id` | text | User holding the role. |
| `team_id` | text | ID of the workspace (team) the object belongs to. |
| `app_instance_id` | text | App instance. |
| `app_instance_name` | text | App instance name. |
| `resource_id` | text | Resource. |
| `resource_name` | text | Resource name. |
| `role_id` | text | Role. |
| `role_name` | text | Role name. |
