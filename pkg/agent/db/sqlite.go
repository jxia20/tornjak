package db

import (
	"database/sql"

	backoff "github.com/cenkalti/backoff/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/spiffe/tornjak/pkg/agent/types"
)

const (
	initAgentsTable = `CREATE TABLE IF NOT EXISTS agents 
                            (id INTEGER PRIMARY KEY AUTOINCREMENT, 
                            spiffeid TEXT, 
                            plugin TEXT, 
                            last_seen DATETIME,
                            status TEXT,
                            UNIQUE (spiffeid))`

	initClustersTable = `CREATE TABLE IF NOT EXISTS clusters 
                            (id INTEGER PRIMARY KEY AUTOINCREMENT, 
                            uid TEXT, 
                            created_at TEXT, 
                            updated_at TEXT,
                            domain_name TEXT, 
                            platform_type TEXT, 
                            managed_by TEXT,
                            description TEXT,
                            UNIQUE (uid))`

	initClusterMemberTable = `CREATE TABLE IF NOT EXISTS cluster_memberships 
                            (id INTEGER PRIMARY KEY AUTOINCREMENT, 
                            agent_id INT, 
                            cluster_id INT,
                            joined_at TEXT,
                            role TEXT,
                            FOREIGN KEY (agent_id) REFERENCES agents(id), 
                            FOREIGN KEY (cluster_id) REFERENCES clusters(id), 
                            UNIQUE (agent_id))`
)

type LocalSqliteDb struct {
	database   *sql.DB
	expBackoff *backoff.BackOff
}

type Config struct {
	DriverName     string
	DbPath         string
	MaxConnections int
	BackOffParams  backoff.BackOff
}

func createDBTable(database *sql.DB, cmd string) error {
	statement, err := database.Prepare(cmd)
	if err != nil {
		return SQLError{cmd, err}
	}
	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		return SQLError{cmd, err}
	}
	return nil
}

func NewLocalSqliteDB(config Config) (AgentDB, error) {
	database, err := sql.Open(config.DriverName, config.DbPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open connection to DB")
	}

	if config.MaxConnections > 0 {
		database.SetMaxOpenConns(config.MaxConnections)
		database.SetMaxIdleConns(config.MaxConnections)
	}

	initTableList := []string{initAgentsTable, initClustersTable, initClusterMemberTable}
	for _, tableCmd := range initTableList {
		if err = createDBTable(database, tableCmd); err != nil {
			database.Close()
			return nil, err
		}
	}

	if err = database.Ping(); err != nil {
		database.Close()
		return nil, errors.Wrap(err, "failed to verify database connection")
	}

	return &LocalSqliteDb{
		database:   database,
		expBackoff: &config.BackOffParams,
	}, nil
}

// ... [The rest of the unchanged code, including all methods like CreateAgentEntry, GetClusterAgents, CreateClusterEntry, etc., are included here] ...

func (db *LocalSqliteDb) retryOp(operation func() error) error {
	err := backoff.Retry(operation, *db.expBackoff)
	if err != nil {
		if serr, ok := err.(*backoff.PermanentError); ok {
			return serr.Unwrap()
		}
	}
	return err
}

func (db *LocalSqliteDb) CreateClusterEntry(cinfo types.ClusterInfo) error {
	operation := func() error {
		return db.createClusterEntryOp(cinfo)
	}
	return db.retryOp(operation)
}

func (db *LocalSqliteDb) EditClusterEntry(cinfo types.ClusterInfo) error {
	operation := func() error {
		return db.editClusterEntryOp(cinfo)
	}
	return db.retryOp(operation)
}

func (db *LocalSqliteDb) DeleteClusterEntry(uid string) error {
	operation := func() error {
		return db.deleteClusterEntryOp(uid)
	}
	return db.retryOp(operation)
}
