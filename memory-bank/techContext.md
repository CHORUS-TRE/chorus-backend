# Tech Context: Chorus Backend

## Technologies

- **Language**: Go (version 1.24.3)
- **Database**: PostgreSQL
- **API**: gRPC, REST (via gRPC-Gateway), OpenAPI
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: Jenkins

## Key Libraries & Frameworks

- **gRPC**: `google.golang.org/grpc` for high-performance RPC.
- **HTTP Framework**: `net/http` implicitly through `grpc-gateway`.
- **CLI**: `github.com/spf13/cobra` for building the application's command-line interface.
- **Configuration**: `github.com/spf13/viper` for managing configuration from files, environment variables, etc.
- **Logging**: `go.uber.org/zap` for structured, high-performance logging.
- **Database Access**: `github.com/jmoiron/sqlx` and `github.com/lib/pq` for interacting with PostgreSQL.
- **Database Migrations**: `github.com/rubenv/sql-migrate` for versioned database schema migrations.
- **Authentication**: `github.com/golang-jwt/jwt` for JWT implementation, `github.com/pquerna/otp` for one-time passwords.
- **Testing**: `github.com/onsi/ginkgo` and `github.com/onsi/gomega` for BDD-style testing, `github.com/stretchr/testify` for assertions.
- **Kubernetes Client**: `k8s.io/client-go` for interacting with the Kubernetes API.

## Development Setup

The `README.md` provides instructions for setting up a local development environment. The key steps are:
1.  **Install Go**: Version 1.24.3 or compatible.
2.  **Run PostgreSQL**: A `docker-compose.yml` file is provided in `.devcontainer` to spin up a PostgreSQL instance.
3.  **Set up Local Kubernetes**: The `scripts/create-local-cluster.sh` script uses `kind` to create a local Kubernetes cluster.
4.  **Run the Backend**: The application can be started with `go run cmd/chorus/main.go start`. A companion logger can be run with `go run cmd/logger/main.go`.
5.  **API Documentation**: Once running, the OpenAPI UI is available at `http://localhost:5000/doc`.

## Technical Constraints

- The project relies on a specific set of tools for development and deployment (Docker, Kind, Jenkins). Adhering to these tools is important for consistency.
- Database migrations must be written in SQL and follow the sequence defined in `internal/migration/postgres/`.
- All new services should follow the established architectural patterns to maintain consistency.
