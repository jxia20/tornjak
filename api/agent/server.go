package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"

	"github.com/spiffe/tornjak/pkg/agent/authentication/authenticator"
	"github.com/spiffe/tornjak/pkg/agent/authorization"
	agentdb "github.com/spiffe/tornjak/pkg/agent/db"
)

type Server struct {
	SpireServerAddr string
	SpireServerInfo TornjakSpireServerInfo
	TornjakConfig   *TornjakConfig
	Db              agentdb.AgentDB
	Authenticator   authenticator.Authenticator
	Authorizer      authorization.Authorizer
}

type hclPluginConfig struct {
	PluginCmd      string   `hcl:"plugin_cmd"`
	PluginArgs     []string `hcl:"plugin_args"`
	PluginChecksum string   `hcl:"plugin_checksum"`
	PluginData     ast.Node `hcl:"plugin_data"`
	Enabled        *bool    `hcl:"enabled"`
}

func cors(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

func retError(w http.ResponseWriter, emsg string, status int) {
	log.Printf("HTTP %d - %s", status, emsg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Error(w, emsg, status)
}

func (s *Server) verificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			cors(w, r)
			return
		}

		userInfo := s.Authenticator.AuthenticateRequest(r)
		err := s.Authorizer.AuthorizeRequest(r, userInfo)
		if err != nil {
			log.Printf("Unauthorized access attempt: %v", err)
			retError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	relPath := filepath.Clean(r.URL.Path)
	absPath := filepath.Join(h.staticPath, relPath)

	if !strings.HasPrefix(absPath, h.staticPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	_, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func (s *Server) HandleRequests() {
	err := s.Configure()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	errChannel := make(chan error, 2)
	serverConfig := s.TornjakConfig.Server

	httpHandler := s.GetRouter()

	if serverConfig.HTTPSConfig == nil {
		log.Println("WARNING: HTTPS is not configured. Running insecure HTTP server.")
	} else {
		go func() {
			tlsConfig, err := serverConfig.HTTPSConfig.Parse()
			if err != nil {
				log.Printf("Error parsing HTTPS config: %v", err)
				errChannel <- err
				return
			}

			addr := fmt.Sprintf(":%d", serverConfig.HTTPSConfig.ListenPort)
			log.Printf("Starting HTTPS server on %s...", addr)
			server := &http.Server{
				Handler:   s.GetRouter(),
				Addr:      addr,
				TLSConfig: tlsConfig,
			}
			errChannel <- server.ListenAndServeTLS(serverConfig.HTTPSConfig.Cert, serverConfig.HTTPSConfig.Key)
		}()
	}

	go func() {
		addr := fmt.Sprintf(":%d", serverConfig.HTTPConfig.ListenPort)
		log.Printf("Starting HTTP server on %s...", addr)
		errChannel <- http.ListenAndServe(addr, httpHandler)
	}()

	for i := 0; i < 2; i++ {
		err := <-errChannel
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}
}
