"""All hardcoded mock data. Swap individual sections for real API calls later."""

DATASOURCES = [
    {
        "id": "pg_prod",
        "name": "pg_prod",
        "type": "postgres",
        "host": "db.prod.acme.internal",
        "status": "connected",
        "datasets": 42,
        "created": "2025-11-01",
    },
    {
        "id": "sf_analytics",
        "name": "sf_analytics",
        "type": "snowflake",
        "host": "acme.snowflakecomputing.com",
        "status": "connected",
        "datasets": 18,
        "created": "2025-12-15",
    },
    {
        "id": "bq_ml",
        "name": "bq_ml",
        "type": "bigquery",
        "host": "bigquery.googleapis.com",
        "status": "degraded",
        "datasets": 7,
        "created": "2026-01-20",
    },
]

DATASETS = [
    {
        "id": "ds_001",
        "name": "orders",
        "datasource": "pg_prod",
        "fqn": "pg_prod/prod_db/public/orders",
        "schema": "public",
        "rows": "2.4M",
        "last_scan": "2026-03-05 08:12",
        "status": "failing",
    },
    {
        "id": "ds_002",
        "name": "users",
        "datasource": "pg_prod",
        "fqn": "pg_prod/prod_db/public/users",
        "schema": "public",
        "rows": "890K",
        "last_scan": "2026-03-05 06:45",
        "status": "passing",
    },
    {
        "id": "ds_003",
        "name": "products",
        "datasource": "pg_prod",
        "fqn": "pg_prod/prod_db/public/products",
        "schema": "public",
        "rows": "15K",
        "last_scan": "2026-03-05 04:00",
        "status": "passing",
    },
    {
        "id": "ds_004",
        "name": "revenue_daily",
        "datasource": "sf_analytics",
        "fqn": "sf_analytics/ANALYTICS/PROD/revenue_daily",
        "schema": "PROD",
        "rows": "730",
        "last_scan": "2026-03-05 00:01",
        "status": "passing",
    },
    {
        "id": "ds_005",
        "name": "ml_features",
        "datasource": "bq_ml",
        "fqn": "bq_ml/acme-ml/features/ml_features",
        "schema": "features",
        "rows": "5.1M",
        "last_scan": "2026-03-04 22:00",
        "status": "failing",
    },
]

JOBS = [
    {
        "id": "sc_abc123",
        "datasource": "pg_prod",
        "dataset": "orders",
        "type": "contract",
        "status": "failing",
        "duration": "4.2s",
        "date": "2026-03-05 08:12",
    },
    {
        "id": "sc_def456",
        "datasource": "pg_prod",
        "dataset": "users",
        "type": "monitor",
        "status": "passing",
        "duration": "2.1s",
        "date": "2026-03-05 06:45",
    },
    {
        "id": "sc_ghi789",
        "datasource": "pg_prod",
        "dataset": "products",
        "type": "contract",
        "status": "passing",
        "duration": "1.8s",
        "date": "2026-03-05 04:00",
    },
    {
        "id": "sc_jkl012",
        "datasource": "sf_analytics",
        "dataset": "revenue_daily",
        "type": "contract",
        "status": "passing",
        "duration": "8.3s",
        "date": "2026-03-05 00:01",
    },
    {
        "id": "sc_mno345",
        "datasource": "bq_ml",
        "dataset": "ml_features",
        "type": "contract",
        "status": "error",
        "duration": "0.3s",
        "date": "2026-03-04 22:00",
    },
]

JOB_LOGS = {
    "sc_abc123": [
        "[08:12:00] Starting contract verification for pg_prod/prod_db/public/orders",
        "[08:12:00] Connecting to pg_prod...",
        "[08:12:01] Connected. Running 8 checks.",
        "[08:12:02] ✓ row_count > 0  (2,412,847 rows)",
        "[08:12:02] ✓ freshness < 1d  (last row: 2026-03-05 07:58:12)",
        "[08:12:03] ✓ no_nulls(order_id)",
        "[08:12:03] ✗ no_nulls(customer_id)  — 143 nulls found",
        "[08:12:04] ✓ valid_values(status)  in [pending, processing, shipped, delivered, cancelled]",
        "[08:12:04] ✗ reference(customer_id) → users(id)  — 23 orphaned rows",
        "[08:12:05] ✓ min(amount) >= 0",
        "[08:12:05] ✓ unique(order_id)",
        "[08:12:05] Verification complete: 6 passed, 2 failed",
        "[08:12:05] Exit code: 1",
    ]
}

