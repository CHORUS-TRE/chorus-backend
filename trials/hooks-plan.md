# Webhooks Specification

## 1. Overview

Chorus needs a webhook system allowing users and platform administrators to subscribe to entity lifecycle events. When a subscribed event occurs, Chorus sends an HTTP POST request to a user-configured endpoint with a structured payload describing the change.

The payload format follows the **CloudEvents v1.0** specification (CNCF graduated project), which is the industry standard for describing event data. GitHub, Azure Event Grid, Google Eventarc, and many other platforms use this format.

---

## 2. Concerned Entities and Events

### 2.1 Workspace

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.workspace.created` | A new workspace is created |
| `ch.chuv.chorus.workspace.updated` | Workspace name, description, or status changes |
| `ch.chuv.chorus.workspace.deleted` | Workspace is soft-deleted |
| `ch.chuv.chorus.workspace.member.added` | A user is added to a workspace (any role) |
| `ch.chuv.chorus.workspace.member.removed` | A user is removed from a workspace |
| `ch.chuv.chorus.workspace.member.role_changed` | A user's workspace role changes |

### 2.2 Workbench

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.workbench.created` | A workbench is created in a workspace |
| `ch.chuv.chorus.workbench.deleted` | A workbench is soft-deleted |
| `ch.chuv.chorus.workbench.status_changed` | Pod status changes (Ready, Failing, Terminated, etc.) |

### 2.3 App Instance

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.app_instance.created` | An app instance is deployed in a workbench |
| `ch.chuv.chorus.app_instance.deleted` | An app instance is removed |
| `ch.chuv.chorus.app_instance.status_changed` | K8s state changes (Running, Stopped, Killed) |

### 2.4 App (Store)

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.app.created` | A new app is added to the store |
| `ch.chuv.chorus.app.updated` | An app definition is updated |
| `ch.chuv.chorus.app.deleted` | An app is removed from the store |

### 2.5 User

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.user.created` | A new user account is created |
| `ch.chuv.chorus.user.updated` | User profile is updated |
| `ch.chuv.chorus.user.deleted` | A user account is deleted |
| `ch.chuv.chorus.user.role_changed` | Platform-level role assignment changes |

### 2.6 Approval Request

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.approval_request.created` | A new data extraction/transfer request is submitted |
| `ch.chuv.chorus.approval_request.approved` | A request is approved |
| `ch.chuv.chorus.approval_request.rejected` | A request is rejected |
| `ch.chuv.chorus.approval_request.cancelled` | A request is cancelled by its requester |

### 2.7 Workspace File

| Event Type | Trigger |
|---|---|
| `ch.chuv.chorus.workspace_file.uploaded` | A file is uploaded to a workspace |
| `ch.chuv.chorus.workspace_file.deleted` | A file is deleted from a workspace |

---

## 3. Scope Levels

Webhooks can be registered at two distinct scopes. A single event may fire hooks at multiple scopes.

### 3.1 Platform-level hooks

