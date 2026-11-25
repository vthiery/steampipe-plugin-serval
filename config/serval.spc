connection "serval" {
  plugin = "ghcr.io/vthiery/serval"

  # OAuth client ID and secret of a Serval API key. Read-only scopes are
  # sufficient. Fall back to the SERVAL_CLIENT_ID and SERVAL_CLIENT_SECRET
  # environment variables when not set here.
  # client_id     = "YOUR_CLIENT_ID"
  # client_secret = "YOUR_CLIENT_SECRET"

  # Base URL of your tenant's region. Defaults to https://public.api.serval.com
  # (US); EU tenants use https://public.eu1.serval.com. Falls back to
  # SERVAL_BASE_URL.
  # base_url = "https://public.api.serval.com"

  # Workspace (team) IDs queried by the workspace-scoped tables. When empty,
  # the plugin lists every workspace, which needs a key allowed to list teams.
  # Falls back to SERVAL_TEAM_IDS (comma-separated).
  # team_ids = ["WORKSPACE_ID_1", "WORKSPACE_ID_2"]
}
