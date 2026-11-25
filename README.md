# Serval Plugin for Steampipe

Use SQL to query users, apps, roles, access policies, access requests and approvals, role memberships, unused-access reclamation, asset database imports, tickets, workflows and the audit log from [Serval](https://serval.com), the IT service management and access management platform.

- **[Get started →](docs/index.md)**
- Community: [Join #steampipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/vthiery/steampipe-plugin-serval/issues)

## Quick start

### Install

Install the plugin with [Steampipe](https://steampipe.io):

```sh
steampipe plugin install ghcr.io/vthiery/serval
```

### Configure

Copy the sample config and set your API key:

```sh
cp config/serval.spc ~/.steampipe/config/serval.spc
```

Edit `~/.steampipe/config/serval.spc`:

```hcl
connection "serval" {
  plugin = "ghcr.io/vthiery/serval"

  client_id     = "YOUR_CLIENT_ID"
  client_secret = "YOUR_CLIENT_SECRET"
  base_url      = "https://public.eu1.serval.com" # omit for US tenants
  team_ids      = ["WORKSPACE_ID"]
}
```

Alternatively, leave the attributes unset and export `SERVAL_CLIENT_ID`, `SERVAL_CLIENT_SECRET`, `SERVAL_BASE_URL` and `SERVAL_TEAM_IDS` (comma-separated). The scopes the key needs depend on the tables you query: `users:read` for `serval_user` and `serval_user_relationship`; `apps:read` for `serval_app_instance`, `serval_app_resource`, `serval_app_resource_role` and `serval_reclamation_candidate`; `settings:read` for `serval_access_policy`, `serval_access_request` and the access-request rows of `serval_approval_request`; `workflows:read` for `serval_workflow` and the workflow-run rows of `serval_approval_request`; `access-memberships:read` for `serval_access_membership` (plus `users:read` for a full scan); `entities:read` for the two ingestion tables; `tickets:read` for `serval_ticket`; and `org:audit` for `serval_audit_log`.

### Run a query

```shell
steampipe query
```

List access policies that grant standing access without approval:

```sql
select
  name,
  team_id
from
  serval_access_policy
where
  approval_procedure_id is null
  and max_access_minutes is null;
```

## Tables

| Table | Description |
|-------|-------------|
| [serval_access_membership](docs/tables/serval_access_membership.md) | Role memberships recorded by Serval access management (provisioned, synced or linked-group). |
| [serval_access_policy](docs/tables/serval_access_policy.md) | Access policies: approval procedure, maximum duration and justification requirements for access requests. |
| [serval_access_request](docs/tables/serval_access_request.md) | Access requests for roles, with their status and time allocations (initial grant and extensions). |
| [serval_app_instance](docs/tables/serval_app_instance.md) | App instances (connected applications) in each workspace, with their owners. |
| [serval_app_resource](docs/tables/serval_app_resource.md) | Resources of each app instance (the app itself, groups, projects, ...) that roles grant access to. |
| [serval_app_resource_role](docs/tables/serval_app_resource_role.md) | Roles (entitlements) on app resources, with their access policy and provisioning method. |
| [serval_approval_request](docs/tables/serval_approval_request.md) | Approval requests raised for access requests and workflow runs. |
| [serval_audit_log](docs/tables/serval_audit_log.md) | The organisation's Serval audit log (last 90 days, newest first). |
| [serval_entity_ingestion_config](docs/tables/serval_entity_ingestion_config.md) | Entity ingestion configs: scheduled imports (devices, SaaS apps, ...) into Serval databases. |
| [serval_entity_ingestion_run](docs/tables/serval_entity_ingestion_run.md) | Runs of each entity ingestion config, newest first, with status and item counts. |
| [serval_reclamation_candidate](docs/tables/serval_reclamation_candidate.md) | Reclamation candidates: access grants flagged for review or reclaim because the user stopped signing in to the app. |
| [serval_ticket](docs/tables/serval_ticket.md) | Tickets in each workspace. |
| [serval_user](docs/tables/serval_user.md) | Users of the Serval organisation, including deactivated users. |
| [serval_user_relationship](docs/tables/serval_user_relationship.md) | Relationships between users, such as a user's manager. |
| [serval_workflow](docs/tables/serval_workflow.md) | Workflows (automation code) in each workspace. |

## Workspaces

Most Serval objects live in a workspace (team). Workspace-scoped tables have a `team_id` column and scan every workspace in `team_ids`, or only the one a `team_id = '...'` filter names. Listing workspaces needs a permission read-only keys usually lack, so set `team_ids` explicitly; the workspace IDs are in the Serval web app URL and in the `team_id` column of `serval_audit_log`.

## API limits

Serval does not document a rate limit. The plugin sends at most 10 requests per second and retries HTTP 429 and 5xx responses with exponential backoff.

`serval_access_membership` is the expensive table: the API returns one user's memberships per call, so a scan without a `user_id` filter makes one call per user per workspace (about four minutes for 1,200 users in two workspaces). Join it from `serval_user` (`join serval_access_membership m on m.user_id = u.id`, or `join lateral`) to query only the users you need. A `user_id = (select ...)` subquery does not narrow the scan: Steampipe does not pass the subquery's value to the plugin.

## Development

### Prerequisites

- [Steampipe](https://steampipe.io/downloads)
- [Golang](https://golang.org/doc/install)

### Build and Install

```sh
make install
```

Configure the plugin as a local build:

```sh
cp config/serval.spc ~/.steampipe/config/serval.spc
```

and set `plugin = "local/serval"` in the connection.

## Testing

Run a smoke query against every table:

```sh
make test
```

The test script ([scripts/test_tables.sh](scripts/test_tables.sh)) builds the plugin, queries each table, and reports pass/fail/skip.

### Further reading

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)
- [Serval API reference](https://docs.serval.com/sections/api-reference/find-endpoints)
