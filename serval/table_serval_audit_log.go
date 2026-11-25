package serval

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// auditRetention is how far back Serval keeps audit events.
const auditRetention = 90 * 24 * time.Hour

func tableServalAuditLog() *plugin.Table {
	return &plugin.Table{
		Name: "serval_audit_log",
		Description: "The organisation's Serval audit log (last 90 days, newest first). Filter timestamp " +
			"(>, >=, <, <=) to bound the scan. Needs the org:audit scope.",
		List: &plugin.ListConfig{
			Hydrate:    listAuditLogs,
			KeyColumns: []*plugin.KeyColumn{{Name: "timestamp", Require: plugin.Optional, Operators: timeQuals}},
		},
		Columns: []*plugin.Column{
			col("id", tStr, "id", "Event ID."),
			col("timestamp", tTS, "timestamp", "When the event happened."),
			{Name: "event_type", Type: tStr, Transform: transform.FromField("eventType").Transform(oneofKey), Description: "Event type, e.g. userLogin, apiKeyCreated, accessProvisioned."},
			{Name: "event", Type: tJSON, Transform: transform.FromField("eventType").Transform(oneofValue), Description: "Event details (the payload of event_type)."},
			col("team_id", tStr, "teamId", "Workspace the event belongs to, when it is workspace-specific."),
			col("severity", tStr, "severity", "SEVERITY_INFO, SEVERITY_WARN, ..."),
			col("outcome_status", tStr, "outcome.status", "STATUS_SUCCESS, STATUS_DENIED, STATUS_FAILURE, ..."),
			col("outcome_reason_code", tStr, "outcome.reasonCode", "Machine-readable outcome reason."),
			col("outcome_reason_detail", tStr, "outcome.reasonDetail", "Outcome detail."),
			col("actor_id", tStr, "actor.id", "Acting user, API key or system ID."),
			{Name: "actor_type", Type: tStr, Transform: transform.FromField("actor").Transform(actorType), Description: "user, apiKey, worker or system."},
			col("actor_email", tStr, "actor.user.email", "Acting user's email."),
			{Name: "actor_display_name", Type: tStr, Transform: transform.FromField("actor").Transform(actorDisplayName), Description: "Acting user's, API key's or system's display name."},
			col("actor", tJSON, "actor", "Actor object."),
			col("ip_address", tStr, "context.ipAddress", "Source IP address."),
			col("user_agent", tStr, "context.userAgent", "Source user agent."),
			col("organization_id", tStr, "organizationId", "Organisation ID."),
		},
	}
}

func listAuditLogs(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	// The API filters only when both ends are set, both inclusive. Fill the open end
	// with the retention window or now; Postgres re-applies the exact predicate.
	params := url.Values{}
	from, to := timeRange(d, "timestamp")
	if !from.IsZero() || !to.IsZero() {
		if from.IsZero() {
			from = time.Now().Add(-auditRetention)
		}
		if to.IsZero() {
			to = time.Now().Add(time.Minute)
		}
		params.Set("startTime", from.Format(time.RFC3339Nano))
		params.Set("endTime", to.Format(time.RFC3339Nano))
	}
	err = c.paginate(ctx, http.MethodGet, "/v2/audit-logs", params, nil, 1000, func(r Row) bool { return stream(ctx, d, r) })
	if err != nil {
		return nil, fmt.Errorf("listing audit logs: %w", err)
	}
	return nil, nil
}

// oneofKey returns the single key of a protobuf oneof rendered as an object.
func oneofKey(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if m, ok := d.Value.(map[string]interface{}); ok {
		for k := range m {
			return k, nil
		}
	}
	return nil, nil
}

// oneofValue returns the value of a protobuf oneof rendered as an object.
func oneofValue(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if m, ok := d.Value.(map[string]interface{}); ok {
		for _, v := range m {
			return v, nil
		}
	}
	return nil, nil
}

// actorType returns the kind of actor: the key next to "id" in the actor object.
func actorType(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if m, ok := d.Value.(map[string]interface{}); ok {
		for k, v := range m {
			if k != "id" && v != nil {
				return k, nil
			}
		}
	}
	return nil, nil
}

func actorDisplayName(_ context.Context, d *transform.TransformData) (interface{}, error) {
	if m, ok := d.Value.(map[string]interface{}); ok {
		for k, v := range m {
			if k == "id" {
				continue
			}
			if inner, ok := v.(map[string]interface{}); ok {
				if s, ok := inner["displayName"].(string); ok {
					return s, nil
				}
			}
		}
	}
	return nil, nil
}
