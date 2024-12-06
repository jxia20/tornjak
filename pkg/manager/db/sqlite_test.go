package db

import (
	"os"
	"testing"

	"github.com/spiffe/tornjak/pkg/manager/types"
)

// cleanup removes the test database file to ensure a clean slate for tests.
func cleanup() {
	err := os.Remove("./local-test-db")
	if err != nil && !os.IsNotExist(err) {
		// Log if there's an unexpected error during cleanup
		panic("Failed to clean up local test database: " + err.Error())
	}
}

func TestServerCreate(t *testing.T) {
	// Ensure cleanup is called at the end of the test
	defer cleanup()

	// Initialize a new local SQLite DB for testing
	db, err := NewLocalSqliteDB("./local-test-db")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Verify that the initial server list is empty
	sList, err := db.GetServers()
	if err != nil {
		t.Fatalf("Failed to get servers from database: %v", err)
	}
	if len(sList.Servers) > 0 {
		t.Fatalf("Expected empty server list, but found %d servers", len(sList.Servers))
	}

	// Define server information for testing
	sinfo := types.ServerInfo{
		Name:    "my-server",
		Address: "http://localhost:10000",
	}

	// Attempt to create a new server entry
	err = db.CreateServerEntry(sinfo)
	if err != nil {
		t.Fatalf("Failed to create server entry: %v", err)
	}

	// Fetch the updated server list
	sList, err = db.GetServers()
	if err != nil {
		t.Fatalf("Failed to retrieve servers after creation: %v", err)
	}
	if len(sList.Servers) != 1 {
		t.Fatalf("Expected server list to contain 1 entry, but found %d", len(sList.Servers))
	}
	if sList.Servers[0].Name != sinfo.Name || sList.Servers[0].Address != sinfo.Address {
		t.Fatalf("Mismatch in server info: got %+v, want %+v", sList.Servers[0], sinfo)
	}
}
