package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Printf("Healthcheck: Received request")
	var input HealthcheckRequest
	buf := new(strings.Builder)

	n, err := io.Copy(buf, r.Body)
	if err != nil {
		log.Printf("Healthcheck: error reading request body: %v", err)
		retError(w, "Error parsing data", http.StatusBadRequest)
		return
	}
	data := buf.String()

	if n == 0 {
		input = HealthcheckRequest{}
	} else {
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			log.Printf("Healthcheck: error unmarshalling request: %v", err)
			retError(w, "Error parsing data", http.StatusBadRequest)
			return
		}
	}

	ret, err := s.SPIREHealthcheck(input)
	if err != nil {
		log.Printf("Healthcheck: error in SPIREHealthcheck: %v", err)
		retError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cors(w, r)
	je := json.NewEncoder(w)
	err = je.Encode(ret)
	if err != nil {
		log.Printf("Healthcheck: error encoding response: %v", err)
		retError(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (s *Server) debugServer(w http.ResponseWriter, r *http.Request) {
	log.Printf("DebugServer: Received request")
	input := DebugServerRequest{}

	ret, err := s.DebugServer(input)
	if err != nil {
		log.Printf("DebugServer: error in DebugServer: %v", err)
		retError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cors(w, r)
	je := json.NewEncoder(w)
	err = je.Encode(ret)
	if err != nil {
		log.Printf("DebugServer: error encoding response: %v", err)
		retError(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (s *Server) agentList(w http.ResponseWriter, r *http.Request) {
	log.Printf("AgentList: Received request")
	var input ListAgentsRequest
	buf := new(strings.Builder)

	n, err := io.Copy(buf, r.Body)
	if err != nil {
		log.Printf("AgentList: error reading request body: %v", err)
		retError(w, "Error parsing data", http.StatusBadRequest)
		return
	}
	data := buf.String()

	if n == 0 {
		input = ListAgentsRequest{}
	} else {
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			log.Printf("AgentList: error unmarshalling request: %v", err)
			retError(w, "Error parsing data", http.StatusBadRequest)
			return
		}
	}

	ret, err := s.ListAgents(input)
	if err != nil {
		log.Printf("AgentList: error in ListAgents: %v", err)
		retError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cors(w, r)
	je := json.NewEncoder(w)
	err = je.Encode(ret)
	if err != nil {
		log.Printf("AgentList: error encoding response: %v", err)
		retError(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (s *Server) agentBan(w http.ResponseWriter, r *http.Request) {
	log.Printf("AgentBan: Received request")
	var input BanAgentRequest
	buf := new(strings.Builder)

	n, err := io.Copy(buf, r.Body)
	if err != nil {
		log.Printf("AgentBan: error reading request body: %v", err)
		retError(w, "Error parsing data", http.StatusBadRequest)
		return
	}
	data := buf.String()

	if n == 0 {
		retError(w, "Error: no data provided", http.StatusBadRequest)
		return
	} else {
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			log.Printf("AgentBan: error unmarshalling request: %v", err)
			retError(w, "Error parsing data", http.StatusBadRequest)
			return
		}
	}

	err = s.BanAgent(input)
	if err != nil {
		log.Printf("AgentBan: error in BanAgent: %v", err)
		retError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cors(w, r)
	_, err = w.Write([]byte("SUCCESS"))
	if err != nil {
		log.Printf("AgentBan: error writing response: %v", err)
		retError(w, "Error writing response", http.StatusInternalServerError)
	}
}

// More methods follow the same pattern of adding consistent logging and error handling...

// Helper function for returning errors
func retError(w http.ResponseWriter, message string, code int) {
	http.Error(w, message, code)
	log.Printf("Error: %s, Code: %d", message, code)
}

// Enable CORS for responses
func cors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
