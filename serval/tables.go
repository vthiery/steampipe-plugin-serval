package serval

import (
	"context"
	"net/url"
	"strconv"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// serval_user — people in the Serval organisation, including deactivated ones.

func tableServalUser() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_user",
		Description: "Users of the Serval organisation, including deactivated users.",
		List: &plugin.ListConfig{
			Hydrate:    listHydrate("/v2/users", 500, []qualParam{{"email", "emails"}}, url.Values{"includeDeactivated": {"true"}}),
			KeyColumns: optionalQuals([]qualParam{{"email", "emails"}}),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/users")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "User ID."),
			col("email", tStr, "email", "Email address."),
			col("name", tStr, "name", "Full name."),
			col("first_name", tStr, "firstName", "First name."),
			col("last_name", tStr, "lastName", "Last name."),
			col("role", tStr, "role", "Organisation role: USER_ROLE_ORG_ADMIN, USER_ROLE_ORG_MEMBER or USER_ROLE_ORG_GUEST."),
			col("department", tStr, "department", "Department."),
			col("job_title", tStr, "jobTitle", "Job title."),
			col("auth_method", tStr, "authMethod", "How the user signs in."),
			col("timezone", tStr, "timezone", "IANA time zone."),
			col("locale", tStr, "locale", "BCP 47 language tag."),
			{Name: "deactivated", Type: tBool, Transform: transform.FromField("deactivatedAt").Transform(isSet), Description: "True when the user is deactivated."},
			col("deactivated_at", tTS, "deactivatedAt", "When the user was deactivated."),
			col("created_at", tTS, "createdAt", "Creation time."),
		},
	}
}

//// serval_user_relationship — manager and other user-to-user relationships.

var relationshipQuals = []qualParam{
	{"user_id", "userId"}, {"related_user_id", "relatedUserId"}, {"relationship_type", "relationshipType"},
}

func tableServalUserRelationship() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_user_relationship",
		Description: "Relationships between users, such as a user's manager.",
		List: &plugin.ListConfig{
			Hydrate:    listHydrate("/v2/user-relationships", 500, relationshipQuals, nil),
			KeyColumns: optionalQuals(relationshipQuals),
		},
		Columns: []*plugin.Column{
			col("user_id", tStr, "userId", "The user."),
			col("related_user_id", tStr, "relatedUserId", "The related user, for example the manager."),
			col("relationship_type", tStr, "relationshipType", "Relationship type, e.g. USER_RELATIONSHIP_TYPE_MANAGER."),
			col("created_at", tTS, "createdAt", "Creation time."),
		},
	}
}

//// serval_app_instance — applications connected to a workspace.

func tableServalAppInstance() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_app_instance",
		Description: "App instances (connected applications) in each workspace, with their owners.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/app-instances", 200, nil),
			KeyColumns: withTeamQual(nil),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/app-instances")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "App instance ID."),
			teamCol(),
			col("name", tStr, "name", "Name."),
			col("service", tStr, "service", "Built-in service slug, e.g. okta or github."),
			col("custom_service_id", tStr, "customServiceId", "Custom service ID, for apps without a built-in integration."),
			col("custom_service_name", tStr, "customServiceName", "Custom service name."),
			col("external_service_instance_id", tStr, "externalServiceInstanceId", "Identifier of the instance in the external service, e.g. its domain."),
			col("access_requests_enabled", tBool, "accessRequestsEnabled", "Whether users can request access to the app."),
			col("default_access_policy_id", tStr, "defaultAccessPolicyId", "Default access policy for the app's roles."),
			col("owners", tJSON, "owners", "Owners: userIds and groupIds."),
			{Name: "owner_user_ids", Type: tJSON, Transform: transform.FromField("owners.userIds").Transform(emptyArray), Description: "IDs of the users who own the app."},
			{Name: "owner_group_ids", Type: tJSON, Transform: transform.FromField("owners.groupIds").Transform(emptyArray), Description: "IDs of the groups that own the app."},
			col("integration_ids", tJSON, "integrationIds", "Integrations connected to the app."),
			col("tag_ids", tJSON, "tagIds", "Tag IDs."),
		},
	}
}

//// serval_app_resource — resources (apps, groups, ...) inside an app instance.

var resourceQuals = []qualParam{{"app_instance_id", "appInstanceId"}}

func tableServalAppResource() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_app_resource",
		Description: "Resources of each app instance (the app itself, groups, projects, ...) that roles grant access to.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/app-resources", 500, resourceQuals),
			KeyColumns: withTeamQual(optionalQuals(resourceQuals)),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/app-resources")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Resource ID."),
			teamCol(),
			col("app_instance_id", tStr, "appInstanceId", "App instance the resource belongs to."),
			col("resource_type", tStr, "resourceType", "Resource type, e.g. access or group."),
			col("name", tStr, "name", "Name."),
			col("description", tStr, "description", "Description."),
			col("external_id", tStr, "externalId", "ID in the external system."),
			col("scoped_id", tStr, "scopedId", "Stable scoped ID (ar_...)."),
			col("tag_ids", tJSON, "tagIds", "Tag IDs."),
		},
	}
}

