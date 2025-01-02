//go:build spanner
// +build spanner

package dbtests

import (
	"context"
	"fmt"
	"os"
	"testing"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/googleapis/go-sql-spanner"
)

func TestSpanner(t *testing.T) {
	prepareSpannerDB(t)

	db := openDB(t, "spanner", os.Getenv("SPANNER_CONN_STRING"))
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/spanner.sql", []byte(";\n"))
	testLoader(t, db, "spanner", testfixtures.DangerousSkipTestDatabaseCheck())
}

func prepareSpannerDB(t *testing.T) {
	t.Helper()

	if err := os.Setenv("SPANNER_EMULATOR_HOST", "spanner:9010"); err != nil {
		t.Fatalf("failed to set SPANNER_EMULATOR_HOST: %v", err)
	}

	projectId, instanceId, databaseId := "test-project", "test-instance", "testdb"
	if err := createInstance(projectId, instanceId); err != nil {
		t.Fatalf("failed to create instance on emulator: %v", err)
	}
	if err := createSampleDB(projectId, instanceId, databaseId); err != nil {
		t.Fatalf("failed to create database on emulator: %v", err)
	}
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
