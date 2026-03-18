  # ═══════════════════════════════════════════════════════════════════
  # AUTH
  # ═══════════════════════════════════════════════════════════════════

  # auth login
  soda auth login                                                          # interactive wizard
  soda auth login --host dev.sodadata.io --api-key-id <id> --api-key-secret <secret>
  soda auth login --host cloud.soda.io --api-key-id <id> --api-key-secret <secret> --profile eu-prod
  soda auth login --host cloud.us.soda.io --api-key-id <id> --api-key-secret <secret> --profile us-prod
  soda auth login --no-interactive                                         # expect: error (missing creds)

  # auth status
  soda auth status
  soda auth status --output json
  soda auth status --profile default

  # auth switch
  soda auth switch default
  soda auth switch eu-prod
  soda auth switch nonexistent-profile                                     # expect: error

  # auth logout
  soda auth logout
  soda auth logout --profile eu-prod

  # ═══════════════════════════════════════════════════════════════════
  # DATASOURCE
  # ═══════════════════════════════════════════════════════════════════

  # datasource list
  soda datasource list
  soda datasource list --output json
  soda datasource list --output csv
  soda datasource list --quiet
  soda ds list                                                             # alias

  # datasource get
  soda datasource get <ds-id>
  soda datasource get <ds-id> --output json
  soda datasource get bad-id                                               # expect: error
  soda ds get <ds-id>                                                      # alias

  # datasource create
  soda datasource create <config-file>
  soda datasource create <config-file> --runner <runner-id>
  soda datasource create nonexistent.yml                                   # expect: error

  # datasource update (API returns 404 — blocked)
  soda datasource update <ds-id> --label "New Label"
  soda datasource update <ds-id> --runner <runner-id>
  soda datasource update <ds-id> --config <config-file>
  soda datasource update <ds-id> --label "New" --runner <id> --config <file>
  soda datasource update bad-id --label "x"                                # expect: error

  # datasource delete
  soda datasource delete <ds-id>                                           # ⚠ destructive
  soda datasource delete bad-id                                            # expect: error

  # datasource test-connection (API returns HTML — blocked)
  soda datasource test-connection <config-file>
  soda datasource test-connection <config-file> --runner <runner-id>
  soda datasource test-connection nonexistent.yml                          # expect: error

  # datasource diagnostics — view (no flags)
  soda datasource diagnostics <ds-id>
  soda datasource diagnostics <ds-id> --output json

  # datasource diagnostics — enable/disable
  soda datasource diagnostics <ds-id> --enable
  soda datasource diagnostics <ds-id> --disable

  # datasource diagnostics — warehouse
  soda datasource diagnostics <ds-id> --warehouse same
  soda datasource diagnostics <ds-id> --warehouse <config-file>

  # datasource diagnostics — schema
  soda datasource diagnostics <ds-id> --schema soda_diagnostics
  soda datasource diagnostics <ds-id> --schema custom_schema

  # datasource diagnostics — collect toggles
  soda datasource diagnostics <ds-id> --collect-results
  soda datasource diagnostics <ds-id> --no-collect-results
  soda datasource diagnostics <ds-id> --collect-failed-rows
  soda datasource diagnostics <ds-id> --no-collect-failed-rows

  # datasource diagnostics — table naming
  soda datasource diagnostics <ds-id> --table-prefix "soda_"
  soda datasource diagnostics <ds-id> --table-suffix "_diag"
  soda datasource diagnostics <ds-id> --table-prefix "soda_" --table-suffix "_diag"

  # datasource diagnostics — failed rows options
  soda datasource diagnostics <ds-id> --failed-rows-description "Failed row context"
  soda datasource diagnostics <ds-id> --expose-failed-rows-query
  soda datasource diagnostics <ds-id> --no-expose-failed-rows-query
  soda datasource diagnostics <ds-id> --failed-rows-cta
  soda datasource diagnostics <ds-id> --no-failed-rows-cta

  # datasource diagnostics — combo
  soda datasource diagnostics <ds-id> --enable --collect-results --collect-failed-rows --schema custom --expose-failed-rows-query --failed-rows-cta

  # datasource diagnostics test-connection
  soda datasource diagnostics test-connection <ds-id>

  # datasource onboard — interactive
  soda datasource onboard <config-file>
  soda datasource onboard <ds-id>

  # datasource onboard — non-interactive combos
  soda datasource onboard <ds-id> --monitoring --profiling --contracts none
  soda datasource onboard <ds-id> --no-monitoring --no-profiling --contracts none
  soda datasource onboard <ds-id> --monitoring --no-profiling --contracts skeleton
  soda datasource onboard <ds-id> --no-monitoring --profiling --contracts ai
  soda datasource onboard <ds-id> --monitoring --profiling --contracts ai
  soda datasource onboard <config-file> --runner <runner-id> --monitoring --profiling --contracts none
  soda datasource onboard <ds-id> --no-interactive --monitoring --profiling --contracts none

  # ═══════════════════════════════════════════════════════════════════
  # DATASET
  # ═══════════════════════════════════════════════════════════════════

  # dataset list — no filters
  soda dataset list
  soda dataset list --output json
  soda dataset list --output csv

  # dataset list — individual filters
  soda dataset list --datasource <ds-name>
  soda dataset list --filter "orders"
  soda dataset list --id <partial-id>
  soda dataset list --status onboarded
  soda dataset list --status not-onboarded
  soda dataset list --tag <tag-name>
  soda dataset list --limit 5
  soda dataset list --limit 50
  soda dataset list --from 2026-03-01
  soda dataset list --until 2026-03-18
  soda dataset list --from 2026-03-01 --until 2026-03-18

  # dataset list — combined filters
  soda dataset list --datasource <ds-name> --status onboarded --limit 10
  soda dataset list --datasource <ds-name> --filter "orders" --tag production --from 2026-01-01

  # dataset get
  soda dataset get <dataset-id>
  soda dataset get <dataset-id> --output json
  soda dataset get bad-id                                                  # expect: error

  # dataset update
  soda dataset update <dataset-id> --owner <user-id>
  soda dataset update <dataset-id> --tag newtag
  soda dataset update <dataset-id> --tag tag1 --tag tag2                   # replaces all tags
  soda dataset update <dataset-id> --owner <user-id> --tag tag1 --tag tag2

  # dataset delete
  soda dataset delete <dataset-id>                                         # ⚠ destructive
  soda dataset delete bad-id                                               # expect: error

  # dataset time-partition
  soda dataset time-partition <dataset-id>                                 # view current
  soda dataset time-partition <dataset-id> --column created_at             # set column
  soda dataset time-partition <dataset-id> --column updated_at

  # dataset profiling — view
  soda dataset profiling <dataset-id>
  soda dataset profiling <dataset-id> --output json

  # dataset profiling — enable/disable
  soda dataset profiling <dataset-id> --enable
  soda dataset profiling <dataset-id> --disable

  # dataset profiling — schedule
  soda dataset profiling <dataset-id> --enable --schedule "0 6 * * *"
  soda dataset profiling <dataset-id> --enable --schedule "0 6 * * *" --timezone "America/New_York"
  soda dataset profiling <dataset-id> --enable --schedule "0 */6 * * *" --timezone "UTC"

  # dataset profiling — sampling
  soda dataset profiling <dataset-id> --enable --sampling-rows 1000
  soda dataset profiling <dataset-id> --enable --schedule "0 6 * * *" --timezone "UTC" --sampling-rows 5000

  # dataset profiling refresh (expect: not yet available)
  soda dataset profiling refresh <dataset-id>

  # dataset diagnostics — view (no flags)
  soda dataset diagnostics <dataset-id>
  soda dataset diagnostics <dataset-id> --output json

  # dataset diagnostics — collect toggles
  soda dataset diagnostics <dataset-id> --collect-results
  soda dataset diagnostics <dataset-id> --no-collect-results
  soda dataset diagnostics <dataset-id> --collect-failed-rows
  soda dataset diagnostics <dataset-id> --no-collect-failed-rows

  # dataset diagnostics — schema/table naming (may not be in API yet)
  soda dataset diagnostics <dataset-id> --schema custom_schema
  soda dataset diagnostics <dataset-id> --table-prefix "soda_"
  soda dataset diagnostics <dataset-id> --table-suffix "_diag"

  # dataset diagnostics — failed rows options (may not be in API yet)
  soda dataset diagnostics <dataset-id> --failed-rows-description "Context"
  soda dataset diagnostics <dataset-id> --expose-failed-rows-query
  soda dataset diagnostics <dataset-id> --no-expose-failed-rows-query
  soda dataset diagnostics <dataset-id> --failed-rows-cta
  soda dataset diagnostics <dataset-id> --no-failed-rows-cta

  # dataset diagnostics — combo
  soda dataset diagnostics <dataset-id> --collect-results --collect-failed-rows

  # dataset permissions
  soda dataset permissions list <dataset-id>
  soda dataset permissions list <dataset-id> --output json
  soda dataset permissions assign <dataset-id> --role <role-id> --user <user-id>
  soda dataset permissions assign <dataset-id> --role <role-id> --group <group-id>
  soda dataset permissions revoke <dataset-id> --role <role-id> --user <user-id>
  soda dataset permissions revoke <dataset-id> --role <role-id> --group <group-id>
  soda dataset permissions assign <dataset-id> --role <role-id>            # expect: error (no user or group)

  # dataset onboard — interactive
  soda dataset onboard <dataset-id>

  # dataset onboard — non-interactive combos
  soda dataset onboard <dataset-id> --monitoring --profiling --contracts none
  soda dataset onboard <dataset-id> --no-monitoring --no-profiling --contracts none
  soda dataset onboard <dataset-id> --monitoring --no-profiling --contracts skeleton
  soda dataset onboard <dataset-id> --no-monitoring --profiling --contracts ai
  soda dataset onboard <dataset-id> --monitoring --profiling --contracts ai

  # ═══════════════════════════════════════════════════════════════════
  # CONTRACT
  # ═══════════════════════════════════════════════════════════════════

  # contract list
  soda contract list
  soda contract list --output json
  soda contract list --output csv

  # contract pull
  soda contract pull <datasource/db/schema/table>
  soda contract pull bad/qualified/name                                    # expect: not found

  # contract push
  soda contract push <file>.yml
  soda contract push                                                       # auto-discover single .yml

  # contract diff
  soda contract diff
  soda contract diff <file>.yml
  soda contract diff <file>.yml --dataset <override-qualified-name>

  # contract lint
  soda contract lint
  soda contract lint <file>.yml
  soda contract validate <file>.yml                                        # alias

  # contract create interactive (no --dataset, triggers wizard)
  soda contract create

  # contract create — skeleton
  soda contract create --dataset <ds/db/schema/table> --mode skeleton
  soda contract create --dataset <ds/db/schema/table> --mode skeleton --output custom.yml
  soda contract create --dataset <ds/db/schema/table>                      # default mode=skeleton

  # contract create — copilot
  soda contract create --dataset <ds/db/schema/table> --mode copilot
  soda contract create --dataset <ds/db/schema/table> --mode copilot --no-wait
  soda contract create --dataset <ds/db/schema/table> --mode copilot --output custom.yml

  # contract create — errors
  soda contract create --no-interactive                                    # expect: error (no --dataset)
  soda contract create --mode badmode --dataset <ds/db/schema/table>       # expect: error

  # contract verify
  soda contract verify <file>.yml
  soda contract verify <file>.yml --output json
  soda contract verify <file>.yml --no-wait
  soda contract verify <file>.yml --push
  soda contract verify <file>.yml --runner
  soda contract verify <file>.yml --datasource <override-config>
  soda contract verify <file>.yml --set key1=value1
  soda contract verify <file>.yml --set key1=value1 --set key2=value2
  soda contract verify <file>.yml --no-wait --output json
  soda contract verify                                                                  # expect: error (no file)
  soda contract verify nonexistent.yml                                                  # expect: error
  soda contract verify <file>.yml --quiet                                               # suppress non-essential output
  soda contract verify <file> --runner soda-core                                        # local mode, needs soda-core on PATH
  soda contract verify <file> --runner soda-core --push                                 # local execution + push results to cloud
  soda contract lint <file>                                                             # real validation via soda-core, basic YAML fallback if not installed
  
  # contract copilot — interactive
  soda contract copilot

  # contract copilot — generate
  soda contract copilot --dataset <ds/db/schema/table>
  soda contract copilot --dataset <ds/db/schema/table> --output custom.yml

  # contract copilot — improve
  soda contract copilot <file>.yml "add freshness checks"
  soda contract copilot <file>.yml "tighten null constraints on order_id"
  soda contract copilot <file>.yml                                         # interactive: asks for prompt
  soda contract copilot --no-interactive                                   # expect: error

  # contract proposal (all expect: not yet available)
  soda contract proposal list
  soda contract proposal list --status open
  soda contract proposal list --status done
  soda contract proposal list --status all
  soda contract proposal pull <id>
  soda contract proposal pull <id> --revision 2
  soda contract proposal push <id>
  soda contract proposal push <id> <file>.yml
  soda contract proposal push <id> <file>.yml --message "Updated checks"
  soda contract proposal close <id>
  soda contract proposal close <id> --status done
  soda contract proposal close <id> --status wontdo

  # ═══════════════════════════════════════════════════════════════════
  # MONITOR
  # ═══════════════════════════════════════════════════════════════════

  # monitor list
  soda monitor list --dataset <dataset-id>
  soda monitor list --dataset <dataset-id> --output json
  soda monitor list --dataset <dataset-id> --type column
  soda monitor list --dataset <dataset-id> --type custom
  soda monitor list --dataset <dataset-id> --type dataset
  soda monitor list                                                        # expect: error (--dataset required)

  # monitor config — view
  soda monitor config <dataset-id>
  soda monitor config <dataset-id> --output json

  # monitor config — enable/disable
  soda monitor config <dataset-id> --enable
  soda monitor config <dataset-id> --disable

  # monitor config — schedule
  soda monitor config <dataset-id> --enable --schedule "0 6 * * *"
  soda monitor config <dataset-id> --enable --schedule "0 */6 * * *" --timezone "UTC"
  soda monitor config <dataset-id> --enable --schedule "0 8 * * MON-FRI" --timezone "America/New_York"

  # monitor add — column type (all metric values)
  soda monitor add --dataset <id> --type column --column <col> --metric count
  soda monitor add --dataset <id> --type column --column <col> --metric missing-pct
  soda monitor add --dataset <id> --type column --column <col> --metric duplicate-pct
  soda monitor add --dataset <id> --type column --column <col> --metric distinct-count
  soda monitor add --dataset <id> --type column --column <col> --metric min
  soda monitor add --dataset <id> --type column --column <col> --metric max
  soda monitor add --dataset <id> --type column --column <col> --metric avg
  soda monitor add --dataset <id> --type column --column <col> --metric sum
  soda monitor add --dataset <id> --type column --column <col> --metric std-dev
  soda monitor add --dataset <id> --type column --column <col> --metric variance
  soda monitor add --dataset <id> --type column --column <col> --metric q1
  soda monitor add --dataset <id> --type column --column <col> --metric median
  soda monitor add --dataset <id> --type column --column <col> --metric q3
  soda monitor add --dataset <id> --type column --column <col> --metric min-length
  soda monitor add --dataset <id> --type column --column <col> --metric max-length
  soda monitor add --dataset <id> --type column --column <col> --metric avg-length
  soda monitor add --dataset <id> --type column --column <col> --metric freshness

  # monitor add — column with group-by
  soda monitor add --dataset <id> --type column --column <col> --metric count --group-by <col2>
  soda monitor add --dataset <id> --type column --column <col> --metric avg --group-by <col2> --group-by <col3>

  # monitor add — custom type
  soda monitor add --dataset <id> --type custom --name "row count check" --sql "SELECT count(*) as cnt FROM table" --result-metric cnt
  soda monitor add --dataset <id> --type custom --name "custom" --sql-file query.sql --result-metric cnt
  soda monitor add --dataset <id> --type custom --name "custom" --sql "SELECT 1" --result-metric cnt --column <col>

  # monitor add — dataset type (expect: error — no write endpoint)
  soda monitor add --dataset <id> --type dataset --metric row-count
  soda monitor add --dataset <id> --type dataset --metric freshness
  soda monitor add --dataset <id> --type dataset --metric schema
  soda monitor add --dataset <id> --type dataset --metric rows-inserted
  soda monitor add --dataset <id> --type dataset --metric row-count-change
  soda monitor add --dataset <id> --type dataset --metric timeliness

  # monitor add — error cases
  soda monitor add --dataset <id> --type column                            # expect: error (no --column)
  soda monitor add --dataset <id> --type custom                            # expect: error (no --name/--sql)
  soda monitor add --type column --column x --metric count                 # expect: error (no --dataset)

  # monitor update
  soda monitor update <monitor-id> --dataset <id> --enable
  soda monitor update <monitor-id> --dataset <id> --disable
  soda monitor update <monitor-id> --dataset <id> --sql "SELECT 1"
  soda monitor update <monitor-id> --dataset <id> --name "renamed"
  soda monitor update <monitor-id> --dataset <id> --result-metric new_col
  soda monitor update <monitor-id> --dataset <id> --enable --name "renamed" --sql "SELECT 1" --result-metric cnt

  # monitor delete
  soda monitor delete <monitor-id> --dataset <id>
  soda monitor delete bad-id --dataset <id>                                # expect: error

  # ═══════════════════════════════════════════════════════════════════
  # RESULTS
  # ═══════════════════════════════════════════════════════════════════

  # results list — no filters
  soda results list
  soda results list --output json
  soda results list --output csv

  # results list — individual filters
  soda results list --dataset <dataset-id>
  soda results list --dataset-name "orders"
  soda results list --dataset-name "public"
  soda results list --status passing
  soda results list --status failing
  soda results list --status error
  soda results list --type check
  soda results list --type monitor                                         # expect: graceful "not available"
  soda results list --type all
  soda results list --limit 5
  soda results list --limit 50

  # results list — sorting
  soda results list --sort dataset
  soda results list --sort name
  soda results list --sort column
  soda results list --sort status
  soda results list --sort date
  soda results list --sort date --order asc
  soda results list --sort date --order desc
  soda results list --sort name --order asc

  # results list — date range
  soda results list --from 2026-03-01
  soda results list --until 2026-03-18
  soda results list --from 2026-03-01 --until 2026-03-18

  # results list — combined filters
  soda results list --dataset <id> --status failing --limit 20 --sort date --order desc
  soda results list --dataset-name "orders" --status passing --from 2026-03-01 --sort name --order asc

  # ═══════════════════════════════════════════════════════════════════
  # JOB / SCAN
  # ═══════════════════════════════════════════════════════════════════

  # job list (mock data)
  soda job list
  soda job list --output json
  soda job list --datasource <id>
  soda job list --dataset <id>
  soda job list --type contract
  soda job list --type monitor
  soda job list --type all
  soda job list --status passing
  soda job list --status failing
  soda job list --status running
  soda job list --status error
  soda job list --datasource <id> --type contract --status failing
  soda scan list                                                           # alias

  # job logs
  soda job logs <scan-id>
  soda job logs <scan-id> --output json
  soda job logs <scan-id> --follow
  soda job logs bad-id                                                     # expect: error
  soda scan logs <scan-id>                                                 # alias

  # job cancel (API returns 404 — blocked)
  soda job cancel <scan-id>
  soda job cancel bad-id                                                   # expect: error
  soda scan cancel <scan-id>                                               # alias

  # ═══════════════════════════════════════════════════════════════════
  # INCIDENT (all expect: "not yet available" — blocked)
  # ═══════════════════════════════════════════════════════════════════

  # incident list
  soda incident list
  soda incident list --output json
  soda incident list --status reported
  soda incident list --status investigating
  soda incident list --status fixing
  soda incident list --status resolved
  soda incident list --dataset <dataset-id>
  soda incident list --status reported --dataset <dataset-id>

  # incident get
  soda incident get <id>
  soda incident get <id> --output json
  soda incident get bad-id                                                 # expect: error

  # incident update
  soda incident update <id> --title "New title"
  soda incident update <id> --severity minor
  soda incident update <id> --severity major
  soda incident update <id> --severity critical
  soda incident update <id> --status reported
  soda incident update <id> --status investigating
  soda incident update <id> --status fixing
  soda incident update <id> --status resolved
  soda incident update <id> --description "Detailed description"
  soda incident update <id> --assigned-to user@example.com
  soda incident update <id> --title "Updated" --severity critical --status investigating --description "Desc" --assigned-to user@example.com

  # ═══════════════════════════════════════════════════════════════════
  # IAM — ROLES
  # ═══════════════════════════════════════════════════════════════════

  # iam role list
  soda iam role list
  soda iam role list --output json
  soda iam role list --scope global
  soda iam role list --scope dataset

  # iam role create (expect: not yet available)
  soda iam role create --name "test-role" --scope dataset
  soda iam role create --name "test-role" --scope global
  soda iam role create --name "test-role" --scope dataset --description "A test role"
  soda iam role create --name "test-role" --scope dataset --permission create-api-keys
  soda iam role create --name "test-role" --scope dataset --permission create-api-keys --permission create-datasets
  soda iam role create --name "test-role" --scope dataset --permission manage-attributes
  soda iam role create --name "test-role" --scope dataset --permission manage-datasources
  soda iam role create --name "test-role" --scope dataset --permission manage-notification-rules
  soda iam role create --name "test-role" --scope dataset --permission manage-org-settings
  soda iam role create --name "test-role" --scope dataset --permission manage-scan-definitions

  # iam role show / delete (expect: not yet available)
  soda iam role show <role-id>
  soda iam role delete <role-id>

  # ═══════════════════════════════════════════════════════════════════
  # IAM — USERS
  # ═══════════════════════════════════════════════════════════════════

  # iam user list
  soda iam user list
  soda iam user list --output json

  # iam user invite (expect: not yet available)
  soda iam user invite --email newuser@example.com

  # iam user remove (expect: not yet available)
  soda iam user remove <user-id>

  # iam user assign/revoke (expect: not yet available)
  soda iam user assign <user-id> --role <role-id>
  soda iam user revoke <user-id> --role <role-id>

  # ═══════════════════════════════════════════════════════════════════
  # IAM — GROUPS
  # ═══════════════════════════════════════════════════════════════════

  # iam group list
  soda iam group list
  soda iam group list --output json

  # iam group create
  soda iam group create --name "test-group"
  soda iam group create --name "test-group-2" --member user1@example.com
  soda iam group create --name "test-group-3" --member user1@example.com --member user2@example.com

  # iam group update
  soda iam group update <group-id> --name "renamed-group"
  soda iam group update <group-id> --add-member new@example.com
  soda iam group update <group-id> --add-member a@example.com --add-member b@example.com
  soda iam group update <group-id> --remove-member old@example.com
  soda iam group update <group-id> --remove-member a@example.com --remove-member b@example.com
  soda iam group update <group-id> --name "new" --add-member a@example.com --remove-member b@example.com

  # iam group delete
  soda iam group delete <group-id>                                         # ⚠ use one we created
  soda iam group delete bad-id                                             # expect: error

  # iam group assign/revoke (expect: not yet available)
  soda iam group assign <group-id> --role <role-id>
  soda iam group revoke <group-id> --role <role-id>

  # ═══════════════════════════════════════════════════════════════════
  # IAM — SERVICE ACCOUNTS (all expect: not yet available)
  # ═══════════════════════════════════════════════════════════════════

  soda iam service-account list
  soda iam service-account create --name "ci-bot" --email ci@example.com
  soda iam service-account delete <id>

  # ═══════════════════════════════════════════════════════════════════
  # RUNNER
  # ═══════════════════════════════════════════════════════════════════

  # runner list
  soda runner list
  soda runner list --output json
  soda runner list --output csv

  # runner get
  soda runner get <runner-id>
  soda runner get <runner-id> --output json
  soda runner get bad-id                                                   # expect: error

  # runner create
  soda runner create --name "test-cli-runner"
  soda runner create                                                       # expect: error (missing --name)

  # runner delete
  soda runner delete <runner-id>                                           # ⚠ use one we created
  soda runner delete bad-id                                                # expect: error

  # ═══════════════════════════════════════════════════════════════════
  # NOTIFICATION — RULES (all expect: not yet available)
  # ═══════════════════════════════════════════════════════════════════

  soda notification rule list
  soda notification rule list --output json

  # notification rule add — minimal
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com

  # notification rule add — all sources
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com
  soda notification rule add --name "test" --source monitor --alert anomaly --notify me@example.com
  soda notification rule add --name "test" --source all --alert warn-fail --notify me@example.com

  # notification rule add — all alert types
  soda notification rule add --name "test" --source check --alert warn-fail --notify me@example.com
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com
  soda notification rule add --name "test" --source check --alert anomaly --notify me@example.com

  # notification rule add — scope filters
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --datasource "prod"
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --dataset "orders"
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --dataset-owner owner@example.com
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --dataset-tag production

  # notification rule add — check-only filters
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --check-name "row_count"
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --check-name "contains:null"
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --check-owner owner@example.com

  # notification rule add — monitor-only filters
  soda notification rule add --name "test" --source monitor --alert anomaly --notify me@example.com --monitor-type column

  # notification rule add — multiple recipients
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --notify team@example.com

  # notification rule add — options
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --granular-results
  soda notification rule add --name "test" --source check --alert fail-only --notify me@example.com --message "Custom alert message"

  # notification rule add — kitchen sink
  soda notification rule add --name "full" --source check --alert warn-fail --notify me@example.com --notify team@example.com --datasource "prod" --dataset "orders" --dataset-owner
  owner@example.com --dataset-tag critical --check-name "contains:null" --check-owner dev@example.com --granular-results --message "Alert"

  # notification rule update
  soda notification rule update <id> --name "renamed"
  soda notification rule update <id> --source monitor
  soda notification rule update <id> --alert anomaly
  soda notification rule update <id> --notify new@example.com

  # notification rule delete
  soda notification rule delete <id>

  # ═══════════════════════════════════════════════════════════════════
  # NOTIFICATION — INTEGRATIONS (all expect: not yet available)
  # ═══════════════════════════════════════════════════════════════════

  soda notification integration list
  soda notification integration list --output json
  soda notification integration add slack
  soda notification integration add teams
  soda notification integration add webhook
  soda notification integration test <id>
  soda notification integration delete <id>

  # ═══════════════════════════════════════════════════════════════════
  # SECRET (all expect: not yet available)
  # ═══════════════════════════════════════════════════════════════════

  soda secret create --name "MY_SECRET" --value "abc123"
  soda secret create --name "DB_PASSWORD" --value "s3cret"
  soda secret update <id> --value "new-value"
  soda secret delete <id>

  # ═══════════════════════════════════════════════════════════════════
  # DASHBOARD
  # ═══════════════════════════════════════════════════════════════════

  soda dashboard
  soda dashboard --output json

  # ═══════════════════════════════════════════════════════════════════
  # UTILITY
  # ═══════════════════════════════════════════════════════════════════

  soda init
  soda version
  soda completion bash
  soda completion zsh
  soda completion fish
  soda --help
  soda --version

  # ═══════════════════════════════════════════════════════════════════
  # GLOBAL FLAG COMBOS (on any working command)
  # ═══════════════════════════════════════════════════════════════════

  soda datasource list -o json
  soda datasource list -o csv
  soda datasource list -o table
  soda datasource list --quiet
  soda datasource list -q
  soda datasource list --verbose
  soda datasource list -v
  soda datasource list --no-color
  soda datasource list --no-interactive
  soda datasource list --profile default
  soda datasource list --profile nonexistent                               # expect: error
  soda datasource list -o json --quiet                                     # quiet + json
  soda datasource list --no-color --no-interactive --output json

  # ═══════════════════════════════════════════════════════════════════
  # PIPED OUTPUT (verify auto-detection)
  # ═══════════════════════════════════════════════════════════════════

  soda datasource list | cat                                               # should auto-switch to json
  soda dataset list | head -5                                              # should auto-switch to json
  soda results list | jq .                                                 # parse as json
  soda runner list | cat                                                   # should auto-switch to json

  soda can

  Total: ~320 test commands covering every command, every flag, and every valid flag value in the CLI.