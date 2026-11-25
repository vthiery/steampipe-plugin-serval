package serval

import "github.com/turbot/steampipe-plugin-sdk/v5/plugin"

// ServalConfig holds the connection configuration for the Serval plugin.
type ServalConfig struct {
	// OAuth client ID of a Serval API key. Falls back to SERVAL_CLIENT_ID.
	ClientID *string `hcl:"client_id"`
	// OAuth client secret of a Serval API key. Falls back to SERVAL_CLIENT_SECRET.
	ClientSecret *string `hcl:"client_secret"`
	// Base URL of the regional API. Defaults to https://public.api.serval.com
	// (US). EU tenants use https://public.eu1.serval.com. Falls back to SERVAL_BASE_URL.
	BaseURL *string `hcl:"base_url"`
	// Workspace (team) IDs to query for workspace-scoped tables. When empty, the
	// plugin lists workspaces, which needs a key allowed to list teams.
	// Falls back to SERVAL_TEAM_IDS (comma-separated).
	TeamIDs []string `hcl:"team_ids,optional"`
}

// ConfigInstance returns a new, empty ServalConfig.
func ConfigInstance() interface{} {
	return &ServalConfig{}
}

// GetConfig retrieves the connection configuration for the given connection.
func GetConfig(connection *plugin.Connection) ServalConfig {
	if connection == nil || connection.Config == nil {
		return ServalConfig{}
	}
	cfg, _ := connection.Config.(ServalConfig)
	return cfg
}
