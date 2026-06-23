package config

import (
	"strings"
	"testing"
)

func TestDatabaseDSNDoesNotEnableMultiStatements(t *testing.T) {
	dsn := Default().Database.DSN()

	if strings.Contains(dsn, "multiStatements=true") {
		t.Fatalf("application DSN must not enable multiStatements: %s", dsn)
	}
	if !strings.Contains(dsn, "parseTime=true") {
		t.Fatalf("application DSN must keep parseTime enabled: %s", dsn)
	}
}
