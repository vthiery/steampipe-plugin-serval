package serval

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const (
	tStr  = proto.ColumnType_STRING
	tBool = proto.ColumnType_BOOL
	tInt  = proto.ColumnType_INT
	tTS   = proto.ColumnType_TIMESTAMP
	tJSON = proto.ColumnType_JSON
)

// col builds a column that reads the camelCase API field `field`.
func col(name string, t proto.ColumnType, field, desc string) *plugin.Column {
	return &plugin.Column{Name: name, Type: t, Transform: transform.FromField(field), Description: desc}
}

// teamCol is the team_id column every workspace-scoped table carries.
func teamCol() *plugin.Column {
	return col("team_id", tStr, "teamId", "ID of the workspace (team) the object belongs to.")
}

// qualParam maps an equality qual on a column to an API query parameter.
type qualParam struct {
	Column string
	Param  string
}

// optionalQuals turns qual mappings into optional key columns. Boolean columns also
// accept "<>": Postgres rewrites `col = false` into `col <> true`.
func optionalQuals(qs []qualParam, boolCols ...string) []*plugin.KeyColumn {
	isBool := map[string]bool{}
	for _, c := range boolCols {
		isBool[c] = true
	}
	kcs := make([]*plugin.KeyColumn, 0, len(qs))
	for _, q := range qs {
		kc := &plugin.KeyColumn{Name: q.Column, Require: plugin.Optional}
		if isBool[q.Column] {
			kc.Operators = []string{"=", "<>"}
		}
		kcs = append(kcs, kc)
	}
	return kcs
}

// withTeamQual adds the optional team_id key column.
func withTeamQual(kcs []*plugin.KeyColumn) []*plugin.KeyColumn {
	return append(kcs, &plugin.KeyColumn{Name: "team_id", Require: plugin.Optional})
}

// paramsFromQuals copies equality quals into API query parameters.
func paramsFromQuals(d *plugin.QueryData, qs []qualParam) url.Values {
	params := url.Values{}
	for _, q := range qs {
		quals := d.Quals[q.Column]
		if quals == nil {
			continue
		}
		for _, qual := range quals.Quals {
			if qual.Value == nil {
				continue
			}
			switch x := qual.Value.Value.(type) {
			case *proto.QualValue_StringValue:
				if qual.Operator == "=" {
					params.Set(q.Param, x.StringValue)
				}
			case *proto.QualValue_BoolValue:
				switch qual.Operator {
				case "=":
					params.Set(q.Param, strconv.FormatBool(x.BoolValue))
				case "<>":
					params.Set(q.Param, strconv.FormatBool(!x.BoolValue))
				}
			}
		}
	}
	return params
}

// selectedTeams returns the workspaces a query should scan: the team_id qual when given,
// otherwise every configured workspace.
func selectedTeams(ctx context.Context, d *plugin.QueryData, c *Client) ([]string, error) {
	if t := d.EqualsQualString("team_id"); t != "" {
		return []string{t}, nil
	}
	return c.teams(ctx)
}

// stream sends a row to Steampipe and reports whether the scan should continue.
func stream(ctx context.Context, d *plugin.QueryData, row Row) bool {
	d.StreamListItem(ctx, row)
	return d.RowsRemaining(ctx) != 0
}

// listHydrate returns a List hydrate for an organisation-wide GET endpoint.
func listHydrate(path string, pageSize int, qs []qualParam, extra url.Values) plugin.HydrateFunc {
	return func(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
		c, err := getClient(ctx, d)
		if err != nil {
			return nil, err
		}
		params := paramsFromQuals(d, qs)
		for k, v := range extra {
			params[k] = v
		}
		err = c.paginate(ctx, http.MethodGet, path, params, nil, pageSize, func(r Row) bool { return stream(ctx, d, r) })
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", path, err)
		}
		return nil, nil
	}
}

// teamListHydrate returns a List hydrate for a workspace-scoped GET endpoint taking a
// teamId query parameter. It scans each selected workspace in turn and sets teamId on
// rows that do not carry it.
func teamListHydrate(path string, pageSize int, qs []qualParam) plugin.HydrateFunc {
	return func(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
		c, err := getClient(ctx, d)
		if err != nil {
			return nil, err
		}
		teams, err := selectedTeams(ctx, d, c)
		if err != nil {
			return nil, err
		}
		for _, team := range teams {
			params := paramsFromQuals(d, qs)
			params.Set("teamId", team)
			done := false
			err := c.paginate(ctx, http.MethodGet, path, params, nil, pageSize, func(r Row) bool {
				if _, ok := r["teamId"]; !ok {
					r["teamId"] = team
				}
				if !stream(ctx, d, r) {
					done = true
					return false
				}
				return true
			})
			if err != nil && !isForeignWorkspace(err) {
				return nil, fmt.Errorf("listing %s for workspace %s: %w", path, team, err)
			}
			if done {
				break
			}
		}
		return nil, nil
	}
}

// getHydrate returns a Get hydrate fetching basePath/{id}.
func getHydrate(basePath string) plugin.HydrateFunc {
	return func(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
		id := d.EqualsQualString("id")
		if id == "" {
			return nil, nil
		}
		c, err := getClient(ctx, d)
		if err != nil {
			return nil, err
		}
		row, err := c.getOne(ctx, basePath+"/"+url.PathEscape(id))
		if err != nil {
			return nil, fmt.Errorf("getting %s/%s: %w", basePath, id, err)
		}
		if row == nil {
			return nil, nil
		}
		return row, nil
	}
}

// timeRange reads >, >=, <, <=, = quals on a timestamp column. Zero values mean unbounded.
func timeRange(d *plugin.QueryData, column string) (from, to time.Time) {
	q := d.Quals[column]
	if q == nil {
		return
	}
	for _, qual := range q.Quals {
		ts := qual.Value.GetTimestampValue()
		if ts == nil {
			continue
		}
		t := ts.AsTime().UTC()
		switch qual.Operator {
		case ">", ">=":
			if from.IsZero() || t.After(from) {
				from = t
			}
		case "<", "<=":
			if to.IsZero() || t.Before(to) {
				to = t
			}
		case "=":
			from, to = t, t
		}
	}
	return
}

// timeQuals are the operators timestamp key columns accept.
var timeQuals = []string{">", ">=", "<", "<=", "="}