//// serval_app_resource_role — requestable roles (entitlements) on a resource.

var roleQuals = []qualParam{{"resource_id", "resourceId"}}

func tableServalAppResourceRole() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_app_resource_role",
		Description: "Roles (entitlements) on app resources, with their access policy and provisioning method.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/app-resource-roles", 500, roleQuals),
			KeyColumns: withTeamQual(optionalQuals(roleQuals)),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/app-resource-roles")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Role ID."),
			teamCol(),
			col("resource_id", tStr, "resourceId", "Resource the role belongs to."),
			col("name", tStr, "name", "Name."),
			col("description", tStr, "description", "Description."),
			col("requests_enabled", tBool, "requestsEnabled", "Whether users can request the role."),
			col("access_policy_id", tStr, "accessPolicyId", "Access policy governing requests for the role."),
			col("provisioning_method", tJSON, "provisioningMethod", "How access is provisioned (builtin, custom workflow, linked roles, manual)."),
			col("is_ingested", tBool, "isIngested", "Whether the role was imported from the external system."),
			col("external_id", tStr, "externalId", "ID in the external system."),
			col("require_form_confirmation", tBool, "requireFormConfirmation", "Whether requests are collected with a form."),
			col("scoped_id", tStr, "scopedId", "Stable scoped ID (role_...)."),
			col("tag_ids", tJSON, "tagIds", "Tag IDs."),
		},
	}
}

//// serval_access_policy — approval and duration rules for access requests.

func tableServalAccessPolicy() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_access_policy",
		Description: "Access policies: approval procedure, maximum duration and justification requirements for access requests.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/access-policies", 500, nil),
			KeyColumns: withTeamQual(nil),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/access-policies")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Policy ID."),
			teamCol(),
			col("name", tStr, "name", "Name."),
			col("description", tStr, "description", "Description."),
			col("approval_procedure_id", tStr, "approvalProcedureId", "Approval procedure that must approve requests. Null means requests are granted without approval."),
			col("max_access_minutes", tInt, "maxAccessMinutes", "Maximum access duration in minutes. Null means access does not expire."),
			col("recommended_access_minutes", tInt, "recommendedAccessMinutes", "Recommended access duration in minutes."),
			col("require_business_justification", tBool, "requireBusinessJustification", "Whether a business justification is required."),
			col("require_on_behalf_of_confirmation", tBool, "requireOnBehalfOfConfirmation", "Whether requests made for someone else need their confirmation."),
			col("require_form_confirmation", tBool, "requireFormConfirmation", "Whether requests are collected with a form."),
			col("scoped_id", tStr, "scopedId", "Stable scoped ID (apo_...)."),
		},
	}
}

//// serval_access_request

func tableServalAccessRequest() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_access_request",
		Description: "Access requests for roles, with their status and time allocations (initial grant and extensions).",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/access-requests", 500, nil),
			KeyColumns: withTeamQual(nil),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/access-requests")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Request ID."),
			teamCol(),
			col("status", tStr, "status", "Overall status, e.g. ACCESS_REQUEST_STATUS_APPROVED."),
			col("target_user_id", tStr, "targetUserId", "User the access is for."),
			col("requested_role_id", tStr, "requestedRoleId", "Requested role."),
			col("linked_ticket_id", tStr, "linkedTicketId", "Ticket the request came from."),
			col("time_allocations", tJSON, "timeAllocations", "Time allocations: each grant or extension, with requester, justification, status and minutes."),
			col("expires_at", tTS, "expiresAt", "When the active time allocation expires."),
			col("created_at", tTS, "createdAt", "Creation time."),
		},
	}
}

//// serval_approval_request

func tableServalApprovalRequest() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_approval_request",
		Description: "Approval requests raised for access requests and workflow runs. The steps column fetches each request's steps and approver decisions (one API call per row).",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/approval-requests", 500, nil),
			KeyColumns: withTeamQual(nil),
		},
		Get:           &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/approval-requests")},
		HydrateConfig: []plugin.HydrateConfig{{Func: getApprovalSteps}},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Approval request ID."),
			teamCol(),
			col("result", tStr, "result", "Overall result: PENDING, APPROVED, DENIED, ..."),
			col("access_request_time_allocation_id", tStr, "accessRequestTimeAllocationId", "Access request time allocation this approval is for."),
			col("workflow_run_id", tStr, "workflowRunId", "Workflow run this approval is for."),
			col("approver_justification_mode", tStr, "approverJustificationMode", "Whether approvers must justify their decision."),
			{Name: "steps", Type: tJSON, Hydrate: getApprovalSteps, Transform: transform.FromValue(), Description: "Steps, each with its result and approver resolutions (approverUserId, result, autoApproved)."},
			col("created_at", tTS, "createdAt", "Creation time."),
			col("updated_at", tTS, "updatedAt", "Last update time."),
		},
	}
}

