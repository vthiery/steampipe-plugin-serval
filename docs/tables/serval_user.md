# Table: serval_user

Users of the Serval organisation, including deactivated users.

## Examples

**Organisation admins:**

```sql
select email, name, created_at
from serval_user
where role = 'USER_ROLE_ORG_ADMIN'
  and not deactivated;
```

**Users deactivated in the last 30 days:**

```sql
select email, deactivated_at
from serval_user
where deactivated_at > now() - interval '30 days';
```

## Columns

| Column | Type | Description |
|--------|------|-------------|
| `id` | text | User ID. |
| `email` | text | Email address. |
| `name` | text | Full name. |
| `first_name` | text | First name. |
| `last_name` | text | Last name. |
| `role` | text | Organisation role: USER_ROLE_ORG_ADMIN, USER_ROLE_ORG_MEMBER or USER_ROLE_ORG_GUEST. |
| `department` | text | Department. |
| `job_title` | text | Job title. |
| `auth_method` | text | How the user signs in. |
| `timezone` | text | IANA time zone. |
| `locale` | text | BCP 47 language tag. |
| `deactivated` | boolean | True when the user is deactivated. |
| `deactivated_at` | timestamp with time zone | When the user was deactivated. |
| `created_at` | timestamp with time zone | Creation time. |
