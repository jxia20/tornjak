package api

import (
	"errors"
	"log"

	"github.com/google/uuid"

	tornjakTypes "github.com/spiffe/tornjak/pkg/agent/types"
)

/*

Agent

ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
BanAgent(BanAgentRequest) returns (google.protobuf.Empty);
DeleteAgent(DeleteAgentRequest) returns (google.protobuf.Empty);
CreateJoinToken(CreateJoinTokenRequest) returns (spire.types.JoinToken);

Entries

ListEntries(ListEntriesRequest) returns (ListEntriesResponse);
BatchCreateEntry(BatchCreateEntryRequest) returns (BatchCreateEntryResponse);
GetEntry(GetEntryRequest) returns (spire.types.Entry);

*/

type ListSelectorsRequest struct{}
type ListSelectorsResponse tornjakTypes.AgentInfoList

// ListSelectors returns list of agents from the local DB with the following info
// spiffeid string
// plugin   string
func (s *Server) ListSelectors(inp ListSelectorsRequest) (*ListSelectorsResponse, error) {
	log.Println("ListSelectors: Fetching agent selectors from the database.")
	resp, err := s.Db.GetAgentSelectors()
	if err != nil {
		log.Printf("ListSelectors: Error fetching selectors - %v", err)
		return nil, err
	}
	return (*ListSelectorsResponse)(&resp), nil
}

type RegisterSelectorRequest tornjakTypes.AgentInfo

// DefineSelectors registers an agent to the local DB with the following info
// spiffeid string
// plugin   string
func (s *Server) DefineSelectors(inp RegisterSelectorRequest) error {
	log.Println("DefineSelectors: Registering agent selectors to the database.")
	sinfo := tornjakTypes.AgentInfo(inp)
	if len(sinfo.Spiffeid) == 0 {
		return errors.New("DefineSelectors: Missing mandatory field - Spiffeid")
	}
	err := s.Db.CreateAgentEntry(sinfo)
	if err != nil {
		log.Printf("DefineSelectors: Error creating agent entry - %v", err)
	}
	return err
}

type ListAgentMetadataRequest tornjakTypes.AgentMetadataRequest
type ListAgentMetadataResponse tornjakTypes.AgentInfoList

// ListAgentMetadata takes in list of agent spiffeids
// and returns list of those agents from the local DB with following info
// spiffeid string
// plugin string
// cluster string
func (s *Server) ListAgentMetadata(inp ListAgentMetadataRequest) (*ListAgentMetadataResponse, error) {
	log.Println("ListAgentMetadata: Fetching agent metadata from the database.")
	inpReq := tornjakTypes.AgentMetadataRequest(inp)
	resp, err := s.Db.GetAgentsMetadata(inpReq)
	if err != nil {
		log.Printf("ListAgentMetadata: Error fetching agent metadata - %v", err)
		return nil, err
	}
	return (*ListAgentMetadataResponse)(&resp), nil
}

type ListClustersRequest struct{}
type ListClustersResponse tornjakTypes.ClusterInfoList

// ListClusters returns list of clusters from the local DB with the following info
// name string
// details json
func (s *Server) ListClusters(inp ListClustersRequest) (*ListClustersResponse, error) {
	log.Println("ListClusters: Fetching clusters from the database.")
	retVal, err := s.Db.GetClusters()
	if err != nil {
		log.Printf("ListClusters: Error fetching clusters - %v", err)
		return nil, err
	}
	return (*ListClustersResponse)(&retVal), nil
}

type RegisterClusterRequest tornjakTypes.ClusterInput

// DefineCluster registers cluster to local DB
func (s *Server) DefineCluster(inp RegisterClusterRequest) error {
	log.Println("DefineCluster: Registering cluster to the database.")
	cinfo := tornjakTypes.ClusterInfo(inp.ClusterInstance)

	// Check mandatory fields
	if len(cinfo.Name) == 0 {
		return errors.New("DefineCluster: Missing mandatory field - Name")
	} else if len(cinfo.PlatformType) == 0 {
		return errors.New("DefineCluster: Missing mandatory field - PlatformType")
	}

	// Generate a new UID if it's not provided
	if len(cinfo.UID) == 0 {
		newUID, err := uuid.NewUUID()
		if err != nil {
			return errors.New("DefineCluster: Failed to generate UID")
		}
		cinfo.UID = newUID.String()
		log.Printf("DefineCluster: Generated new UID for cluster - %s", cinfo.UID)
	}

	err := s.Db.CreateClusterEntry(cinfo)
	if err != nil {
		log.Printf("DefineCluster: Error creating cluster entry - %v", err)
	}
	return err
}

type EditClusterRequest tornjakTypes.ClusterInput

// EditCluster registers cluster to local DB
func (s *Server) EditCluster(inp EditClusterRequest) error {
	log.Println("EditCluster: Editing cluster in the database.")
	cinfo := tornjakTypes.ClusterInfo(inp.ClusterInstance)

	// Check mandatory fields
	if len(cinfo.Name) == 0 {
		return errors.New("EditCluster: Missing mandatory field - Name")
	} else if len(cinfo.PlatformType) == 0 {
		return errors.New("EditCluster: Missing mandatory field - PlatformType")
	} else if len(cinfo.UID) == 0 {
		return errors.New("EditCluster: Missing mandatory field - UID")
	}

	err := s.Db.EditClusterEntry(cinfo)
	if err != nil {
		log.Printf("EditCluster: Error editing cluster entry - %v", err)
	}
	return err
}

type DeleteClusterRequest tornjakTypes.ClusterInput

// DeleteCluster deletes cluster with name cinfo.Name and assignment to agents
func (s *Server) DeleteCluster(inp DeleteClusterRequest) error {
	log.Println("DeleteCluster: Deleting cluster from the database.")
	cinfo := tornjakTypes.ClusterInfo(inp.ClusterInstance)

	if len(cinfo.UID) == 0 {
		return errors.New("DeleteCluster: Missing mandatory field - UID")
	}

	err := s.Db.DeleteClusterEntry(cinfo.UID)
	if err != nil {
		log.Printf("DeleteCluster: Error deleting cluster entry - %v", err)
	}
	return err
}
