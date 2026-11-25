#!/usr/bin/env bash
# Run a smoke-test query against every serval table and report results.
# Exit code is the number of failed tables (0 = all passed).

set -euo pipefail

# Colour codes (disabled if not a terminal)
if [ -t 1 ]; then
  GREEN=$'\033[0;32m'
  RED=$'\033[0;31m'
  YELLOW=$'\033[0;33m'
  RESET=$'\033[0m'
else
  GREEN="" RED="" YELLOW="" RESET=""
fi

PASS=0
FAIL=0
SKIP=0

run_test() {
  local table="$1"
  local query="$2"
  printf "  %-40s" "$table"
  local output
  if output=$(steampipe query "$query" 2>&1); then
    printf "${GREEN}PASS${RESET}\n"
    ((PASS++)) || true
  else
    # A 403 / authorization error is treated as a skip, not a hard failure
    if echo "$output" | grep -qiE "403|unauthorized|forbidden"; then
      printf "${YELLOW}SKIP${RESET} (authorization error)\n"
      ((SKIP++)) || true
    else
      printf "${RED}FAIL${RESET}\n"
      echo "$output" | sed 's/^/    /'
      ((FAIL++)) || true
    fi
  fi
}


echo ""
echo "Serval Steampipe plugin — table smoke tests"
echo "========================================="
echo ""


run_test "serval_user"                    "select id, email, role, deactivated from serval_user limit 1"
run_test "serval_user (email)"            "select id from serval_user where email = (select email from serval_user limit 1)"
run_test "serval_user_relationship"       "select user_id, related_user_id, relationship_type from serval_user_relationship limit 1"
run_test "serval_app_instance"            "select id, team_id, name, owner_user_ids from serval_app_instance limit 1"
run_test "serval_app_resource"            "select id, team_id, app_instance_id, name from serval_app_resource limit 1"
run_test "serval_app_resource_role"       "select id, team_id, resource_id, name, access_policy_id from serval_app_resource_role limit 1"
run_test "serval_access_policy"           "select id, team_id, name, approval_procedure_id, max_access_minutes from serval_access_policy limit 1"
run_test "serval_access_request"          "select id, team_id, status, target_user_id from serval_access_request limit 1"
run_test "serval_approval_request"        "select id, team_id, result, steps from serval_approval_request limit 1"
run_test "serval_access_membership"       "select m.user_id, m.team_id, m.role_name from serval_user u join serval_access_membership m on m.user_id = u.id where u.role = 'USER_ROLE_ORG_ADMIN'"
run_test "serval_reclamation_candidate"   "select id, team_id, status from serval_reclamation_candidate limit 1"
run_test "serval_entity_ingestion_config" "select id, team_id, workflow_name, enabled, sync_interval_seconds from serval_entity_ingestion_config limit 1"
run_test "serval_entity_ingestion_run"    "select id, ingestion_config_id, status, started_at from serval_entity_ingestion_run where started_at > now() - interval '1 day' limit 1"
run_test "serval_workflow"                "select id, team_id, name, is_published from serval_workflow limit 1"
run_test "serval_ticket"                  "select id, team_id, type, status_name from serval_ticket where created_at > now() - interval '30 days' limit 1"
run_test "serval_audit_log"               "select id, timestamp, event_type, actor_type from serval_audit_log where timestamp > now() - interval '1 day' limit 1"
echo ""
echo "Results: ${GREEN}${PASS} passed${RESET}  ${YELLOW}${SKIP} skipped${RESET}  ${RED}${FAIL} failed${RESET}"
echo ""

exit "$FAIL"
