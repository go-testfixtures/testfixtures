package dbtests

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Fill any of those env variables if you want to set up a database on your own.
	clickhouseConnStringEnv = "CLICKHOUSE_CONN_STRING"
	crdbConnStringEnv       = "CRDB_CONN_STRING"
	mysqlConnStringEnv      = "MYSQL_CONN_STRING"
	pgConnStringEnv         = "PG_CONN_STRING"
	sqliteConnStringEnv     = "SQLITE_CONN_STRING"
	sqlserverConnStringEnv  = "SQLSERVER_CONN_STRING"
	spannerEndpointEnv      = "SPANNER_ENDPOINT"
)

func createClickhouseContainer(t *testing.T) string {
	t.Helper()

	if connStr := os.Getenv(clickhouseConnStringEnv); connStr != "" {
		return connStr
	}

	createConnString := func(host string, port string) string {
		return fmt.Sprintf("clickhouse://default:password@%s:%s/testdb", host, port)
	}

	const targetPort = "9000/tcp"
	const httpInterfacePort = "8123/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{targetPort, httpInterfacePort},
		Env: map[string]string{
			"CLICKHOUSE_DB":       "testdb",
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "password",
		},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/ping").WithPort(httpInterfacePort),
			wait.ForListeningPort(targetPort),
			wait.ForSQL(targetPort, "clickhouse", func(host string, port nat.Port) string {
				return createConnString(host, port.Port())
			}),
		).WithStartupTimeoutDefault(60 * time.Second),
	}
	host, port := createGenericContainer(t, req, targetPort)
	return createConnString(host, port)
}

func createCockroachDBContainer(t *testing.T) string {
	t.Helper()

	if connStr := os.Getenv(crdbConnStringEnv); connStr != "" {
		return connStr
	}

	createConnString := func(host string, port string) string {
		return fmt.Sprintf("postgres://root@%s:%s/defaultdb?sslmode=disable", host, port)
	}

	const targetPort = "26257/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "cockroachdb/cockroach:latest",
		ExposedPorts: []string{targetPort},
		Cmd: []string{
			"start-single-node",
			"--insecure",
			"--store=type=mem,size=0.25",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("CockroachDB node starting"),
			wait.ForListeningPort(targetPort),
			wait.ForSQL(targetPort, "postgres", func(host string, port nat.Port) string {
				return createConnString(host, port.Port())
			}),
		).WithStartupTimeoutDefault(60 * time.Second),
	}
	host, port := createGenericContainer(t, req, targetPort)
	return createConnString(host, port)
}

func createMySQLContainer(t *testing.T) string {
	t.Helper()

	if s := os.Getenv(mysqlConnStringEnv); s != "" {
		return s
	}

	createConnString := func(host string, port string) string {
		return fmt.Sprintf("root:testpass@tcp(%s:%s)/testdb", host, port)
	}

	const targetPort = "3306/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:latest",
		ExposedPorts: []string{targetPort},
		Env: map[string]string{
			"MARIADB_ROOT_PASSWORD": "testpass",
			"MARIADB_DATABASE":      "testdb",
		},
		// WaitFor looks crazy for MariaDB, because I had a lot of issues with it and I want to make as
		// robust as possible.
		WaitingFor: wait.ForAll(
			wait.ForLog("Temporary server stopped"),
			wait.ForLog("ready for connections"),
			wait.ForListeningPort(targetPort),
			wait.ForSQL(targetPort, "mysql", func(host string, port nat.Port) string {
				return createConnString(host, port.Port())
			}),
		).WithStartupTimeoutDefault(60 * time.Second),
	}
	host, port := createGenericContainer(t, req, targetPort)
	return createConnString(host, port)
}

func createPostgreSQLContainer(t *testing.T) string {
	t.Helper()

	if connStr := os.Getenv(pgConnStringEnv); connStr != "" {
		return connStr
	}

	createConnString := func(host string, port string) string {
		return fmt.Sprintf("postgres://postgres:testpass@%s:%s/testdb?sslmode=disable", host, port)
	}

	const targetPort = "5432/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{targetPort},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(targetPort),
			wait.ForSQL(targetPort, "postgres", func(host string, port nat.Port) string {
				return createConnString(host, port.Port())
			}),
		).WithStartupTimeoutDefault(60 * time.Second),
	}
	host, port := createGenericContainer(t, req, targetPort)
	return createConnString(host, port)
}

func createSQLite(t *testing.T) string {
	t.Helper()

	if connStr := os.Getenv(sqliteConnStringEnv); connStr != "" {
		return connStr
	}

	// Combination of params, which work the best for testing:
	// * name is always unique, so there is no interaction between tests
	// * mode memory stores a database in memory, so it is less dependent on a file system
	// * cache shared allows opening multiple connections without creating a brand-new database each time
	return fmt.Sprintf("file:testdb-%s?mode=memory&cache=shared", uuid.New().String())
}

func createSQLServerContainer(t *testing.T) string {
	t.Helper()

	if connStr := os.Getenv(sqlserverConnStringEnv); connStr != "" {
		return connStr
	}

	createConnString := func(host string, port string) string {
		return fmt.Sprintf("server=%s;port=%s;user id=sa;password=SQL@1server;database=master;encrypt=disable", host, port)
	}

	const targetPort = "1433/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "mcr.microsoft.com/mssql/server:latest",
		ExposedPorts: []string{targetPort},
		Env: map[string]string{
			"ACCEPT_EULA": "Y",
			"MSSQL_PID":   "Express",
			"SA_PASSWORD": "SQL@1server", // password needs to be "strong"
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("SQL Server is now ready for client connections"),
			wait.ForListeningPort(targetPort),
			wait.ForSQL(targetPort, "sqlserver", func(host string, port nat.Port) string {
				return createConnString(host, port.Port())
			}),
		).WithStartupTimeoutDefault(60 * time.Second),
	}

	host, port := createGenericContainer(t, req, targetPort)
	return createConnString(host, port)
}

func createSpannerContainer(t *testing.T) string {
	t.Helper()

	if endpoint := os.Getenv(spannerEndpointEnv); endpoint != "" {
		return endpoint
	}

	const targetPort = "9010/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/cloud-spanner-emulator/emulator:latest",
		ExposedPorts: []string{targetPort},
		WaitingFor: wait.ForAll(
			wait.ForLog("gRPC server listening"),
			wait.ForListeningPort(targetPort),
		).WithStartupTimeoutDefault(60 * time.Second),
	}

	host, port := createGenericContainer(t, req, targetPort)
	return net.JoinHostPort(host, port)
}

func createGenericContainer(t *testing.T, req testcontainers.ContainerRequest, targetPort string) (host string, port string) {
	t.Helper()

	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Logger:           log.TestLogger(t),
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}
	t.Cleanup(cleanup)

	return containerHostAndPort(t, container, targetPort)
}

func containerHostAndPort(t *testing.T, container testcontainers.Container, port string) (host string, mappedPort string) {
	t.Helper()
	natPort := nat.Port(port)
	mappedPortObj, err := container.MappedPort(t.Context(), natPort)
	if err != nil {
		t.Fatalf("failed to get mapped port for port %s: %v", port, err)
	}

	host, err = container.Host(t.Context())
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	return host, mappedPortObj.Port()
}
