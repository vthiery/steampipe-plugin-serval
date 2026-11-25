# Table: serval_user_relationship

Relationships between users, such as a user's manager.

## Examples

**Active users with no manager:**

```sql
select u.email
from serval_user u
left join serval_user_relationship r
  on r.user_id = u.id and r.relationship_type = 'USER_RELATIONSHIP_TYPE_MANAGER'
where not u.deactivated
  and u.role <> 'USER_ROLE_ORG_GUEST'
  and r.user_id is null;
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `user_id` | text | The user. |
| `related_user_id` | text | The related user, for example the manager. |
| `relationship_type` | text | Relationship type, e.g. USER_RELATIONSHIP_TYPE_MANAGER. |
| `created_at` | timestamp with time zone | Creation time. |
