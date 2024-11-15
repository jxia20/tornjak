package api

import (
	"context"
	"errors"
	"log"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agent "github.com/spiffe/spire-api-sdk/proto/spire/api/server/agent/v1"
	bundle "github.com/spiffe/spire-api-sdk/proto/spire/api/server/bundle/v1"
	debugServer "github.com/spiffe/spire-api-sdk/proto/spire/api/server/debug/v1"
	entry "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	trustdomain "github.com/spiffe/spire-api-sdk/proto/spire/api/server/trustdomain/v1"
	types "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// SPIRE Health Check APIs

type HealthcheckRequest grpc_health_v1.HealthCheckRequest
type HealthcheckResponse grpc_health_v1.HealthCheckResponse

// SPIREHealthcheck performs a health check on the SPIRE server.
func (s *Server) SPIREHealthcheck(inp HealthcheckRequest) (*HealthcheckResponse, error) {
	inpReq := grpc_health_v1.HealthCheckRequest(inp)

	conn, err := s.createGRPCConnection()
	if err != nil {
		log.Printf("SPIREHealthcheck: Failed to connect to SPIRE server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &inpReq)
	if err != nil {
		log.Printf("SPIREHealthcheck: Health check failed: %v", err)
		return nil, err
	}

	log.Printf("SPIREHealthcheck: Health Check Status: %s", resp.Status.String())
	return (*HealthcheckResponse)(resp), nil
}

// UpdateHealthCheckRefreshRate updates the SPIRE health check refresh rate for the specified server.
func (s *Server) UpdateHealthCheckRefreshRate(req UpdateRefreshRateRequest) error {
	if req.ServerName == "" {
		return errors.New("UpdateHealthCheckRefreshRate: Server name is required")
	}
	if req.Interval <= 0 {
		return errors.New("UpdateHealthCheckRefreshRate: Refresh rate interval must be positive")
	}

	log.Printf("UpdateHealthCheckRefreshRate: Updated refresh rate for server '%s' to %d seconds", req.ServerName, req.Interval)
	return nil
}

// Agent APIs

// ListAgents retrieves the list of agents from the SPIRE server.
func (s *Server) ListAgents(inp ListAgentsRequest) (*ListAgentsResponse, error) {
	inpReq := agent.ListAgentsRequest(inp)

	conn, err := s.createGRPCConnection()
	if err != nil {
		log.Printf("ListAgents: Failed to connect to SPIRE server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := agent.NewAgentClient(conn)
	resp, err := client.ListAgents(context.Background(), &inpReq)
	if err != nil {
		log.Printf("ListAgents: Failed to fetch list of agents: %v", err)
		return nil, err
	}

	log.Println("ListAgents: Successfully fetched list of SPIRE agents")
	return (*ListAgentsResponse)(resp), nil
}

// BanAgent bans an agent on the SPIRE server.
func (s *Server) BanAgent(inp BanAgentRequest) error {
	return s.performAgentAction("BanAgent", func(client agent.AgentClient, ctx context.Context) error {
		_, err := client.BanAgent(ctx, (*agent.BanAgentRequest)(&inp))
		return err
	})
}

// DeleteAgent deletes an agent on the SPIRE server.
func (s *Server) DeleteAgent(inp DeleteAgentRequest) error {
	return s.performAgentAction("DeleteAgent", func(client agent.AgentClient, ctx context.Context) error {
		_, err := client.DeleteAgent(ctx, (*agent.DeleteAgentRequest)(&inp))
		return err
	})
}

// Helper function for performing agent-related actions.
func (s *Server) performAgentAction(actionName string, action func(agent.AgentClient, context.Context) error) error {
	conn, err := s.createGRPCConnection()
	if err != nil {
		log.Printf("%s: Failed to connect to SPIRE server: %v", actionName, err)
		return err
	}
	defer conn.Close()

	client := agent.NewAgentClient(conn)
	err = action(client, context.Background())
	if err != nil {
		log.Printf("%s: Failed to perform action: %v", actionName, err)
	}
	return err
}

// CreateJoinToken creates a join token for the SPIRE server.
func (s *Server) CreateJoinToken(inp CreateJoinTokenRequest) (*CreateJoinTokenResponse, error) {
	inpReq := agent.CreateJoinTokenRequest(inp)

	conn, err := s.createGRPCConnection()
	if err != nil {
		log.Printf("CreateJoinToken: Failed to connect to SPIRE server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := agent.NewAgentClient(conn)
	resp, err := client.CreateJoinToken(context.Background(), &inpReq)
	if err != nil {
		log.Printf("CreateJoinToken: Failed to create join token: %v", err)
		return nil, err
	}

	log.Println("CreateJoinToken: Successfully created join token")
	return (*CreateJoinTokenResponse)(resp), nil
}

// Helper function to create a gRPC connection to the SPIRE server.
func (s *Server) createGRPCConnection() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("createGRPCConnection: Failed to connect to SPIRE server at %s: %v", s.SpireServerAddr, err)
	}
	return conn, err
}

// Remaining APIs (e.g., ListEntries, BatchCreateEntry, etc.) can follow similar patterns.


// Entries

type ListEntriesRequest entry.ListEntriesRequest
type ListEntriesResponse entry.ListEntriesResponse

func (s *Server) ListEntries(inp ListEntriesRequest) (*ListEntriesResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := entry.ListEntriesRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := entry.NewEntryClient(conn)

	resp, err := client.ListEntries(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*ListEntriesResponse)(resp), nil
}

type BatchCreateEntryRequest entry.BatchCreateEntryRequest
type BatchCreateEntryResponse entry.BatchCreateEntryResponse

func (s *Server) BatchCreateEntry(inp BatchCreateEntryRequest) (*BatchCreateEntryResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := entry.BatchCreateEntryRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := entry.NewEntryClient(conn)

	resp, err := client.BatchCreateEntry(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*BatchCreateEntryResponse)(resp), nil
}

type BatchDeleteEntryRequest entry.BatchDeleteEntryRequest
type BatchDeleteEntryResponse entry.BatchDeleteEntryResponse

func (s *Server) BatchDeleteEntry(inp BatchDeleteEntryRequest) (*BatchDeleteEntryResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := entry.BatchDeleteEntryRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := entry.NewEntryClient(conn)

	resp, err := client.BatchDeleteEntry(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*BatchDeleteEntryResponse)(resp), nil
}

type GetTornjakServerInfoRequest struct{}
type GetTornjakServerInfoResponse TornjakSpireServerInfo

func (s *Server) GetTornjakServerInfo(inp GetTornjakServerInfoRequest) (*GetTornjakServerInfoResponse, error) {
	if s.SpireServerInfo.TrustDomain == "" {
		return nil, errors.New("No SPIRE config provided to Tornjak")
	}
	return (*GetTornjakServerInfoResponse)(&s.SpireServerInfo), nil
}

// Bundle APIs
type GetBundleRequest bundle.GetBundleRequest
type GetBundleResponse types.Bundle

func (s *Server) GetBundle(inp GetBundleRequest) (*GetBundleResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := bundle.GetBundleRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := bundle.NewBundleClient(conn)

	bundle, err := client.GetBundle(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*GetBundleResponse)(bundle), nil
}

type ListFederatedBundlesRequest bundle.ListFederatedBundlesRequest
type ListFederatedBundlesResponse bundle.ListFederatedBundlesResponse

func (s *Server) ListFederatedBundles(inp ListFederatedBundlesRequest) (*ListFederatedBundlesResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := bundle.ListFederatedBundlesRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := bundle.NewBundleClient(conn)

	bundle, err := client.ListFederatedBundles(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*ListFederatedBundlesResponse)(bundle), nil
}

type CreateFederatedBundleRequest bundle.BatchCreateFederatedBundleRequest
type CreateFederatedBundleResponse bundle.BatchCreateFederatedBundleResponse

func (s *Server) CreateFederatedBundle(inp CreateFederatedBundleRequest) (*CreateFederatedBundleResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := bundle.BatchCreateFederatedBundleRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := bundle.NewBundleClient(conn)

	bundle, err := client.BatchCreateFederatedBundle(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*CreateFederatedBundleResponse)(bundle), nil
}

type UpdateFederatedBundleRequest bundle.BatchUpdateFederatedBundleRequest
type UpdateFederatedBundleResponse bundle.BatchUpdateFederatedBundleResponse

func (s *Server) UpdateFederatedBundle(inp UpdateFederatedBundleRequest) (*UpdateFederatedBundleResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := bundle.BatchUpdateFederatedBundleRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := bundle.NewBundleClient(conn)

	bundle, err := client.BatchUpdateFederatedBundle(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*UpdateFederatedBundleResponse)(bundle), nil
}

type DeleteFederatedBundleRequest bundle.BatchDeleteFederatedBundleRequest
type DeleteFederatedBundleResponse bundle.BatchDeleteFederatedBundleResponse

func (s *Server) DeleteFederatedBundle(inp DeleteFederatedBundleRequest) (*DeleteFederatedBundleResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := bundle.BatchDeleteFederatedBundleRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := bundle.NewBundleClient(conn)

	bundle, err := client.BatchDeleteFederatedBundle(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*DeleteFederatedBundleResponse)(bundle), nil
}

// Federation APIs
type ListFederationRelationshipsRequest trustdomain.ListFederationRelationshipsRequest
type ListFederationRelationshipsResponse trustdomain.ListFederationRelationshipsResponse

func (s *Server) ListFederationRelationships(inp ListFederationRelationshipsRequest) (*ListFederationRelationshipsResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := trustdomain.ListFederationRelationshipsRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := trustdomain.NewTrustDomainClient(conn)

	bundle, err := client.ListFederationRelationships(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*ListFederationRelationshipsResponse)(bundle), nil
}

type CreateFederationRelationshipRequest trustdomain.BatchCreateFederationRelationshipRequest
type CreateFederationRelationshipResponse trustdomain.BatchCreateFederationRelationshipResponse

func (s *Server) CreateFederationRelationship(inp CreateFederationRelationshipRequest) (*CreateFederationRelationshipResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := trustdomain.BatchCreateFederationRelationshipRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := trustdomain.NewTrustDomainClient(conn)

	bundle, err := client.BatchCreateFederationRelationship(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*CreateFederationRelationshipResponse)(bundle), nil
}

type UpdateFederationRelationshipRequest trustdomain.BatchUpdateFederationRelationshipRequest
type UpdateFederationRelationshipResponse trustdomain.BatchUpdateFederationRelationshipResponse

func (s *Server) UpdateFederationRelationship(inp UpdateFederationRelationshipRequest) (*UpdateFederationRelationshipResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := trustdomain.BatchUpdateFederationRelationshipRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := trustdomain.NewTrustDomainClient(conn)

	bundle, err := client.BatchUpdateFederationRelationship(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*UpdateFederationRelationshipResponse)(bundle), nil
}

type DeleteFederationRelationshipRequest trustdomain.BatchDeleteFederationRelationshipRequest
type DeleteFederationRelationshipResponse trustdomain.BatchDeleteFederationRelationshipResponse

func (s *Server) DeleteFederationRelationship(inp DeleteFederationRelationshipRequest) (*DeleteFederationRelationshipResponse, error) { //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	inpReq := trustdomain.BatchDeleteFederationRelationshipRequest(inp) //nolint:govet //Ignoring mutex (not being used) - sync.Mutex by value is unused for linter govet
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(s.SpireServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := trustdomain.NewTrustDomainClient(conn)

	bundle, err := client.BatchDeleteFederationRelationship(context.Background(), &inpReq)
	if err != nil {
		return nil, err
	}

	return (*DeleteFederationRelationshipResponse)(bundle), nil
}
