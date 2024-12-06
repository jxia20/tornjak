package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/spiffe/tornjak/pkg/agent/types"
)

const (
	// agent table with fields spiffeid and plugin
	initAgentsTable = `CREATE TABLE IF NOT EXISTS agents 
                            (id INTEGER PRIMARY KEY AUTOINCREMENT, 
                            spiffeid TEXT, 
                            plugin TEXT, 
                            last_seen DATETIME,
                            status TEXT,
                            UNIQUE (spiffeid))`

	// cluster table with enhanced fields
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

	// cluster - agent relation table with additional metadata
	initClusterMemberTable = `CREATE TABLE IF NOT EXISTS cluster_memberships 
                            (id INTEGER PRIMARY KEY AUTOINCREMENT, 
                            agent_id int, 
                            cluster_id int,
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

// Config holds database configuration
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

	// Set connection pool parameters
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

	// Verify database connection
	if err = database.Ping(); err != nil {
		database.Close()
		return nil, errors.Wrap(err, "failed to verify database connection")
	}

	return &LocalSqliteDb{
		database:   database,
		expBackoff: &config.BackOffParams,
	}, nil
}

// AGENT - SELECTOR/PLUGIN HANDLERS

func (db *LocalSqliteDb) CreateAgentEntry(sinfo types.AgentInfo) error {
	cmdInsert := `INSERT INTO agents (spiffeid, plugin, last_seen, status) VALUES `
	cmdUpdate := ` ON CONFLICT(spiffeid) DO UPDATE SET plugin=?, last_seen=?, status=?`

	now := time.Now().UTC().Format(time.RFC3339)
	status := "ACTIVE"

	if len(sinfo.Plugin) > 0 {
		cmdInsert += `(?, ?, ?, ?)`
	} else {
		cmdInsert += `(?, NULL, ?, ?)`
	}

	cmd := cmdInsert + cmdUpdate
	statement, err := db.database.Prepare(cmd)
	if err != nil {
		return SQLError{cmd, err}
	}
	defer statement.Close()

	if len(sinfo.Plugin) > 0 {
		_, err = statement.Exec(sinfo.Spiffeid, sinfo.Plugin, now, status, sinfo.Plugin, now, status)
	} else {
		_, err = statement.Exec(sinfo.Spiffeid, now, status, nil, now, status)
	}

	if err != nil {
		return SQLError{cmd, err}
	}
	return nil
}

func (db *LocalSqliteDb) GetAgentSelectors() (types.AgentInfoList, error) {
	cmd := `SELECT spiffeid, plugin, last_seen, status 
           FROM agents 
           WHERE plugin IS NOT NULL`

	rows, err := db.database.Query(cmd)
	if err != nil {
		return types.AgentInfoList{}, SQLError{cmd, err}
	}
	defer rows.Close()

	var sinfos []types.AgentInfo
	for rows.Next() {
		var (
			spiffeid string
			plugin   string
			lastSeen string
			status   string
		)

		if err = rows.Scan(&spiffeid, &plugin, &lastSeen, &status); err != nil {
			return types.AgentInfoList{}, SQLError{cmd, err}
		}

		sinfos = append(sinfos, types.AgentInfo{
			Spiffeid: spiffeid,
			Plugin:   plugin,
			LastSeen: lastSeen,
			Status:   status,
		})
	}

	if err = rows.Err(); err != nil {
		return types.AgentInfoList{}, SQLError{cmd, err}
	}

	return types.AgentInfoList{
		Agents: sinfos,
	}, nil
}

func (db *LocalSqliteDb) GetAgentPluginInfo(spiffeid string) (types.AgentInfo, error) {
	cmd := `SELECT spiffeid, plugin, last_seen, status 
           FROM agents 
           WHERE spiffeid=?`

	row := db.database.QueryRow(cmd, spiffeid)

	var sinfo types.AgentInfo
	var plugin, lastSeen, status sql.NullString

	err := row.Scan(&sinfo.Spiffeid, &plugin, &lastSeen, &status)
	if err == sql.ErrNoRows {
		return types.AgentInfo{}, GetError{fmt.Sprintf("Agent %v has no assigned plugin", spiffeid)}
	} else if err != nil {
		return types.AgentInfo{}, SQLError{cmd, err}
	}

	if plugin.Valid {
		sinfo.Plugin = plugin.String
	}
	if lastSeen.Valid {
		sinfo.LastSeen = lastSeen.String
	}
	if status.Valid {
		sinfo.Status = status.String
	}

	return sinfo, nil
}

// GetClusterAgents takes in string cluster name and outputs array of spiffeids of agents assigned to the cluster
func (db *LocalSqliteDb) GetClusterAgents(name string) ([]string, error) {
	cmdGetMemberships := `SELECT GROUP_CONCAT(agents.spiffeid) 
                        FROM clusters 
                        LEFT JOIN cluster_memberships ON clusters.id=cluster_memberships.cluster_id
                        LEFT JOIN agents ON cluster_memberships.agent_id=agents.id
                        WHERE clusters.uid=? 
                        GROUP BY clusters.uid`

	row := db.database.QueryRow(cmdGetMemberships, name)

	var spiffeidList []string
	var spiffeids sql.NullString

	err := row.Scan(&spiffeids)
	if err == sql.ErrNoRows {
		return nil, GetError{fmt.Sprintf("Cluster %v not registered", name)}
	} else if err != nil {
		return nil, SQLError{cmdGetMemberships, err}
	}

	if spiffeids.Valid {
		spiffeidList = strings.Split(spiffeids.String, ",")
	} else {
		spiffeidList = []string{}
	}

	return spiffeidList, nil
}

// GetAgentClusterName takes in string of spiffeid of agent and outputs the name of the cluster
func (db *LocalSqliteDb) GetAgentClusterName(spiffeid string) (string, error) {
	cmdGetName := `SELECT clusters.uid, clusters.domain_name 
                  FROM agents 
                  LEFT JOIN cluster_memberships ON agents.id=cluster_memberships.agent_id
                  LEFT JOIN clusters ON cluster_memberships.cluster_id=clusters.id
                  WHERE agents.spiffeid=?`

	row := db.database.QueryRow(cmdGetName, spiffeid)

	var uid, domainName sql.NullString
	err := row.Scan(&uid, &domainName)
	if err == sql.ErrNoRows {
		return "", GetError{fmt.Sprintf("Agent %v unassigned to any cluster", spiffeid)}
	} else if err != nil {
		return "", SQLError{cmdGetName, err}
	}

	if !uid.Valid || !domainName.Valid {
		return "", GetError{fmt.Sprintf("Agent %v assigned to unregistered cluster", spiffeid)}
	}

	return fmt.Sprintf("%s (%s)", uid.String, domainName.String), nil
}

// GetAgentsMetadata returns detailed information about specified agents
func (db *LocalSqliteDb) GetAgentsMetadata(req types.AgentMetadataRequest) (types.AgentInfoList, error) {
	cmd := `SELECT a.spiffeid, a.plugin, a.last_seen, a.status,
                  c.uid, c.domain_name, cm.role, cm.joined_at
           FROM agents a
           LEFT JOIN cluster_memberships cm ON a.id = cm.agent_id
           LEFT JOIN clusters c ON cm.cluster_id = c.id`

	var args []interface{}
	if len(req.Agents) > 0 {
		placeholders := make([]string, len(req.Agents))
		for i, spiffeid := range req.Agents {
			placeholders[i] = "?"
			args = append(args, spiffeid)
		}
		cmd += " WHERE a.spiffeid IN (" + strings.Join(placeholders, ",") + ")"
	}

	rows, err := db.database.Query(cmd, args...)
	if err != nil {
		return types.AgentInfoList{}, SQLError{cmd, err}
	}
	defer rows.Close()

	var ainfos []types.AgentInfo
	for rows.Next() {
		var (
			spiffeid, lastSeen, status                     string
			plugin, clusterUID, domainName, role, joinedAt sql.NullString
		)

		if err = rows.Scan(&spiffeid, &plugin, &lastSeen, &status,
			&clusterUID, &domainName, &role, &joinedAt); err != nil {
			return types.AgentInfoList{}, SQLError{cmd, err}
		}

		agent := types.AgentInfo{
			Spiffeid: spiffeid,
			LastSeen: lastSeen,
			Status:   status,
		}

		if plugin.Valid {
			agent.Plugin = plugin.String
		}
		if clusterUID.Valid && domainName.Valid {
			agent.Cluster = fmt.Sprintf("%s (%s)", clusterUID.String, domainName.String)
		}
		if role.Valid {
			agent.Role = role.String
		}
		if joinedAt.Valid {
			agent.JoinedAt = joinedAt.String
		}

		ainfos = append(ainfos, agent)
	}

	if err = rows.Err(); err != nil {
		return types.AgentInfoList{}, SQLError{cmd, err}
	}

	return types.AgentInfoList{
		Agents: ainfos,
	}, nil
}

// GetClusters returns information about all registered clusters
func (db *LocalSqliteDb) GetClusters() (types.ClusterInfoList, error) {
	cmd := `SELECT c.uid, c.created_at, c.updated_at, c.domain_name, c.managed_by, 
                  c.platform_type, c.description, GROUP_CONCAT(a.spiffeid) 
           FROM clusters c
           LEFT JOIN cluster_memberships cm ON c.id=cm.cluster_id
           LEFT JOIN agents a ON cm.agent_id=a.id
           GROUP BY c.uid`

	rows, err := db.database.Query(cmd)
	if err != nil {
		return types.ClusterInfoList{}, SQLError{cmd, err}
	}
	defer rows.Close()

	var clusters []types.ClusterInfo
	for rows.Next() {
		var (
			uid, createdAt, updatedAt, domainName, managedBy, platformType, description string
			agentsList                                                                  sql.NullString
		)

		if err = rows.Scan(&uid, &createdAt, &updatedAt, &domainName, &managedBy,
			&platformType, &description, &agentsList); err != nil {
			return types.ClusterInfoList{}, SQLError{cmd, err}
		}

		agents := []string{}
		if agentsList.Valid {
			agents = strings.Split(agentsList.String, ",")
		}

		clusters = append(clusters, types.ClusterInfo{
			UID:          uid,
			CreationTime: createdAt,
			UpdatedAt:    updatedAt,
			DomainName:   domainName,
			ManagedBy:    managedBy,
			PlatformType: platformType,
			Description:  description,
			AgentsList:   agents,
		})
	}

	if err = rows.Err(); err != nil {
		return types.ClusterInfoList{}, SQLError{cmd, err}
	}

	return types.ClusterInfoList{
		Clusters: clusters,
	}, nil
}

// CreateClusterEntry takes in struct cinfo of type ClusterInfo.  If a cluster with cinfo.Name already registered, returns error.
func (db *LocalSqliteDb) createClusterEntryOp(cinfo types.ClusterInfo) error {
	// BEGIN transaction
	ctx := context.Background()
	tx, err := db.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Errorf("Error initializing context: %v", err)
	}
	txHelper := getTornjakTxHelper(ctx, tx)

	// Generate a new UID if it is not provided
	if cinfo.UID == "" {
		newUID, uuidErr := uuid.NewUUID()
		if uuidErr != nil {
			return errors.Errorf("Error generating UID: %v", uuidErr)
		}
		cinfo.UID = newUID.String()
	}

	// INSERT cluster metadata
	err = txHelper.insertClusterMetadata(cinfo)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	// ADD agents to cluster
	err = txHelper.addAgentBatchToCluster(cinfo.UID, cinfo.AgentsList)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}
	return tx.Commit()
}

// EditClusterEntry takes in struct cinfo of type ClusterInfo.  If cluster with cinfo.Name does not exist, throws error.
func (db *LocalSqliteDb) editClusterEntryOp(cinfo types.ClusterInfo) error {
	// BEGIN transaction
	ctx := context.Background()
	tx, err := db.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Errorf("Error initializing context: %v", err)
	}
	txHelper := getTornjakTxHelper(ctx, tx)

	// UPDATE cluster metadata
	err = txHelper.updateClusterMetadata(cinfo)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	// REMOVE all currently assigned cluster agents
	err = txHelper.deleteClusterAgents(cinfo.UID)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	// ADD agents to cluster
	err = txHelper.addAgentBatchToCluster(cinfo.UID, cinfo.AgentsList)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	return tx.Commit()
}

// DeleteClusterEntry takes in string name of cluster and removes cluster information and agent membership of cluster from the database.  If not all agents can be removed from the cluster, cluster information remains in the database.
func (db *LocalSqliteDb) deleteClusterEntryOp(uid string) error {
	// BEGIN transaction
	ctx := context.Background()
	tx, err := db.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Errorf("Error initializing context: %v", err)
	}
	txHelper := getTornjakTxHelper(ctx, tx)

	// REMOVE all currently assigned cluster agents (requires metadata still entered)
	err = txHelper.deleteClusterAgents(uid)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	// REMOVE cluster metadata
	err = txHelper.deleteClusterMetadata(uid)
	if err != nil {
		return backoff.Permanent(txHelper.rollbackHandler(err))
	}

	return tx.Commit()
}

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
