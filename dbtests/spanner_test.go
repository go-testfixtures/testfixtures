//go:build spanner
// +build spanner

package dbtests

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/go-testfixtures/testfixtures/v3/shared"
	_ "github.com/googleapis/go-sql-spanner"
)

func TestSpanner(t *testing.T) {
	prepareSpannerDB(t)

	dialect := "spanner"
	db := openDB(t, dialect, os.Getenv("SPANNER_CONN_STRING"))
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/spanner.sql", []byte(";\n"))
	additionalOptions := []func(*testfixtures.Loader) error{testfixtures.DangerousSkipTestDatabaseCheck()}
	testLoader(t, db, dialect, additionalOptions...)

	t.Run("SpannerConstraints", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures"),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}

		constraintsBefore, _ := shared.GetConstraints(db)

		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		constraintsAfter, _ := shared.GetConstraints(db)

		assertSpannerConstraints(t, constraintsBefore, constraintsAfter)

		// Call load again to test against a database with existing data.
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		assertFixturesLoaded(t, db)
	})
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

//nolint:unused
func assertSpannerConstraints(t *testing.T, constraintsBefore, constraintsAfter map[string][]shared.SpannerConstraint) {
	if len(constraintsBefore) != len(constraintsAfter) {
		t.Errorf("constraints before and after should have the same length")
	}

	for key, cBefore := range constraintsBefore {
		cAfter := constraintsAfter[key]
		if len(cBefore) != len(cAfter) {
			t.Errorf("constraints before and after should have the same length")
		}
		for _, cBefore := range cBefore {
			found := false
			for _, cAfter := range cAfter {
				if reflect.DeepEqual(cBefore, cAfter) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("constraint %s not found in constraintsAfter", cBefore.ConstraintName)
			}
		}
	}
}
