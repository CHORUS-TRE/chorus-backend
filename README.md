# Chorus backend

## Introduction
This project is the backend of the chorus platform.

## Features
- **Proto Interface Description**: Implements Protocol Buffers for defining data structures and service interfaces.
- **OpenAPI API Definition**: Utilizes OpenAPI specifications for a clear and standardized RESTful API description.
- **CI/CD Pipeline**: Integrates Continuous Integration and Continuous Deployment using Jenkins on Kubernetes.
- **Versioned & Iterative Migrations**: Manages database schema changes efficiently and safely.
- **Unit Testing**: Incorporates extensive unit tests to ensure code quality and reliability.
- **Acceptance Testing**: Implements acceptance tests to validate functionality against business requirements.
- **User & Authentication Service**: Provides a dedicated service for user management and authentication.
- **JWT Generation**: Implements JSON Web Tokens for secure data transmission.
- **Authorization Middlewares**: Ensures secure access control within the application.

## Requirements

### Kubernetes

A running kubernetes cluster reachable from your machine — the backend defaults to `$HOME/.kube/config` to bootstrap its k8s client, so any cluster your existing kubeconfig already points to (kind, minikube, a real dev cluster, etc.) works out of the box.

The kubeconfig path (`clients.kubernetes.kube_config`) can be overridden like any other setting (see [Advanced Configuration](#advanced-configuration)). Alternatively, skip the kubeconfig file entirely and set these `clients.kubernetes` fields individually:

- `api_server` — the service account API server URL
- `sa_secret_path` — path to the service account secret
- `sa_override_ca` — CA certificate content, optional, for private clusters with custom CAs
- `token` — the service account token
- `ca` — the service account CA

A `backend` namespace must also exist in the cluster, containing a `kubernetes.io/dockerconfigjson` Secret named `regcred` (the default for `clients.kubernetes.image_pull_secret_name`) — used to pull images from a private registry.

The [workbench-operator](https://github.com/CHORUS-TRE/workbench-operator) must also be deployed on the cluster — it reconciles the workbench/workspace custom resources the backend creates.

### Object Storage

MinIO — run `make deps` to start it locally.

### Postgresql

Postgres — run `make deps` to start it locally.

## Quick Start

1. Install [go](https://go.dev/doc/install)
1. Download this repository
1. Make sure the [requirements](#requirements) are met — kubeconfig in place, `make deps` running
1. Build the CLI
    ```bash
    make build
    ```
    Every config command below is `./bin/chorus <command>` from here on — alias it once per shell session if you'd rather type `chorus`:
    ```bash
    alias chorus=./bin/chorus
    ```
1. Configure your environment
    ```bash
    chorus init-config
    ```
    Creates `configs/config.yaml` with just the fields `check-config` actually requires, generating what it can (private key, JWT secret, salt, steward password, JWKS, and datastore/object-store credentials matching `make deps`). The image registry and Harbor URL need a real external system, so they're left as an explicit `CHANGEME` placeholder — since both clients are enabled by default, fill those in (or disable the clients) before launching, or `chorus start` won't boot.

    Prefer to see every overridable field yourself and hand-edit `configs/config.yaml` instead? See [Advanced Configuration](#advanced-configuration).
1. Launch the backend
    ```bash
    make run
    ```
1. Browse to [localhost:5000/doc](localhost:5000/doc)
1. Import the [swagger documentation](api/openapiv2/v1-tags/apis.swagger.yaml) into a new [Postman collection](https://learning.postman.com/docs/getting-started/importing-and-exporting/importing-from-swagger)
1. Authenticate with the default user to retrieve a bearer token
1. Save the bearer token in a collection-wide variable named ```token```
1. [Configure the collection-wide authorization](https://learning.postman.com/docs/use/use-collections/create-collections#configure-a-collection)
    ```
    Auth Type: Bearer Token
    Token: {{token}}
    ```
1. Now you can play around and send request to the chorus backend server running locally

## Advanced Configuration

### Naming Convention

The backend's own config schema (everything under `internal/config/config.go`, i.e. every field documented in this section) uses **snake_case**, e.g. `daemon.http.headers.access_control_allow_origins`. This is deliberate, not an oversight: Viper lowercases every key inside any `map[string]...`-typed field during decode (`Log.Loggers`, `Daemon.Jobs`, `Storage.Datastores`, `Storage.FileStores`, `Services.AuthenticationService.Modes`), regardless of how it's written in the file — so a camelCase field like `appSync` would silently decode as `appsync`, breaking any Go code that looks it up by exact name (e.g. `internal/cmd/provider/jobber.go` dispatching a job by its config key). Since that limitation applies schema-wide in practice, the whole backend config stays snake_case for consistency, including named map instances like the `app_sync` job or the `chorus`/`audit` datastores.

This is separate from the Helm chart's own `values.yaml` (`deploy/backend/values.yaml`, `configs/*/values.yaml`), which follows the standard Helm/Kubernetes convention of camelCase (`imagePullPolicy`, `podLabels`, etc.) for its own chart-level fields — everything *outside* the `config:` block that gets passed straight through into `configs/config.yaml` as-is.

### Build a Full Config

`chorus init-config` (see [Quick Start](#quick-start)) covers local dev in one command. If you'd rather see every overridable field yourself and hand-edit `configs/config.yaml`, build it from the code-level defaults instead:

```bash
chorus export-default-config > configs/config.yaml
```

At minimum you'll still need to fill in every field `chorus check-config` reports as missing, including a `daemon.private_key`:
```bash
chorus generate-private-key
```
and a `services.openid_connect_provider.jwks`:
```bash
chorus generate-jwks
```
Both print to stdout — paste the result into the matching field in `configs/config.yaml`.

### Keep it in Sync

Whichever way you built `configs/config.yaml`:

* Check for drift against the code-level defaults at any time:
    ```bash
    make diff-config
    ```
* Validate your config against the required-field rules at any time:
    ```bash
    make check-config
    ```
* Optional: trim `configs/config.yaml` down to only the fields you actually changed (backs up to `configs/config.yaml.bak` first):
    ```bash
    make trim-config
    ```

The make commands above run against live source rather than the built `chorus` binary — handy if you're actively editing `internal/config/config.go` yourself.

### Resolution Order

Beyond editing `configs/config.yaml` directly, any value can be overridden two other ways: a `CHORUS_*` environment variable, or a one-off `--set path.to.key=value` flag.
Configuration resolves in this order — later wins:

1. code-level defaults (`provider.SetDefaultConfig()`, see `chorus export-default-config`)
2. `--config <file>`, repeatable — later files override earlier ones
3. `CHORUS_*` environment variables
4. `--set path.to.key=value`, repeatable

```bash
chorus start \
  --config configs/config.yaml \
  --set storage.datastores.chorus.database=chorus_ci
```

Environment variables use the dotted config path, uppercased, with `.` replaced by `_`, prefixed with `CHORUS_`: `storage.datastores.chorus.database` → `CHORUS_STORAGE_DATASTORES_CHORUS_DATABASE`.

## Developer doc.

Create a complete service (here the workbench service)

1. Interface
    - create service & entity protocol buffer definitions in api/proto/v1/workbench-service.proto and api/proto/v1/workbench.proto
    - `make protos`

1. Write the migration
    - implement migration in internal/migration/chorus/postgres/00003_workbench.sql
    - launch the backend to apply the new migration

1. Plug autogenerated server
    - tell the server there is a new service (both grpc and http) in internal/cmd/start.go

1. Provider (dependency injection)
    - implement provider in internal/cmd/provider/workbench.go (copy from internal/cmd/provider/workspace.go for instance)
    - implement provider for ctrl, service and store

1. Controller
    - implement controller in internal/api/v1/workbench-controller.go
    - implement controller auth middleware in internal/api/v1/middleware/workbench-authorization.go
    - implement converter in internal/api/v1/converter/workbench.go

1. Model
    - implement model in pkg/workbench/model/workbench.go

1. Service
    - implement service in pkg/workbench/service/workbench-service.go
    - implement service caching middleware in pkg/workbench/service/middleware/caching.go
    - implement service logging middleware in pkg/workbench/service/middleware/logging.go
    - implement service validation middleware in pkg/workbench/service/middleware/validation.go

1. Store
    - implement store in pkg/workbench/store/postgres/workbench-storage.go
    - implement store logging middleware in pkg/workbench/store/middleware/logging.go

1. Tests

    All test tiers are driven through make (see `make help` for every target).
    Coverage profiles land in `tests/coverage/` and can be browsed with
    `make coverage-html REPORT=unit|integration|acceptance`.

    **Unit Tests**

    Run all unit tests (writes `tests/coverage/unit.out`)
    ```bash
    make test-unit
    ```

    Run unit tests for a specific domain
    ```bash
    make test-unit PKG=workspace
    ```

    **Integration Tests**

    Store tests against an embedded postgres, no dependencies needed
    (writes `tests/coverage/integration.out`)
    ```bash
    make test-integration
    ```

    **Acceptance Tests**

    The suites talk to a backend over HTTP and need the `chorus_ci` database
    (created automatically by the dev-container init script; for an older
    postgres volume, create it once with `CREATE DATABASE chorus_ci;`).

    There's no separate CI config file — acceptance testing reuses
    `CONFIG_FILE` (default `configs/config.yaml`), with a handful of overrides
    (`chorus_ci` database, a separate disk file-store path, docker/k8s clients
    disabled) applied via `--set`. See `CONFIG_FILE`/`ACCEPTANCE_CONFIG_SET`
    at the top of `scripts/run_acceptance_tests.sh` to change the defaults.

    `make test-acceptance` is self-managed: it builds the backend, starts it,
    runs the suite, then stops it — one command, using whatever
    `daemon.http.port` is set in `CONFIG_FILE`
    ```bash
    make test-acceptance
    ```
    Or a single suite
    ```bash
    make test-acceptance SUITE=workspace
    ```

## License and Usage Restrictions

Any use of the software for purposes other than academic research, including for commercial purposes, shall be requested in advance from [CHUV](mailto:pactt.legal@chuv.ch).

## Acknowledgments

This project has received funding from the Swiss State Secretariat for Education, Research and Innovation (SERI) under contract number 23.00638, as part of the Horizon Europe project “EBRAINS 2.0”.