- **Who can manage:** `SuperAdmin`, `PlatformSettingsManager`.
- **Fires on:** Every event of the subscribed type across the entire platform (all tenants or within the admin's tenant, depending on role).
- **Use cases:**
  - External audit trail ingestion.
  - Organization-wide SIEM integration.
  - Triggering provisioning pipelines on any workspace creation.
  - Billing / usage tracking systems.

### 3.2 Workspace-level hooks

- **Who can manage:** `WorkspaceAdmin`, `WorkspaceMaintainer`.
- **Fires on:** Events scoped to the workspace the hook is registered in (workbench created in *this* workspace, member added to *this* workspace, file uploaded to *this* workspace, approval request within *this* workspace, etc.).
- **Use cases:**
  - Notifying an external project management tool when members join.
  - CI/CD pipelines triggered on workbench or app instance status changes.
  - Slack/Teams notifications for a specific workspace's activity.

### 3.3 Scope and event matrix

| Event | Platform-level | Workspace-level |
|---|---|---|
| `workspace.created` | Yes | N/A (workspace doesn't exist yet) |
| `workspace.updated` | Yes | Yes |
| `workspace.deleted` | Yes | Yes |
| `workspace.member.*` | Yes | Yes |
| `workbench.*` | Yes | Yes |
| `app_instance.*` | Yes | Yes |
| `app.*` | Yes | N/A (apps are global) |
| `user.*` | Yes | N/A (users are global) |
| `approval_request.*` | Yes | Yes |
| `workspace_file.*` | Yes | Yes |

---

## 4. Payload Format — CloudEvents v1.0

All webhook deliveries use the **CloudEvents v1.0 structured-mode JSON** format over HTTP, as defined by the [CloudEvents spec](https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md) and its [HTTP protocol binding](https://github.com/cloudevents/spec/blob/main/cloudevents/bindings/http-protocol-binding.md).

A Go SDK is available: [`github.com/cloudevents/sdk-go`](https://github.com/cloudevents/sdk-go).

### 4.1 HTTP request

```
POST {hook.url} HTTP/1.1
Content-Type: application/cloudevents+json; charset=utf-8
X-Chorus-Signature: sha256=<HMAC-SHA256 hex digest of body using hook secret>
X-Chorus-Hook-ID: <hook registration ID>
X-Chorus-Delivery: <unique delivery attempt UUID>
User-Agent: Chorus-Webhooks/1.0
```

### 4.2 Example payload — workspace created

```json
{
  "specversion": "1.0",
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "type": "ch.chuv.chorus.workspace.created",
  "source": "/tenants/1/workspaces",
  "subject": "42",
  "time": "2026-04-13T14:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "workspace": {
      "id": 42,
      "tenantId": 1,
      "name": "Oncology Research Q2",
      "shortName": "onco-q2",
      "status": "active",
      "createdBy": {
        "id": 7,
        "username": "jdoe"
      },
      "createdAt": "2026-04-13T14:30:00Z"
    }
  }
}
```

### 4.3 Example payload — workspace member added

```json
{
  "specversion": "1.0",
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "type": "ch.chuv.chorus.workspace.member.added",
  "source": "/tenants/1/workspaces/42/members",
  "subject": "15",
  "time": "2026-04-13T15:10:00Z",
  "datacontenttype": "application/json",
  "data": {
    "workspaceId": 42,
    "user": {
      "id": 15,
      "username": "asmith"
    },
    "role": "WorkspaceMember",
    "addedBy": {
      "id": 7,
      "username": "jdoe"
    }
  }
}
```

### 4.4 Data payload guidelines

- The `data` field contains the relevant entity snapshot at the time of the event.
- Sensitive fields (passwords, TOTP secrets, tokens) MUST NEVER appear in payloads.
- The `subject` field is the string ID of the primary affected resource.
- The `source` field is a URI-reference path: `/tenants/{tenantId}/{resource-path}`.

---

## 5. Webhook Registration Model

### 5.1 Database schema

```sql
-- Migration: 00037_webhooks.sql

-- +migrate Up

CREATE TABLE public.webhooks (
    id            BIGINT PRIMARY KEY DEFAULT nextval('chorus_seq'),
    tenantid      BIGINT NOT NULL REFERENCES tenants(id),
    workspace_id  BIGINT REFERENCES workspaces(id),  -- NULL = platform-level hook
    name          TEXT NOT NULL,
    url           TEXT NOT NULL,                      -- Target endpoint URL (HTTPS required in prod)
    secret_enc    BYTEA NOT NULL,                     -- AES-256-GCM encrypted HMAC secret
    event_types   TEXT[] NOT NULL,                    -- Array of subscribed event type patterns
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_by    BIGINT NOT NULL REFERENCES users(id),
    createdat     TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat     TIMESTAMP NOT NULL DEFAULT NOW(),
    deletedat     TIMESTAMP,

    CONSTRAINT webhooks_scope_check CHECK (
        -- Platform-level hooks have no workspace_id
        -- Workspace-level hooks must have one
        true
    )
);

CREATE INDEX idx_webhooks_tenant ON public.webhooks(tenantid) WHERE deletedat IS NULL;
CREATE INDEX idx_webhooks_workspace ON public.webhooks(tenantid, workspace_id) WHERE deletedat IS NULL;
CREATE INDEX idx_webhooks_event_types ON public.webhooks USING GIN(event_types) WHERE deletedat IS NULL;

CREATE TABLE public.webhook_deliveries (
    id              BIGINT PRIMARY KEY DEFAULT nextval('chorus_seq'),
    webhook_id      BIGINT NOT NULL REFERENCES webhooks(id),
    event_id        TEXT NOT NULL,                    -- CloudEvent id
    event_type      TEXT NOT NULL,
    request_body    JSONB NOT NULL,                   -- Full CloudEvent payload (for retry/debug)
    response_status INT,                              -- HTTP status code from target
    response_body   TEXT,                             -- Truncated response body (max 10 KB)
    duration_ms     INT,                              -- Round-trip time
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending | success | failed | retrying
    attempt         INT NOT NULL DEFAULT 1,
    next_retry_at   TIMESTAMP,                        -- Scheduled time for next retry
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook ON public.webhook_deliveries(webhook_id, created_at DESC);
CREATE INDEX idx_webhook_deliveries_retry ON public.webhook_deliveries(next_retry_at)
    WHERE status IN ('pending', 'retrying');

-- +migrate Down
DROP TABLE IF EXISTS public.webhook_deliveries;
DROP TABLE IF EXISTS public.webhooks;
```

### 5.2 Go model

```go
type Webhook struct {
    ID           uint64
    TenantID     uint64
    WorkspaceID  *uint64   // nil = platform-level
    Name         string
    URL          string
    SecretEnc    []byte    // AES-256-GCM encrypted; never exposed via API
    EventTypes   []string  // e.g. ["ch.chuv.chorus.workspace.created", "ch.chuv.chorus.workbench.*"]
    IsActive     bool
    CreatedBy    uint64
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time
}
```

### 5.3 Event type pattern matching

Subscriptions support a trailing wildcard:

| Pattern | Matches |
|---|---|
| `ch.chuv.chorus.workspace.created` | Exact match only |
| `ch.chuv.chorus.workspace.*` | All workspace events |
| `ch.chuv.chorus.workbench.*` | All workbench events |
| `ch.chuv.chorus.*` | All events (useful for platform-level audit hooks) |

---

## 6. Secret Encryption at Rest

Webhook secrets are used to compute the `X-Chorus-Signature` HMAC for each delivery so the receiver can verify payloads are authentic. These secrets must be stored encrypted.

### 6.1 Algorithm

Follow the existing pattern in `internal/utils/crypto/crypto.go`:

- **Cipher:** AES-256-GCM (authenticated encryption with associated data).
- **Key derivation:** PBKDF2-SHA256 from the daemon's master encryption key (already present in `config.Sensitive`).
- **Nonce:** 12-byte random nonce, prepended to ciphertext (standard GCM convention).
- **Stored format:** `nonce (12 bytes) || ciphertext || GCM tag (16 bytes)` — stored as `BYTEA` in `secret_enc`.

### 6.2 Flow

```
User creates hook with plaintext secret
        │
        ▼
  ┌──────────────┐
  │  API layer   │  secret arrives in CreateWebhookRequest
  └──────┬───────┘
         │
         ▼
  ┌──────────────────────┐
  │  crypto.Encrypt()    │  AES-256-GCM with daemon key
  └──────┬───────────────┘
         │  → []byte ciphertext
         ▼
  ┌──────────────┐
  │  PostgreSQL  │  stored in webhooks.secret_enc as BYTEA
  └──────────────┘

Delivery time:
  ┌──────────────┐
  │  PostgreSQL  │  read secret_enc
  └──────┬───────┘
         │
         ▼
  ┌──────────────────────┐
  │  crypto.Decrypt()    │  AES-256-GCM with daemon key
  └──────┬───────────────┘
         │  → plaintext secret
         ▼
  ┌──────────────────────────────────────────┐
  │  HMAC-SHA256(secret, requestBody)        │
  │  → X-Chorus-Signature: sha256=<hex>      │
  └──────────────────────────────────────────┘
```

### 6.3 API behavior

- **Create/Update:** Accept `secret` in plaintext; never return it in responses.
- **Read/List:** Return `secretSet: true/false` — never the plaintext or ciphertext.
- **Rotation:** Accept a new `secret` on update; re-encrypt and overwrite.

---

## 7. Delivery Mechanics

### 7.1 Request signing

Every delivery is signed so the receiver can authenticate it:

```
signature = HMAC-SHA256(webhook_secret, raw_request_body)
Header: X-Chorus-Signature: sha256=<hex-encoded signature>
```

The receiver should compute the same HMAC and compare in constant time (same approach as GitHub webhooks).

### 7.2 Delivery expectations

| Property | Value |
|---|---|
| HTTP method | `POST` |
| Content-Type | `application/cloudevents+json; charset=utf-8` |
| Timeout | 10 seconds |
| Success | Any `2xx` response |
| TLS | HTTPS required in production; HTTP allowed in dev only |

### 7.3 Retry policy

Exponential backoff with jitter on failure (`4xx >= 400` except `410 Gone`, all `5xx`, timeouts, connection errors):

| Attempt | Delay |
|---|---|
| 1 | Immediate |
| 2 | 1 minute |
| 3 | 5 minutes |
| 4 | 30 minutes |
| 5 | 2 hours |
| 6 | 8 hours |

After 6 failed attempts, the delivery is marked `failed`. No further retries.

A `410 Gone` response automatically **deactivates** the webhook (`is_active = false`).

### 7.4 Automatic disabling

If a webhook accumulates **10 consecutive delivery failures** (across any events), it is automatically deactivated. A notification is sent to the hook creator. The hook can be re-enabled manually.

---

## 8. API Endpoints

### 8.1 Platform-level hooks

Requires `SuperAdmin` or `PlatformSettingsManager` role.

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/rest/v1/webhooks` | Create a platform-level hook |
| `GET` | `/api/rest/v1/webhooks` | List platform-level hooks |
| `GET` | `/api/rest/v1/webhooks/{id}` | Get hook details |
| `PUT` | `/api/rest/v1/webhooks/{id}` | Update hook (URL, events, active, secret) |
| `DELETE` | `/api/rest/v1/webhooks/{id}` | Soft-delete hook |
| `GET` | `/api/rest/v1/webhooks/{id}/deliveries` | List recent deliveries (paginated) |
| `POST` | `/api/rest/v1/webhooks/{id}/deliveries/{deliveryId}/retry` | Manually retry a failed delivery |
| `POST` | `/api/rest/v1/webhooks/{id}/ping` | Send a test `ch.chuv.chorus.ping` event |

### 8.2 Workspace-level hooks

Requires `WorkspaceAdmin` or `WorkspaceMaintainer` role in the workspace.

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/rest/v1/workspaces/{workspaceId}/webhooks` | Create a workspace hook |
| `GET` | `/api/rest/v1/workspaces/{workspaceId}/webhooks` | List workspace hooks |
| `GET` | `/api/rest/v1/workspaces/{workspaceId}/webhooks/{id}` | Get hook details |
| `PUT` | `/api/rest/v1/workspaces/{workspaceId}/webhooks/{id}` | Update hook |
| `DELETE` | `/api/rest/v1/workspaces/{workspaceId}/webhooks/{id}` | Soft-delete hook |
| `GET` | `/api/rest/v1/workspaces/{workspaceId}/webhooks/{id}/deliveries` | List recent deliveries |
| `POST` | `/api/rest/v1/workspaces/{workspaceId}/webhooks/{id}/ping` | Send a test ping event |

---

## 9. Protobuf Definitions (summary)

```protobuf
// webhook.proto

message Webhook {
  uint64 id = 1;
  optional uint64 workspace_id = 2;
  string name = 3;
  string url = 4;
  repeated string event_types = 5;
  bool is_active = 6;
  bool secret_set = 7;             // true if a secret is configured; secret value never returned
  uint64 created_by = 8;
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
}

message CreateWebhookRequest {
  optional uint64 workspace_id = 1;
  string name = 2;
  string url = 3;
  string secret = 4;              // plaintext; encrypted before storage
  repeated string event_types = 5;
}

message UpdateWebhookRequest {
  uint64 id = 1;
  optional string name = 2;
  optional string url = 3;
  optional string secret = 4;     // if set, rotates the secret
  repeated string event_types = 5;
  optional bool is_active = 6;
}

message WebhookDelivery {
  uint64 id = 1;
  string event_id = 2;
  string event_type = 3;
  int32 response_status = 4;
  string status = 5;              // pending | success | failed | retrying
  int32 attempt = 6;
  google.protobuf.Timestamp created_at = 7;
}
```

---

## 10. Internal Architecture

### 10.1 Event emitter

Add an internal `EventBus` (in-process) that services call after completing mutations:

```go
// pkg/webhook/event_bus.go

type Event struct {
    TenantID    uint64
    WorkspaceID *uint64   // nil for global events
    Type        string    // e.g. "ch.chuv.chorus.workspace.created"
    Subject     string    // resource ID
    Data        any       // entity snapshot
    Actor       UserRef   // who triggered this
    Time        time.Time
}

type EventBus interface {
    Publish(ctx context.Context, event Event)
    Subscribe(handler EventHandler)
}
```

The webhook dispatcher subscribes to the bus, matches events against registered hooks, and enqueues deliveries.

### 10.2 Delivery worker

Use the existing `internal/job` framework (PostgreSQL-backed job queue with `lock_store`) to process deliveries asynchronously:

1. On event publish → insert rows into `webhook_deliveries` for each matching hook.
2. A background job polls `webhook_deliveries` for `pending` / `retrying` rows where `next_retry_at <= NOW()`.
3. The worker executes the HTTP POST, records the response, and updates status.
4. On failure, computes `next_retry_at` with exponential backoff and increments `attempt`.

This approach avoids adding new infrastructure (no Redis, no message broker) and reuses the existing job/lock pattern.

### 10.3 Integration points

Services that need to emit events add a call after the mutation succeeds:

```go
// In workspace service, after creating a workspace:
w.eventBus.Publish(ctx, webhook.Event{
    TenantID:    req.TenantID,
    Type:        "ch.chuv.chorus.workspace.created",
    Subject:     strconv.FormatUint(workspace.ID, 10),
    Data:        workspace,
    Actor:       UserRef{ID: req.UserID, Username: req.Username},
    Time:        workspace.CreatedAt,
})
```

This is a thin integration — each service only adds a `Publish` call; all delivery logic is in the webhook package.

---

## 11. Security Considerations

| Concern | Mitigation |
|---|---|
| Secret storage | AES-256-GCM encryption at rest (§6). Never logged, never returned via API. |
| Payload authenticity | HMAC-SHA256 signature in `X-Chorus-Signature` header (§7.1). |
| SSRF (server-side request forgery) | Validate hook URLs: reject private/loopback IPs, require HTTPS in production, deny `file://` / non-HTTP schemes. Optionally maintain an allowlist of permitted CIDR ranges. |
| Secrets in payloads | Never include passwords, TOTP secrets, tokens, or encryption keys in event `data`. |
| Denial of service | Rate-limit outbound deliveries per hook (e.g., max 100/min). Timeout after 10s. |
| Multi-tenant isolation | All queries filter by `tenantid`. Workspace-scoped hooks enforce workspace membership via authorization. |
| TLS | Require valid TLS certificates on target URLs. Reject self-signed in production. |
| Replay attacks | Each delivery has a unique `X-Chorus-Delivery` UUID and a `time` field. Receivers can deduplicate. |

---

## 12. Observability

- **Audit log:** Hook creation, update, deletion, and manual retries are recorded in the existing audit system.
- **Metrics:** Expose Prometheus counters/histograms:
  - `chorus_webhook_deliveries_total{event_type, status}` — delivery outcomes.
  - `chorus_webhook_delivery_duration_seconds{event_type}` — latency histogram.
  - `chorus_webhook_active_count{scope}` — gauge of active hooks.
- **Delivery log:** The `webhook_deliveries` table serves as a queryable delivery log accessible via API (§8). Entries are retained for 30 days, then pruned by a scheduled job.

---

## 13. Future Considerations

- **Filtering within event types:** Allow hooks to specify JSONPath or CEL filters on the `data` payload (e.g., "only workspace.created where data.workspace.name starts with 'onco'").
- **Batch mode:** CloudEvents supports batch-mode messages; could batch multiple events in a single HTTP call for high-volume hooks.
- **Tenant-level scope:** Add a third scope between platform and workspace where a tenant admin manages hooks for all workspaces in their tenant.
- **Webhook management UI:** Frontend page for managing hooks, viewing delivery history, and resending failed deliveries.
- **Event replay:** Allow replaying all events from a time range to a specific hook (useful for initial sync or recovery).
