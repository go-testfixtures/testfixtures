//go:build googlesql
// +build googlesql

package testfixtures

import (
	"context"
	"fmt"
	"os"
	"testing"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
)

func TestGoogleSQL(t *testing.T) {
	prepareSpannerDB(t)

	testLoader(
		t,
		"spanner",
		os.Getenv("GOOGLESQL_CONN_STRING"),
		"testdata/schema/googlesql.sql",
	)
}

func prepareSpannerDB(t *testing.T) {
	t.Helper()

	var err error
	if err = startEmulator(); err != nil {
		t.Fatalf("failed to start emulator: %v", err)
	}
	defer func() {
		stopEmulator()
	}()

	projectId, instanceId, databaseId := "test-project", "test-instance", "testdb"
	if err = createInstance(projectId, instanceId); err != nil {
		t.Fatalf("failed to create instance on emulator: %v", err)
	}
	if err = createSampleDB(projectId, instanceId, databaseId); err != nil {
		t.Fatalf("failed to create database on emulator: %v", err)
	}
}

func startEmulator() error {
	// ctx := context.Background()
	if err := os.Setenv("SPANNER_EMULATOR_HOST", "googlesql:9010"); err != nil {
		return err
	}

	// 	// Initialize a Docker client.
	// 	var err error
	// 	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// Pull the Spanner Emulator docker image.
	// 	reader, err := cli.ImagePull(ctx, "gcr.io/cloud-spanner-emulator/emulator", types.ImagePullOptions{})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer func() { _ = reader.Close() }()
	// 	// cli.ImagePull is asynchronous.
	// 	// The reader needs to be read completely for the pull operation to complete.
	// 	if _, err := io.Copy(io.Discard, reader); err != nil {
	// 		return err
	// 	}

	// 	// Create and start a container with the emulator.
	// 	resp, err := cli.ContainerCreate(ctx, &container.Config{
	// 		Image:        "gcr.io/cloud-spanner-emulator/emulator",
	// 		ExposedPorts: nat.PortSet{"9010": {}},
	// 	}, &container.HostConfig{
	// 		PortBindings: map[nat.Port][]nat.PortBinding{"9010": {{HostIP: "0.0.0.0", HostPort: "9010"}}},
	// 	}, nil, nil, "")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	containerId = resp.ID
	// 	if err := cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{}); err != nil {
	// 		return err
	// 	}
	// 	// Wait max 10 seconds or until the emulator is running.
	// 	for c := 0; c < 20; c++ {
	// 		// Always wait at least 500 milliseconds to ensure that the emulator is actually ready, as the
	// 		// state can be reported as ready, while the emulator (or network interface) is actually not ready.
	// 		<-time.After(500 * time.Millisecond)
	// 		resp, err := cli.ContainerInspect(ctx, containerId)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to inspect container state: %v", err)
	// 		}
	// 		if resp.State.Running {
	// 			break
	// 		}
	// 	}

	return nil
}

func createInstance(projectId, instanceId string) error {
	ctx := context.Background()
	instanceAdmin, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return err
	}
	defer instanceAdmin.Close()
	op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectId),
		InstanceId: instanceId,
		Instance: &instancepb.Instance{
			Config:      fmt.Sprintf("projects/%s/instanceConfigs/%s", projectId, "emulator-config"),
			DisplayName: instanceId,
			NodeCount:   1,
		},
	})
	if err != nil {
		return fmt.Errorf("could not create instance %s: %v", fmt.Sprintf("projects/%s/instances/%s", projectId, instanceId), err)
	}
	// Wait for the instance creation to finish.
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for instance creation to finish failed: %v", err)
	}
	return nil
}

func createSampleDB(projectId, instanceId, databaseId string, statements ...string) error {
	ctx := context.Background()
	databaseAdminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer databaseAdminClient.Close()
	opDB, err := databaseAdminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectId, instanceId),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseId),
		ExtraStatements: statements,
	})
	if err != nil {
		return err
	}
	// Wait for the database creation to finish.
	if _, err := opDB.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for database creation to finish failed: %v", err)
	}
	return nil
}

func stopEmulator() {
	//	if cli == nil || containerId == "" {
	//		return
	//	}
	//
	// ctx := context.Background()
	// timeout := 10
	//
	//	if err := cli.ContainerStop(ctx, containerId, container.StopOptions{Timeout: &timeout}); err != nil {
	//		log.Printf("failed to stop emulator: %v\n", err)
	//	}
}
