package mock

// Datasources is mock data for the datasource list command.
var Datasources = []map[string]string{
	{"id": "pg_prod", "name": "pg_prod", "type": "postgres", "status": "connected", "datasets": "42", "created": "2025-11-01"},
	{"id": "sf_analytics", "name": "sf_analytics", "type": "snowflake", "status": "connected", "datasets": "18", "created": "2025-09-15"},
	{"id": "bq_staging", "name": "bq_staging", "type": "bigquery", "status": "disconnected", "datasets": "7", "created": "2025-12-03"},
}

// Datasets is mock data for the dataset list command.
var Datasets = []map[string]string{
	{"id": "pg_prod.public.orders", "name": "orders", "datasource": "pg_prod", "schema": "public", "status": "passing", "checks": "12", "updated": "2026-03-05"},
	{"id": "pg_prod.public.users", "name": "users", "datasource": "pg_prod", "schema": "public", "status": "failing", "checks": "8", "updated": "2026-03-05"},
	{"id": "pg_prod.public.products", "name": "products", "datasource": "pg_prod", "schema": "public", "status": "passing", "checks": "5", "updated": "2026-03-04"},
	{"id": "sf_analytics.raw.events", "name": "events", "datasource": "sf_analytics", "schema": "raw", "status": "passing", "checks": "3", "updated": "2026-03-03"},
	{"id": "sf_analytics.raw.sessions", "name": "sessions", "datasource": "sf_analytics", "schema": "raw", "status": "failing", "checks": "6", "updated": "2026-03-05"},
}

// Jobs is mock data for the job list command.
var Jobs = []map[string]string{
	{"id": "sc_abc123", "datasource": "pg_prod", "dataset": "orders", "type": "contract", "status": "failing", "date": "2026-03-05 08:12"},
	{"id": "sc_def456", "datasource": "pg_prod", "dataset": "users", "type": "monitor", "status": "passing", "date": "2026-03-05 06:45"},
	{"id": "sc_ghi789", "datasource": "sf_analytics", "dataset": "events", "type": "contract", "status": "passing", "date": "2026-03-04 22:30"},
	{"id": "sc_jkl012", "datasource": "pg_prod", "dataset": "products", "type": "contract", "status": "passing", "date": "2026-03-04 20:15"},
	{"id": "sc_mno345", "datasource": "sf_analytics", "dataset": "sessions", "type": "monitor", "status": "failing", "date": "2026-03-05 01:00"},
}

// Results is mock data for the results list command.
var Results = []map[string]string{
	{"dataset": "pg/prod/orders", "type": "check", "name": "row_count > 0", "status": "passing", "date": "2026-03-05 08:12"},
	{"dataset": "pg/prod/orders", "type": "check", "name": "no_nulls(order_id)", "status": "failing", "date": "2026-03-05 08:12"},
	{"dataset": "pg/prod/orders", "type": "check", "name": "no_nulls(customer_id)", "status": "passing", "date": "2026-03-05 08:12"},
	{"dataset": "pg/prod/users", "type": "monitor", "name": "daily_signups anomaly", "status": "alert", "date": "2026-03-05 06:45"},
	{"dataset": "pg/prod/users", "type": "check", "name": "valid_email_format", "status": "failing", "date": "2026-03-05 08:12"},
	{"dataset": "sf/raw/events", "type": "check", "name": "row_count > 1000", "status": "passing", "date": "2026-03-04 22:30"},
	{"dataset": "sf/raw/sessions", "type": "monitor", "name": "session_duration spike", "status": "alert", "date": "2026-03-05 01:00"},
}

// Incidents is mock data for the incident list command.
var Incidents = []map[string]string{
	{"id": "INC-001", "title": "Null order IDs in orders table", "dataset": "pg_prod.orders", "status": "open", "created": "2026-03-05 08:12"},
	{"id": "INC-002", "title": "User email format violations", "dataset": "pg_prod.users", "status": "open", "created": "2026-03-05 06:45"},
	{"id": "INC-003", "title": "Sessions anomaly resolved", "dataset": "sf_analytics.sessions", "status": "closed", "created": "2026-03-03 14:00"},
}

// NotificationRules is mock data for the notification rule list command.
var NotificationRules = []map[string]string{
	{"id": "rule_001", "name": "orders-failures", "source": "check", "alert": "fail-only", "dataset": "pg_prod.orders", "notify": "data-team@acme.com"},
	{"id": "rule_002", "name": "all-anomalies", "source": "monitor", "alert": "anomaly", "dataset": "(all)", "notify": "role_admin"},
	{"id": "rule_003", "name": "users-warn-fail", "source": "check", "alert": "warn-fail", "dataset": "pg_prod.users", "notify": "alice@acme.com"},
}

// Integrations is mock data for the notification integration list command.
var Integrations = []map[string]string{
	{"id": "int_slack_001", "name": "slack-data-alerts", "type": "slack", "status": "connected"},
	{"id": "int_wh_002", "name": "webhook-pagerduty", "type": "webhook", "status": "connected"},
	{"id": "int_teams_003", "name": "teams-engineering", "type": "teams", "status": "disconnected"},
}

// Roles is mock data for the role list command.
var Roles = []map[string]string{
	{"id": "role_admin", "name": "Admin", "scope": "global", "members": "3"},
	{"id": "role_editor", "name": "Editor", "scope": "global", "members": "12"},
	{"id": "role_viewer", "name": "Viewer", "scope": "global", "members": "34"},
	{"id": "role_ds_owner", "name": "Dataset Owner", "scope": "dataset", "members": "8"},
}

