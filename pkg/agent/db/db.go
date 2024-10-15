package db

import (
	"github.com/spiffe/tornjak/pkg/agent/types"
)

type AgentDB interface {
	// AGENT - SELECTOR/PLUGIN interface
	CreateAgentEntry(sinfo types.AgentInfo) error
	GetAgentSelectors() (types.AgentInfoList, error)
	GetAgentPluginInfo(uid string) (types.AgentInfo, error)

	// CLUSTER interface
	GetClusters() (types.ClusterInfoList, error)
	CreateClusterEntry(cinfo types.ClusterInfo) error
	EditClusterEntry(cinfo types.ClusterInfo) error
	DeleteClusterEntry(name string) error

	// AGENT - CLUSTER Get interface (for testing)e
	GetAgentClusterName(spiffeid string) (string, error)
	GetClusterAgents(uid string) ([]string, error)
	GetAgentsMetadata(req types.AgentMetadataRequest) (types.AgentInfoList, error)
}
