# Responses web search usage details

`usage_logs` remains the only request/token/cost ledger. `usage_web_search_events`
contains display-only children, inserted in the same transaction as their parent.
Retries of an existing `(request_id, api_key_id)` do not append or replace children.
Deleting a parent cascades to its children. Migration:
`migrations/241_add_usage_web_search_events.sql`.

## Scope and retained data

- New OpenAI/Codex Responses `web_search_call` output only; historical records are
  not backfilled. Standalone alpha/search and Grok search billing remain unchanged.
- JSON, SSE, buffered SSE-to-JSON, and Responses WebSocket paths share the bounded
  collector. Item completion and final response snapshots are deduplicated by IDs
  or output position; final snapshots can supply missing source metadata.
- At most 32 events per request and 20 unique valid sources per event are retained.
  `source_count` is the number of retained sources, not an unbounded upstream count.
- Query: 1,024 Unicode characters; call ID: 128; status: 64; source title: 512.
  Source URLs must be absolute HTTP(S), have no embedded credentials, and fit in
  2,048 bytes. Oversized URLs are discarded rather than truncated.
- Only query/status/identifiers and source URL/title are stored, not response text,
  snippets, tool arguments, or raw upstream payloads. Queries can contain user
  search terms; detailed events are exposed only by the admin DTO.
- If upstream omits query/sources, available fields are kept. Requests are not
  modified to force source inclusion; callers can request
  `include: ["web_search_call.action.sources"]` when supported by their upstream.

## Admin UI and API

`GET /api/v1/admin/usage` adds optional `web_search_events` on each parent record.
One batched child query loads only the selected parent page. Statistics endpoints,
model/endpoint aggregation, pagination totals and billing still use parents only.

The admin usage table starts collapsed. Its search button expands indented event
rows in the same table. Each event independently expands its sources. Source text
is escaped and only HTTP(S) links are clickable, with `noopener noreferrer`.
Mobile cards use the same details. The shared table suspends virtual row
measurement while detail rows are expanded.

## Verification

Focused backend coverage (parser limits, malformed data, deduplication, forwarding,
WS bridge isolation, billing invariants, transactional writes, API contract):

```sh
cd backend
GOSUMDB=sum.golang.org go test -race -tags unit \
  ./internal/service ./internal/handler ./internal/handler/admin \
  ./internal/handler/dto ./internal/repository ./migrations \
  -run 'WebSearchEvents|TestCountGrokNativeSearch|TestOpenAIGatewayServiceRecordUsage_' -count=1
```

`GOSUMDB` is a process-only override needed by this checkout's local environment
for its Go 1.27 toolchain; no module files or global Go settings were changed.
Frontend regression tests cover `UsageTable.webSearch.spec.ts`, existing usage
and DataTable suites, and locales; `pnpm build` also verifies the production build.

PostgreSQL integration tests are in
`internal/repository/usage_web_search_events_integration_test.go` and
`internal/repository/usage_web_search_events_sql_integration_test.go`. They cover
Ent and database/sql transactions, page totals, duplicate/concurrent retries,
parent-only aggregate statistics, cascade deletion, orphan rejection and rollback.
The runner requires Docker and cleans only containers belonging to its own
Testcontainers session, including failed runs:

```sh
# From the repository root, using locally cached compatibility images:
GOSUMDB=sum.golang.org \
SUB2API_TEST_POSTGRES_IMAGE=postgres:18-alpine \
SUB2API_TEST_REDIS_IMAGE=redis:7-alpine \
TESTCONTAINERS_RYUK_DISABLED=true \
bash backend/scripts/test-usage-web-search-events.sh
```

The harness defaults are unchanged; the two image overrides are explicit opt-ins.
Normally Ryuk stays enabled. This offline/cached-image example disables it only
for this process and uses the runner's session-scoped cleanup instead; do not
remove the cleanup when running without Ryuk.

### Verification status (2026-10-03)

- PostgreSQL 16 and 18, with Redis 7: feature integration tests passed with
  `-race`, including the production SQL transaction paths and concurrent retries.
- PostgreSQL 18: the API-key token-quota projection, billing-cache suite and
  subscription repository suite passed with `-race`.
- Full repository, DTO and migration unit suites passed with `-race`. Local HTTP
  mock tests require clearing inherited proxies for the test process:

  ```sh
  cd backend
  env -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY \
      -u http_proxy -u https_proxy -u all_proxy GOSUMDB=sum.golang.org \
      go test -race -tags unit ./internal/repository ./internal/handler/dto ./migrations -count=1
  ```

- Stale integration call signatures and the reasoning-effort argument-index test
  were corrected without changing production billing. The group test fixture now
  persists its token-quota fields instead of silently omitting them.
- The first attempt stalled pulling the Ryuk image. Cached-image runs succeeded;
  their containers, and the first attempt's unstarted container, were removed by
  verified session labels. Existing deployment containers/data were not touched.
- No global proxy/Go settings were changed. No migration was applied to a deployed
  database. Live Codex acceptance still needs an explicitly selected test instance
  and authorized account; no paid/live upstream request has been sent.
