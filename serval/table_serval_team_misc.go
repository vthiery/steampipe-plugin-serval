package serval

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// serval_reclamation_candidate — grants flagged because the user stopped signing in.

func tableServalReclamationCandidate() *plugin.Table {
	return &plugin.Table{
		Name:        "serval_reclamation_candidate",
		Description: "Reclamation candidates: access grants flagged for review or reclaim because the user stopped signing in to the app.",
		List: &plugin.ListConfig{
			Hydrate: listReclamationCandidates,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
				{Name: "status", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Reclamation ID."),
			teamCol(),
			col("status", tStr, "status", "Lifecycle state: in review, reclaiming, reclaimed, dismissed, reclaim failed, ..."),
			col("user_id", tStr, "userId", "User whose access is flagged."),
			col("entitlement_id", tStr, "entitlementId", "Role (entitlement) the grant is on."),
			col("app_instance_id", tStr, "appInstanceId", "App instance."),
			col("resource_type", tStr, "resourceType", "Entitlement's resource type."),
			col("last_login_at", tTS, "lastLoginAt", "User's last sign-in to the app, the inactivity signal."),
			col("decided_by_user_id", tStr, "decidedByUserId", "Reviewer who concluded the review. Null for automatic reclaims and open reviews."),
			col("created_at", tTS, "createdAt", "When the grant was flagged."),
			col("concluded_at", tTS, "concludedAt", "When the reclamation concluded."),
		},
	}
}

func listReclamationCandidates(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	teams, err := selectedTeams(ctx, d, c)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		params := paramsFromQuals(d, []qualParam{{"status", "status"}})
		path := "/v2/teams/" + url.PathEscape(team) + "/reclamation-candidates"
		done := false
		err := c.paginate(ctx, http.MethodGet, path, params, nil, 500, func(r Row) bool {
			if _, ok := r["teamId"]; !ok {
				r["teamId"] = team
			}
			done = !stream(ctx, d, r)
			return !done
		})
		if err != nil {
			return nil, fmt.Errorf("listing reclamation candidates for workspace %s: %w", team, err)
		}
		if done {
			break
		}
	}
	return nil, nil
}

//// serval_entity_ingestion_run — history of each ingestion config's syncs.

func tableServalEntityIngestionRun() *plugin.Table {
	return &plugin.Table{
		Name: "serval_entity_ingestion_run",
		Description: "Runs of each entity ingestion config, newest first, with status and item counts. " +
			"Filter ingestion_config_id and started_at (>, >=) to bound the scan.",
		List: &plugin.ListConfig{
			ParentHydrate: teamListHydrate("/v2/entity-ingestion-configs", 500, nil),
			Hydrate:       listIngestionRuns,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "ingestion_config_id", Require: plugin.Optional},
				{Name: "team_id", Require: plugin.Optional},
				{Name: "started_at", Require: plugin.Optional, Operators: timeQuals},
			},
		},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Run ID."),
			col("ingestion_config_id", tStr, "ingestionConfigId", "Ingestion config."),
			teamCol(),
			col("status", tStr, "status", "ENTITY_INGESTION_RUN_STATUS_COMPLETED, ..._FAILED, ..._RUNNING, ..."),
			col("is_full_sync", tBool, "isFullSync", "Whether the run was a full sync."),
			col("started_at", tTS, "startedAt", "Start time."),
			col("completed_at", tTS, "completedAt", "Completion time."),
			col("items_added_count", tInt, "itemsAddedCount", "Records added."),
			col("items_updated_count", tInt, "itemsUpdatedCount", "Records updated."),
			col("items_deleted_count", tInt, "itemsDeletedCount", "Records deleted."),
			col("items_unchanged_count", tInt, "itemsUnchangedCount", "Records unchanged."),
			col("items_errored_count", tInt, "itemsErroredCount", "Records that failed."),
			col("items_publish_failed_count", tInt, "itemsPublishFailedCount", "Publish targets reported as failed."),
			col("publish_outcome", tStr, "publishOutcome", "published, partial, failed, not_ready or unknown."),
			col("failure_reason", tStr, "failureReason", "Pipeline phase that failed."),
			col("failure_code", tStr, "failureCode", "Machine-readable failure cause."),
			col("failure_message", tStr, "failureMessage", "Failure message."),
		},
	}
}

