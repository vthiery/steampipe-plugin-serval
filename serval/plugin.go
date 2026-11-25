// Package serval provides a Steampipe plugin for the Serval public API.
package serval

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Plugin returns the definition of the Serval Steampipe plugin.
func Plugin(ctx context.Context) *plugin.Plugin {
	return &plugin.Plugin{
		Name:             "steampipe-plugin-serval",
		DefaultTransform: transform.FromGo().NullIfZero(),
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		TableMap: map[string]*plugin.Table{
			"serval_access_membership":       tableServalAccessMembership(),
			"serval_access_policy":           tableServalAccessPolicy(),
			"serval_access_request":          tableServalAccessRequest(),
			"serval_app_instance":            tableServalAppInstance(),
			"serval_app_resource":            tableServalAppResource(),
			"serval_app_resource_role":       tableServalAppResourceRole(),
			"serval_approval_request":        tableServalApprovalRequest(),
			"serval_audit_log":               tableServalAuditLog(),
			"serval_entity_ingestion_config": tableServalEntityIngestionConfig(),
			"serval_entity_ingestion_run":    tableServalEntityIngestionRun(),
			"serval_reclamation_candidate":   tableServalReclamationCandidate(),
			"serval_ticket":                  tableServalTicket(),
			"serval_user":                    tableServalUser(),
			"serval_user_relationship":       tableServalUserRelationship(),
			"serval_workflow":                tableServalWorkflow(),
		},
	}
}
