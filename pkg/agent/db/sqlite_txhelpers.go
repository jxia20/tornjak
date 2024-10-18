package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spiffe/tornjak/pkg/agent/types"
)

type tornjakTxHelper struct {
	ctx context.Context
	tx  *sql.Tx
}

func getTornjakTxHelper(ctx context.Context, tx *sql.Tx) *tornjakTxHelper {
	return &tornjakTxHelper{ctx, tx}
}

func (t *tornjakTxHelper) rollbackHandler(err error) error {
	if err == nil { // THIS SHOULD NOT HAPPEN
		return errors.New("Rollback handler called upon no error")
	} else {
		rollbackErr := t.tx.Rollback()
		var rollbackStatus string
		if rollbackErr != nil {
			rollbackStatus = fmt.Sprintf("[Unsuccessful rollback [%v] upon error]", rollbackErr.Error())
		} else {
			rollbackStatus = "[Successful rollback upon error]"
		}
		if serr, ok := err.(SQLError); ok {
			return SQLError{serr.Cmd, errors.Errorf("%v: %v", serr.Err, rollbackStatus)}
		} else if serr, ok := err.(GetError); ok {
			return GetError{fmt.Sprintf("%v: %v", serr.Message, rollbackStatus)}
		} else if serr, ok := err.(PostFailure); ok {
			return PostFailure{fmt.Sprintf("%v: %v", serr.Message, rollbackStatus)}
		} else {
			return errors.Errorf("%v: %v", err.Error(), rollbackStatus)
		}
	}
}

// insertClusterMetadata attempts insert into table clusters
// returns SQLError upon failure and PostFailure on cluster existence
func (t *tornjakTxHelper) insertClusterMetadata(cinfo types.ClusterInfo) error {
	// INSERT statement now includes UID
	cmdInsert := `INSERT INTO clusters (uid, name, created_at, domain_name, managed_by, platform_type) VALUES (?,?,?,?,?,?)`
	statement, err := t.tx.PrepareContext(t.ctx, cmdInsert)
	if err != nil {
		return SQLError{cmdInsert, err}
	}
	defer statement.Close()

	// Using cinfo.UID instead of relying only on name
	_, err = statement.ExecContext(t.ctx, cinfo.UID, cinfo.Name, time.Now().Format("Jan 02 2006 15:04:05"), cinfo.DomainName, cinfo.ManagedBy, cinfo.PlatformType)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return PostFailure{"Cluster already exists; use Edit Cluster"}
		}
		return SQLError{cmdInsert, err}
	}
	return nil
}

// updateClusterMetadata attempts update of entry in table clusters
// returns SQLError on failure and PostFailure on cluster non-existence
func (t *tornjakTxHelper) updateClusterMetadata(cinfo types.ClusterInfo) error {
	// Update based on UID instead of name
	cmdUpdate := `UPDATE clusters SET name=?, domain_name=?, managed_by=?, platform_type=? WHERE uid=?`
	statement, err := t.tx.PrepareContext(t.ctx, cmdUpdate)
	if err != nil {
		return SQLError{cmdUpdate, err}
	}
	defer statement.Close()

	res, err := statement.ExecContext(t.ctx, cinfo.Name, cinfo.DomainName, cinfo.ManagedBy, cinfo.PlatformType, cinfo.UID) // Using UID here
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return PostFailure{"Cluster already exists; use Edit Cluster"}
		}
		return SQLError{cmdUpdate, err}
	}

	// check if update was successful
	numRows, err := res.RowsAffected()
	if err != nil {
		return SQLError{cmdUpdate, err}
	}
	if numRows != 1 {
		return PostFailure{"Cluster does not exist; use Create Cluster"}
	}

	return nil
}

// deleteClusterMetadata attemps delete of entry in table clusters
// returns SQLError on failure and PostFailure on cluster non-existence
func (t *tornjakTxHelper) deleteClusterMetadata(uid string) error {
	// Delete based on UID instead of name
	cmdDelete := `DELETE FROM clusters WHERE uid=?`
	statement, err := t.tx.PrepareContext(t.ctx, cmdDelete)
	if err != nil {
		return SQLError{cmdDelete, err}
	}
	res, err := statement.ExecContext(t.ctx, uid)
	if err != nil {
		return SQLError{cmdDelete, err}
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return SQLError{cmdDelete, err}
	}
	if numRows != 1 {
		return PostFailure{"Cluster does not exist"}
	}
	return nil
}

// addAgentBatchToCluster adds entries in clusterMemberships table
// takes in cluster name and list of agent spiffeids
// returns SQLError on failure and PostFailure on conflict (an agent is already assigned)
func (t *tornjakTxHelper) addAgentBatchToCluster(clusterUID string, agentsList []string) error {
	if len(agentsList) == 0 {
		return nil
	}
	// Add into agents table
	cmdAgents := "INSERT OR IGNORE INTO agents (spiffeid, plugin) VALUES "
	agents := []interface{}{}
	for i := 0; i < len(agentsList); i++ {
		cmdAgents += "(?, NULL),"
		agents = append(agents, agentsList[i])
	}
	cmdAgents = strings.TrimSuffix(cmdAgents, ",")
	statementAgentInsert, err := t.tx.PrepareContext(t.ctx, cmdAgents)
	if err != nil {
		return SQLError{cmdAgents, err}
	}
	_, err = statementAgentInsert.ExecContext(t.ctx, agents...)
	if err != nil {
		return SQLError{cmdAgents, err}
	}

	// generate single statement
	cmdBatch := "INSERT OR ABORT INTO cluster_memberships (agent_id, cluster_id) VALUES "
	vals := []interface{}{}
	for i := 0; i < len(agentsList); i++ {
		// Using UID to find the cluster
		cmdBatch += "((SELECT id FROM agents WHERE spiffeid=?), (SELECT id FROM clusters WHERE uid=?)),"
		vals = append(vals, agentsList[i], clusterUID)
	}
	cmdBatch = strings.TrimSuffix(cmdBatch, ",")

	// prepare statement
	statementInsert, err := t.tx.PrepareContext(t.ctx, cmdBatch)
	if err != nil {
		return SQLError{cmdBatch, err}
	}
	_, err = statementInsert.ExecContext(t.ctx, vals...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return PostFailure{"Agent already assigned to another cluster"}
		}
		return SQLError{cmdBatch, err}
	}
	return nil
}

// deleteClusterAgents attempts removal of all agent-cluster pairs in clusterMemberships table
// returns SQLError on failure
func (t *tornjakTxHelper) deleteClusterAgents(clusterUID string) error {
	// Delete based on UID instead of name
	cmdDelete := "DELETE FROM cluster_memberships WHERE cluster_id=(SELECT id FROM clusters WHERE uid=?)"
	statementDelete, err := t.tx.PrepareContext(t.ctx, cmdDelete)
	if err != nil {
		return SQLError{cmdDelete, err}
	}
	_, err = statementDelete.ExecContext(t.ctx, clusterUID) // Use UID here
	if err != nil {
		return SQLError{cmdDelete, err}
	}
	return nil
}