RESULTS = [
    {
        "dataset": "pg_prod/prod_db/public/orders",
        "type": "check",
        "name": "row_count > 0",
        "status": "passing",
        "value": "2412847",
        "date": "2026-03-05 08:12",
    },
    {
        "dataset": "pg_prod/prod_db/public/orders",
        "type": "check",
        "name": "no_nulls(customer_id)",
        "status": "failing",
        "value": "143 nulls",
        "date": "2026-03-05 08:12",
    },
    {
        "dataset": "pg_prod/prod_db/public/orders",
        "type": "check",
        "name": "reference(customer_id) → users(id)",
        "status": "failing",
        "value": "23 orphaned",
        "date": "2026-03-05 08:12",
    },
    {
        "dataset": "pg_prod/prod_db/public/orders",
        "type": "check",
        "name": "freshness < 1d",
        "status": "passing",
        "value": "7m ago",
        "date": "2026-03-05 08:12",
    },
    {
        "dataset": "pg_prod/prod_db/public/users",
        "type": "monitor",
        "name": "daily_signups anomaly",
        "status": "alert",
        "value": "↓ 68% below baseline",
        "date": "2026-03-05 06:45",
    },
    {
        "dataset": "pg_prod/prod_db/public/users",
        "type": "check",
        "name": "no_nulls(email)",
        "status": "passing",
        "value": "0 nulls",
        "date": "2026-03-05 06:45",
    },
    {
        "dataset": "sf_analytics/ANALYTICS/PROD/revenue_daily",
        "type": "check",
        "name": "row_count = 365",
        "status": "passing",
        "value": "365",
        "date": "2026-03-05 00:01",
    },
]

INCIDENTS = [
    {
        "id": "inc_001",
        "title": "Null customer_id in orders",
        "dataset": "pg_prod/prod_db/public/orders",
        "status": "open",
        "severity": "high",
        "opened": "2026-03-05 08:12",
        "updated": "2026-03-05 08:15",
    },
    {
        "id": "inc_002",
        "title": "Orphaned order references",
        "dataset": "pg_prod/prod_db/public/orders",
        "status": "open",
        "severity": "medium",
        "opened": "2026-03-05 08:12",
        "updated": "2026-03-05 09:00",
    },
    {
        "id": "inc_003",
        "title": "Daily signups anomaly",
        "dataset": "pg_prod/prod_db/public/users",
        "status": "open",
        "severity": "high",
        "opened": "2026-03-05 06:45",
        "updated": "2026-03-05 07:30",
    },
    {
        "id": "inc_004",
        "title": "Missing revenue rows",
        "dataset": "sf_analytics/ANALYTICS/PROD/revenue_daily",
        "status": "closed",
        "severity": "low",
        "opened": "2026-02-28 00:01",
        "updated": "2026-03-01 10:00",
    },
]

NOTIFICATIONS = [
    {
        "id": "ntf_001",
        "channel": "slack-data-alerts",
        "trigger": "check-failure",
        "dataset": "pg_prod/prod_db/public/orders",
        "status": "active",
    },
    {
        "id": "ntf_002",
        "channel": "slack-data-alerts",
        "trigger": "incident-opened",
        "dataset": "(all)",
        "status": "active",
    },
    {
        "id": "ntf_003",
        "channel": "webhook-pagerduty",
        "trigger": "check-failure",
        "dataset": "sf_analytics/ANALYTICS/PROD/revenue_daily",
        "status": "active",
    },
]

CHANNELS = [
    {
        "id": "ch_001",
        "name": "slack-data-alerts",
        "type": "slack",
        "workspace": "acme-corp.slack.com",
        "channel": "#data-alerts",
        "status": "connected",
    },
    {
        "id": "ch_002",
        "name": "webhook-pagerduty",
        "type": "webhook",
        "url": "https://events.pagerduty.com/integration/abc123/enqueue",
        "status": "connected",
    },
    {
        "id": "ch_003",
        "name": "teams-engineering",
        "type": "teams",
        "workspace": "acme-corp.teams.microsoft.com",
        "channel": "Engineering Alerts",
        "status": "disconnected",
    },
]

ROLES = [
    {
        "id": "role_001",
        "name": "Admin",
        "scope": "global",
        "permissions": "all",
        "members": 3,
    },
    {
        "id": "role_002",
        "name": "Data Engineer",
        "scope": "global",
        "permissions": "read, write contracts, run scans",
        "members": 12,
    },
    {
        "id": "role_003",
        "name": "Viewer",
        "scope": "global",
        "permissions": "read",
        "members": 28,
    },
    {
        "id": "role_004",
        "name": "Dataset Owner",
        "scope": "dataset",
        "permissions": "manage dataset, write contracts",
        "members": 8,
    },
]

