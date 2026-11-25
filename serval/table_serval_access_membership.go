package serval

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// membershipWorkers bounds the parallel per-user calls of a full scan. The shared
// rate limiter still caps the overall request rate.
const membershipWorkers = 8

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func tableServalAccessMembership() *plugin.Table {
	return &plugin.Table{
		Name: "serval_access_membership",
		Description: "Role memberships recorded by Serval access management (provisioned, synced or linked-group). " +
			"The API answers one user at a time: filter user_id to query one user, otherwise every user " +
			"(including deactivated ones) is queried in every workspace, one call each.",
		List: &plugin.ListConfig{
			Hydrate: listAccessMemberships,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "user_id", Require: plugin.Optional},
				{Name: "team_id", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			col("user_id", tStr, "userId", "User holding the role."),
			teamCol(),
			col("app_instance_id", tStr, "appInstanceId", "App instance."),
			col("app_instance_name", tStr, "appInstanceName", "App instance name."),
			col("resource_id", tStr, "resourceId", "Resource."),
			col("resource_name", tStr, "resourceName", "Resource name."),
			col("role_id", tStr, "roleId", "Role."),
			col("role_name", tStr, "roleName", "Role name."),
		},
	}
}

func listAccessMemberships(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	c, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}
	teams, err := selectedTeams(ctx, d, c)
	if err != nil {
		return nil, err
	}

	var users []string
	// A qual fed by a scalar subquery (user_id = (select ...)) reaches the plugin as
	// "?", not as the value. Treat anything that is not a user ID as no filter: the
	// scan then covers every user and Postgres applies the predicate itself.
	if u := d.EqualsQualString("user_id"); uuidPattern.MatchString(u) {
		users = []string{u}
	} else {
		params := url.Values{"includeDeactivated": {"true"}}
		err := c.paginate(ctx, http.MethodGet, "/v2/users", params, nil, 500, func(r Row) bool {
			if id, ok := r["id"].(string); ok {
				users = append(users, id)
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("listing users: %w", err)
		}
	}

	type job struct{ team, user string }
	jobs := make(chan job)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)
	for i := 0; i < membershipWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				body := Row{"teamId": j.team, "targetUserId": j.user}
				err := c.paginate(ctx, http.MethodPost, "/v2/access-memberships/search", nil, body, 200, func(r Row) bool {
					r["teamId"] = j.team
					if !stream(ctx, d, r) {
						cancel()
						return false
					}
					return true
				})
				if err != nil && ctx.Err() == nil {
					errOnce.Do(func() {
						firstErr = fmt.Errorf("searching memberships of user %s in workspace %s: %w", j.user, j.team, err)
						cancel()
					})
				}
			}
		}()
	}

feed:
	for _, t := range teams {
		for _, u := range users {
			select {
			case jobs <- job{t, u}:
			case <-ctx.Done():
				break feed
			}
		}
	}
	close(jobs)
	wg.Wait()
	return nil, firstErr
}
