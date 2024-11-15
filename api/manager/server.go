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
		if r.Method == http.MethodOptions {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			w.WriteHeader(http.StatusOK)
			return
		}
		f(w, r)
	}
}

func retError(w http.ResponseWriter, emsg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, emsg)))
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
	if serverName == "" {
		retError(w, "Server name not provided", http.StatusBadRequest)
		return
	}

	sinfo, err := s.db.GetServer(serverName)
	if err != nil {
		retError(w, fmt.Sprintf("Error retrieving server info: %v", err), http.StatusBadRequest)
		return
	}

	client, err := sinfo.HttpClient()
	if err != nil {
		retError(w, fmt.Sprintf("Error creating HTTP client: %v", err), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(sinfo.Address, "/")+"/api/v1/spire/healthcheck", nil)
	if err != nil {
		retError(w, fmt.Sprintf("Error creating HTTP request: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		retError(w, fmt.Sprintf("Error calling SPIRE health check: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		retError(w, fmt.Sprintf("Error copying response body: %v", err), http.StatusInternalServerError)
	}
}

// SPIRE Refresh Rate Handler
func (s *Server) spireRefreshRateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["server"]
	if serverName == "" {
		retError(w, "Server name not provided", http.StatusBadRequest)
		return
	}

	var refreshRate struct {
		Interval int `json:"interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&refreshRate); err != nil {
		retError(w, fmt.Sprintf("Invalid JSON input: %v", err), http.StatusBadRequest)
		return
	}

	if refreshRate.Interval <= 0 {
		retError(w, "Invalid refresh rate: must be greater than zero", http.StatusBadRequest)
		return
	}

	log.Printf("Server: %s, New Refresh Rate: %d seconds", serverName, refreshRate.Interval)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success"}`))
}

// Serve static files for SPA
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticPath, r.URL.Path)
	path, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// Register API endpoints
func (s *Server) HandleRequests() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/manager-api/server/list", corsHandler(s.serverList))
	rtr.HandleFunc("/manager-api/server/register", corsHandler(s.serverRegister))
	rtr.HandleFunc("/manager-api/spire/health/{server:.*}", corsHandler(s.spireHealthCheckHandler)).Methods(http.MethodGet)
	rtr.HandleFunc("/manager-api/spire/refresh-rate/{server:.*}", corsHandler(s.spireRefreshRateHandler)).Methods(http.MethodPost)

	spa := spaHandler{staticPath: "ui-manager", indexPath: "index.html"}
	rtr.PathPrefix("/").Handler(spa)

	log.Printf("Server listening on %s", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, rtr))
}

// NewManagerServer initializes a new server with DB
func NewManagerServer(listenAddr, dbString string) (*Server, error) {
	db, err := managerdb.NewLocalSqliteDB(dbString)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	return &Server{
		listenAddr: listenAddr,
		db:         db,
	}, nil
}

func (s *Server) serverList(w http.ResponseWriter, r *http.Request) {
	log.Println("Endpoint Hit: Server List")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"servers":["server1","server2"]}`))
}

func (s *Server) serverRegister(w http.ResponseWriter, r *http.Request) {
	log.Println("Endpoint Hit: Server Register")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"registered"}`))
}