USERS = [
    {
        "id": "usr_001",
        "email": "alice@acme.com",
        "name": "Alice Chen",
        "role": "Admin",
        "group": "data-platform",
        "last_active": "2026-03-05",
    },
    {
        "id": "usr_002",
        "email": "bob@acme.com",
        "name": "Bob Martinez",
        "role": "Data Engineer",
        "group": "data-platform",
        "last_active": "2026-03-05",
    },
    {
        "id": "usr_003",
        "email": "carol@acme.com",
        "name": "Carol White",
        "role": "Data Engineer",
        "group": "analytics",
        "last_active": "2026-03-04",
    },
    {
        "id": "usr_004",
        "email": "dave@acme.com",
        "name": "Dave Kim",
        "role": "Viewer",
        "group": "analytics",
        "last_active": "2026-03-03",
    },
]

GROUPS = [
    {
        "id": "grp_001",
        "name": "data-platform",
        "role": "Data Engineer",
        "members": 8,
        "created": "2025-10-01",
    },
    {
        "id": "grp_002",
        "name": "analytics",
        "role": "Viewer",
        "members": 14,
        "created": "2025-10-01",
    },
    {
        "id": "grp_003",
        "name": "data-leads",
        "role": "Admin",
        "members": 3,
        "created": "2025-11-15",
    },
]

PROPOSALS = [
    {
        "id": "prp_001",
        "dataset": "pg_prod/prod_db/public/orders",
        "status": "open",
        "revision": 2,
        "author": "alice@acme.com",
        "message": "Add freshness check",
        "created": "2026-03-04",
    },
    {
        "id": "prp_002",
        "dataset": "sf_analytics/ANALYTICS/PROD/revenue_daily",
        "status": "done",
        "revision": 1,
        "author": "bob@acme.com",
        "message": "Initial contract",
        "created": "2026-03-01",
    },
]

MONITORS = {
    "ds_001": [
        {
            "id": "mon_001",
            "column": "amount",
            "metric": "avg",
            "status": "active",
            "alert_threshold": "3σ",
            "last_run": "2026-03-05 08:12",
        },
        {
            "id": "mon_002",
            "column": "customer_id",
            "metric": "count_distinct",
            "status": "active",
            "alert_threshold": "2σ",
            "last_run": "2026-03-05 08:12",
        },
    ],
    "ds_002": [
        {
            "id": "mon_003",
            "column": "id",
            "metric": "row_count",
            "status": "active",
            "alert_threshold": "3σ",
            "last_run": "2026-03-05 06:45",
        },
    ],
}

PROFILE_DATA = {
    "ds_001": {
        "dataset": "orders",
        "datasource": "pg_prod",
        "rows": "2,412,847",
        "columns": 14,
        "last_profiled": "2026-03-05 08:12",
        "freshness": "7 minutes",
        "columns_detail": [
            {"column": "order_id", "type": "bigint", "nulls": "0%", "unique": "100%", "min": "1", "max": "2412847"},
            {"column": "customer_id", "type": "bigint", "nulls": "0.006%", "unique": "36.2%", "min": "1", "max": "890234"},
            {"column": "status", "type": "varchar", "nulls": "0%", "unique": "5 values", "min": "cancelled", "max": "shipped"},
            {"column": "amount", "type": "numeric", "nulls": "0%", "unique": "98.4%", "min": "0.99", "max": "4999.00"},
            {"column": "created_at", "type": "timestamp", "nulls": "0%", "unique": "99.9%", "min": "2020-01-01", "max": "2026-03-05"},
        ],
    }
}

AUTH_STATUS = {
    "profile": "production",
    "org": "acme-corp",
    "host": "cloud.soda.io",
    "user": "alice@acme.com",
    "status": "connected",
    "last_verified": "2026-03-05 08:00",
}

DASHBOARD_SUMMARY = {
    "org": "acme-corp",
    "as_of": "2026-03-05 09:00",
    "datasets": 67,
    "passing": 58,
    "failing": 7,
    "erroring": 2,
    "open_incidents": 3,
    "jobs_today": 24,
    "recent_failures": [
        {"dataset": "pg_prod/prod_db/public/orders", "check": "no_nulls(customer_id)", "time": "08:12"},
        {"dataset": "pg_prod/prod_db/public/orders", "check": "reference(customer_id)", "time": "08:12"},
        {"dataset": "bq_ml/acme-ml/features/ml_features", "check": "row_count > 0", "time": "22:00 (yesterday)"},
    ],
}
