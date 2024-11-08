package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	managerdb "github.com/spiffe/tornjak/pkg/manager/db"
)

const (
	keyShowLen  int = 40
	certShowLen int = 50
)

type Server struct {
	listenAddr string
	db         managerdb.ManagerDB
}

// Handle preflight checks
func corsHandler(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			cors(w, r)
			return
		} else {
			f(w, r)
		}
	}
}

func cors(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
	w.WriteHeader(http.StatusOK)
}

func retError(w http.ResponseWriter, emsg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
	http.Error(w, emsg, status)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// SPIRE Health Check Handler
func (s *Server) spireHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["server"]

	// Retrieve server info
	sinfo, err := s.db.GetServer(serverName)
	if err != nil {
		retError(w, fmt.Sprintf("Error retrieving server info: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// Create HTTP client
	client, err := sinfo.HttpClient()
	if err != nil {
		retError(w, fmt.Sprintf("Error creating HTTP client: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// Make API request to SPIRE server
	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(sinfo.Address, "/")+"/api/v1/spire/healthcheck", nil)
	if err != nil {
		retError(w, fmt.Sprintf("Error creating HTTP request: %v", err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		retError(w, fmt.Sprintf("Error calling SPIRE health check: %v", err.Error()), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers and body
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		retError(w, fmt.Sprintf("Error copying response body: %v", err.Error()), http.StatusInternalServerError)
	}
}

// SPIRE Refresh Rate Handler
func (s *Server) spireRefreshRateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["server"]

	// Parse the refresh rate from the request body
	var refreshRate struct {
		Interval int `json:"interval"` // in seconds
	}
	if err := json.NewDecoder(r.Body).Decode(&refreshRate); err != nil {
		retError(w, fmt.Sprintf("Invalid JSON input: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// Simulate saving refresh rate (can be stored in DB or server config)
	fmt.Printf("Server: %s, New Refresh Rate: %d seconds\n", serverName, refreshRate.Interval)

	// Respond to the client
	cors(w, r)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success"}`))
}

// Serve static files for SPA
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Prepend static path to the request path
	path = filepath.Join(h.staticPath, path)

	// Check if file exists
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// Register API endpoints
func (s *Server) HandleRequests() {
	rtr := mux.NewRouter()

	// Manager-specific endpoints
	rtr.HandleFunc("/manager-api/server/list", corsHandler(s.serverList))
	rtr.HandleFunc("/manager-api/server/register", corsHandler(s.serverRegister))

	// SPIRE-specific endpoints
	rtr.HandleFunc("/manager-api/spire/health/{server:.*}", corsHandler(s.spireHealthCheckHandler)).Methods(http.MethodGet)
	rtr.HandleFunc("/manager-api/spire/refresh-rate/{server:.*}", corsHandler(s.spireRefreshRateHandler)).Methods(http.MethodPost)

	// Static SPA handler
	spa := spaHandler{staticPath: "ui-manager", indexPath: "index.html"}
	rtr.PathPrefix("/").Handler(spa)

	fmt.Println("Server listening on", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, rtr))
}

// NewManagerServer initializes a new server with DB
func NewManagerServer(listenAddr, dbString string) (*Server, error) {
	db, err := managerdb.NewLocalSqliteDB(dbString)
	if err != nil {
		return nil, err
	}
	return &Server{
		listenAddr: listenAddr,
		db:         db,
	}, nil
}

func (s *Server) serverList(w http.ResponseWriter, r *http.Request) {
	// Simulated server list handler
	fmt.Println("Endpoint Hit: Server List")
	cors(w, r)
	_, _ = w.Write([]byte(`{"servers":["server1","server2"]}`))
}

func (s *Server) serverRegister(w http.ResponseWriter, r *http.Request) {
	// Simulated server registration handler
	fmt.Println("Endpoint Hit: Server Register")
	cors(w, r)
	_, _ = w.Write([]byte(`{"status":"registered"}`))
}

