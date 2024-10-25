package api

import (
	"errors"

	tornjakTypes "github.com/spiffe/tornjak/pkg/agent/types"
)

// Existing types and functions...

type ListSelectorsRequest struct{}
type ListSelectorsResponse tornjakTypes.AgentInfoList

// ListSelectors returns list of selectors from the local DB
func (s *Server) ListSelectors(inp ListSelectorsRequest) (*ListSelectorsResponse, error) {
	resp, err := s.Db.GetAgentSelectors()
	if err != nil {
		return nil, err
	}
	return (*ListSelectorsResponse)(&resp), nil
}

type RegisterSelectorRequest tornjakTypes.AgentInfo

// DefineSelectors registers an agent to the local DB
func (s *Server) DefineSelectors(inp RegisterSelectorRequest) error {
	sinfo := tornjakTypes.AgentInfo(inp)
	if len(sinfo.Spiffeid) == 0 {
		return errors.New("agent's info missing mandatory field - Spiffeid")
	}
	return s.Db.CreateAgentEntry(sinfo)
}

type UpdateSelectorRequest tornjakTypes.AgentInfo

// UpdateSelectors updates an existing selector
func (s *Server) UpdateSelectors(inp UpdateSelectorRequest) error {
	sinfo := tornjakTypes.AgentInfo(inp)
	if len(sinfo.Spiffeid) == 0 {
		return errors.New("agent's info missing mandatory field - Spiffeid")
	}
	return s.Db.UpdateAgentEntry(sinfo) // Assume UpdateAgentEntry is implemented in your Db interface
}

type DeleteSelectorRequest struct {
	Spiffeid string `json:"spiffeid"` // Identifier for the selector to delete
}

// DeleteSelectors deletes a specified selector
func (s *Server) DeleteSelectors(inp DeleteSelectorRequest) error {
	if len(inp.Spiffeid) == 0 {
		return errors.New("input missing mandatory field - Spiffeid")
	}
	return s.Db.DeleteAgentEntry(inp.Spiffeid) // Assume DeleteAgentEntry is implemented in your Db interface
}

// Existing ListAgentMetadata, ListClusters, etc...