// Users is mock data for the users list command.
var Users = []map[string]string{
	{"id": "usr_001", "email": "alice@acme.com", "name": "Alice Chen", "role": "Admin", "status": "active"},
	{"id": "usr_002", "email": "bob@acme.com", "name": "Bob Patel", "role": "Editor", "status": "active"},
	{"id": "usr_003", "email": "carol@acme.com", "name": "Carol Smith", "role": "Viewer", "status": "active"},
	{"id": "usr_004", "email": "dave@acme.com", "name": "Dave Liu", "role": "Editor", "status": "active"},
}

// Groups is mock data for the users group list command.
var Groups = []map[string]string{
	{"id": "grp_001", "name": "Data Engineering", "members": "5", "role": "Editor"},
	{"id": "grp_002", "name": "Analytics", "members": "8", "role": "Viewer"},
	{"id": "grp_003", "name": "Platform", "members": "3", "role": "Admin"},
}

// Proposals is mock data for the contract proposal list command.
var Proposals = []map[string]string{
	{"id": "prop_001", "dataset": "pg_prod.orders", "status": "open", "message": "Add freshness check", "created": "2026-03-04"},
	{"id": "prop_002", "dataset": "pg_prod.users", "status": "open", "message": "Tighten null constraints", "created": "2026-03-03"},
	{"id": "prop_003", "dataset": "sf_analytics.events", "status": "done", "message": "Initial contract", "created": "2026-02-20"},
}

// ContractChecks is mock verification result for a contract verify run.
var ContractChecks = []struct {
	Name   string
	Value  string
	Status string // "pass" | "fail"
}{
	{"row_count > 0", "row_count = 48231", "pass"},
	{"no_nulls(order_id)", "null_count = 0", "pass"},
	{"no_nulls(customer_id)", "null_count = 142 (0.29%)", "fail"},
	{"order_status in (pending, processing, shipped, delivered, cancelled)", "invalid_count = 0", "pass"},
	{"freshness(created_at) < 24h", "last_value = 31h ago", "fail"},
	{"avg(order_value) between 50 and 500", "avg = 127.43", "pass"},
}

// LogLines is mock log output for job logs.
var LogLines = []string{
	"2026-03-05 08:12:00  [INFO]   Starting contract verification for pg_prod.orders",
	"2026-03-05 08:12:01  [INFO]   Connecting to datasource pg_prod (postgres)",
	"2026-03-05 08:12:02  [INFO]   Connection established",
	"2026-03-05 08:12:03  [INFO]   Running check: row_count > 0",
	"2026-03-05 08:12:04  [INFO]   ✓ row_count > 0 [row_count = 48231]",
	"2026-03-05 08:12:05  [INFO]   Running check: no_nulls(order_id)",
	"2026-03-05 08:12:06  [INFO]   ✓ no_nulls(order_id) [null_count = 0]",
	"2026-03-05 08:12:07  [INFO]   Running check: no_nulls(customer_id)",
	"2026-03-05 08:12:08  [WARN]   ✗ no_nulls(customer_id) [null_count = 142 (0.29%)]",
	"2026-03-05 08:12:09  [INFO]   Running check: freshness(created_at) < 24h",
	"2026-03-05 08:12:10  [WARN]   ✗ freshness(created_at) < 24h [last_value = 31h ago]",
	"2026-03-05 08:12:11  [INFO]   Verification complete: 4 passing, 2 failing",
}

// Monitors is mock data for the monitor list command.
var Monitors = []map[string]string{
	{"id": "mon_001", "dataset": "pg_prod.orders", "type": "column", "metric": "row-count", "status": "enabled", "last_run": "2026-03-05 08:12"},
	{"id": "mon_002", "dataset": "pg_prod.users", "type": "column", "metric": "missing-pct(email)", "status": "enabled", "last_run": "2026-03-05 06:45"},
	{"id": "mon_003", "dataset": "sf_analytics.events", "type": "dataset", "metric": "row-count-change", "status": "disabled", "last_run": "2026-03-04 22:30"},
	{"id": "mon_004", "dataset": "sf_analytics.sessions", "type": "custom", "metric": "daily_session_duration", "status": "enabled", "last_run": "2026-03-05 01:00"},
}

// Runners is mock data for the runner list command.
var Runners = []map[string]string{
	{"id": "runner_001", "name": "prod-runner", "status": "running", "version": "1.4.2", "last_seen": "2026-03-05 08:10"},
	{"id": "runner_002", "name": "staging-runner", "status": "running", "version": "1.4.1", "last_seen": "2026-03-05 07:55"},
	{"id": "runner_003", "name": "legacy-runner", "status": "offline", "version": "1.3.0", "last_seen": "2026-02-28 12:00"},
}

// ServiceAccounts is mock data for the iam service-account list command.
var ServiceAccounts = []map[string]string{
	{"id": "sa_001", "name": "ci-pipeline", "email": "ci@svc.acme.com", "created": "2025-10-01"},
	{"id": "sa_002", "name": "data-agent", "email": "agent@svc.acme.com", "created": "2025-11-15"},
}

// DashboardSummary holds mock data for the dashboard command.
var DashboardSummary = map[string]string{
	"organization":     "acme-corp",
	"profile":          "production",
	"datasources":      "3",
	"datasets":         "67",
	"passing_datasets": "61",
	"failing_datasets": "6",
	"open_incidents":   "2",
	"jobs_today":       "14",
}