func getApprovalSteps(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	row, ok := h.Item.(Row)
	if !ok {
		return nil, nil
	}
	id, _ := row["id"].(string)
	if id == "" {
		return nil, nil
	}
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data []Row `json:"data"`
	}
	if err := c.do(ctx, "GET", "/v2/approval-requests/"+url.PathEscape(id)+"/steps", nil, nil, &res); err != nil {
		if err == errNotFound {
			return nil, nil
		}
		return nil, err
	}
	return res.Data, nil
}

//// serval_entity_ingestion_config — scheduled imports into Serval databases.

func tableServalEntityIngestionConfig() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_entity_ingestion_config",
		Description: "Entity ingestion configs: scheduled imports (devices, SaaS apps, ...) into Serval databases.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/entity-ingestion-configs", 500, []qualParam{{"app_instance_id", "appInstanceId"}}),
			KeyColumns: withTeamQual(optionalQuals([]qualParam{{"app_instance_id", "appInstanceId"}})),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/entity-ingestion-configs")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Ingestion config ID."),
			teamCol(),
			col("app_instance_id", tStr, "appInstanceId", "App instance the data comes from."),
			col("service_name", tStr, "serviceName", "Source service name."),
			col("workflow_id", tStr, "workflowId", "Ingestion workflow."),
			col("workflow_name", tStr, "workflowName", "Ingestion workflow name."),
			col("is_workflow_published", tBool, "isWorkflowPublished", "Whether the workflow is published."),
			col("enabled", tBool, "enabled", "Whether scheduled syncs run."),
			{Name: "sync_interval_seconds", Type: tInt, Transform: transform.FromField("syncIntervalSeconds").Transform(toInt), Description: "Minimum interval between scheduled syncs. Null means the platform default."},
			col("sync_deletions", tBool, "syncDeletions", "Whether records removed at the source are removed."),
			col("internal_entity_type_id", tStr, "internalEntityTypeId", "Entity type (database) the records land in."),
			col("allow_create_new_entities", tBool, "allowCreateNewEntities", "Whether a sync may create new records."),
			col("has_ingestion_context", tBool, "hasIngestionContext", "Whether the workflow uses the ingestion context."),
			col("created_at", tTS, "createdAt", "Creation time."),
			col("updated_at", tTS, "updatedAt", "Last update time."),
		},
	}
}

//// serval_workflow

func tableServalWorkflow() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_workflow",
		Description: "Workflows (automation code) in each workspace.",
		List: &plugin.ListConfig{
			Hydrate:    teamListHydrate("/v2/workflows", 200, nil),
			KeyColumns: withTeamQual(nil),
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/workflows")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Workflow ID."),
			teamCol(),
			col("name", tStr, "name", "Name."),
			col("description", tStr, "description", "Description."),
			col("is_published", tBool, "isPublished", "Whether the workflow is published."),
			col("has_unpublished_changes", tBool, "hasUnpublishedChanges", "Whether there are unpublished changes."),
			col("approval_procedure_id", tStr, "approvalProcedureId", "Approval procedure every run needs. Null means runs need no approval."),
			col("show_in_service_catalog", tBool, "showInServiceCatalog", "Whether users can run it from the service catalog."),
			col("require_form_confirmation", tBool, "requireFormConfirmation", "Whether runs are collected with a form."),
			col("package", tStr, "package", "Access package whose connections the workflow uses."),
			col("folder_id", tStr, "folderId", "Folder."),
			col("scoped_id", tStr, "scopedId", "Stable scoped ID (wf_...)."),
			col("tag_ids", tJSON, "tagIds", "Tag IDs."),
			col("content", tStr, "content", "Workflow source code."),
		},
	}
}

//// transforms

func isSet(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return false, nil
	}
	if s, ok := d.Value.(string); ok {
		return s != "", nil
	}
	return true, nil
}

func emptyArray(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if d.Value == nil {
		return []interface{}{}, nil
	}
	return d.Value, nil
}

func toInt(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch v := d.Value.(type) {
	case nil:
		return nil, nil
	case float64:
		return int64(v), nil
	case string:
		if v == "" {
			return nil, nil
		}
		return strconv.ParseInt(v, 10, 64)
	}
	return d.Value, nil
}