func listIngestionRuns(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	cfg, ok := h.Item.(Row)
	if !ok {
		return nil, nil
	}
	id, _ := cfg["id"].(string)
	if want := d.EqualsQualString("ingestion_config_id"); id == "" || (want != "" && want != id) {
		return nil, nil
	}
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	// Runs come newest first: stop paging once past the lower bound of started_at.
	from, _ := timeRange(d, "started_at")
	const size = 200
	path := "/v2/entity-ingestion-configs/" + url.PathEscape(id) + "/runs"
	for offset := 0; ; offset += size {
		var p page
		params := url.Values{"pageSize": {strconv.Itoa(size)}, "pageOffset": {strconv.Itoa(offset)}}
		if err := c.do(ctx, http.MethodGet, path, params, nil, &p); err != nil {
			return nil, fmt.Errorf("listing runs of ingestion config %s: %w", id, err)
		}
		for _, r := range p.Data {
			r["teamId"] = cfg["teamId"]
			if !from.IsZero() {
				if s, ok := r["startedAt"].(string); ok {
					if t, err := time.Parse(time.RFC3339Nano, s); err == nil && t.Before(from) {
						return nil, nil
					}
				}
			}
			if !stream(ctx, d, r) {
				return nil, nil
			}
		}
		if len(p.Data) < size {
			return nil, nil
		}
	}
}

//// serval_ticket — help desk tickets.

func tableServalTicket() *plugin.Table {
	return &plugin.Table{
		Name: "serval_ticket",
		Description: "Tickets in each workspace. Without a created_at filter the API returns the last year; " +
			"filter created_at (>, >=, <, <=) to change the window.",
		List: &plugin.ListConfig{
			Hydrate: listTickets,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "team_id", Require: plugin.Optional},
				{Name: "created_at", Require: plugin.Optional, Operators: timeQuals},
			},
		},
		Get: &plugin.GetConfig{KeyColumns: plugin.SingleColumn("id"), Hydrate: getHydrate("/v2/tickets")},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Ticket ID."),
			teamCol(),
			col("friendly_identifier", tStr, "friendlyIdentifier", "Human-readable identifier."),
			col("type", tStr, "type", "Ticket type, e.g. request, incident or change."),
			col("name", tStr, "name", "Title."),
			col("description", tStr, "description", "Description."),
			col("status_id", tStr, "statusId", "Status option ID."),
			col("status_name", tStr, "statusName", "Status, e.g. In Progress."),
			col("priority_id", tStr, "priorityId", "Priority option ID."),
			col("priority_name", tStr, "priorityName", "Priority, e.g. High."),
			col("escalation_level", tStr, "escalationLevel", "Escalation level."),
			col("is_overdue", tBool, "isOverdue", "Whether the due date has passed."),
			col("due_date", tStr, "dueDate", "Due date (YYYY-MM-DD)."),
			col("due_at", tTS, "dueAt", "Planned end."),
			col("created_by_user_id", tStr, "createdByUserId", "Creator."),
			col("requester_user_id", tStr, "requesterUserId", "Requester."),
			col("assigned_to_user_id", tStr, "assignedToUserId", "Assignee."),
			col("assigned_to_group_id", tStr, "assignedToGroupId", "Assignment group."),
			col("category_option_id", tStr, "categoryOptionId", "Category option."),
			col("parent_ticket_id", tStr, "parentTicketId", "Parent ticket, for tasks."),
			col("label_ids", tJSON, "labelIds", "Label IDs."),
			col("url", tStr, "url", "Ticket URL."),
			col("created_at", tTS, "createdAt", "Creation time."),
			col("escalated_at", tTS, "escalatedAt", "Escalation time."),
			col("completed_at", tTS, "completedAt", "Completion time."),
		},
	}
}

func listTickets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	teams, err := selectedTeams(ctx, d, c)
	if err != nil {
		return nil, err
	}
	from, to := timeRange(d, "created_at")
	for _, team := range teams {
		params := url.Values{"teamId": {team}}
		if !from.IsZero() {
			params.Set("startTime", from.Format(time.RFC3339Nano))
		}
		if !to.IsZero() {
			params.Set("endTime", to.Format(time.RFC3339Nano))
		}
		done := false
		err := c.paginate(ctx, http.MethodGet, "/v2/tickets", params, nil, 2000, func(r Row) bool {
			done = !stream(ctx, d, r)
			return !done
		})
		if err != nil {
			return nil, fmt.Errorf("listing tickets for workspace %s: %w", team, err)
		}
		if done {
			break
		}
	}
	return nil, nil
}
