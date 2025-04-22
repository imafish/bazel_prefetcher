package httpserver

import (
	"encoding/json"
	"internal/common"
	"net/http"
	"strconv"
	"sync"
)

func bazelCommandsGetList(commands *[][]string, mtx *sync.Mutex) http.HandlerFunc {
	l := common.NewLoggerWithPrefixAndColor("restful_server.ServeApiV1BazelCommands: ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		l.Printf("Received request for Bazel command list")
		w.Header().Set("Content-Type", "application/json")
		mtx.Lock()
		defer mtx.Unlock()
		if err := json.NewEncoder(w).Encode(*commands); err != nil {
			l.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func bazelCommandsGetOne(commands *[][]string, mtx *sync.Mutex) http.HandlerFunc {
	l := common.NewLoggerWithPrefixAndColor("restful_server.ServeApiV1BazelCommands: ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		l.Printf("Received request for Bazel command")
		mtx.Lock()
		defer mtx.Unlock()

		index := r.PathValue("index")
		if index == "" {
			http.Error(w, "Index parameter is required", http.StatusBadRequest)
			return
		}

		idx, err := strconv.Atoi(index)
		if err != nil || idx < 0 || idx >= len(*commands) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode((*commands)[idx]); err != nil {
			l.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func bazelCommandsDelete(commands *[][]string, mtx *sync.Mutex) http.HandlerFunc {
	l := common.NewLoggerWithPrefixAndColor("restful_server.ServeApiV1BazelCommands: ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		l.Printf("Received request to delete Bazel command list")
		mtx.Lock()
		defer mtx.Unlock()

		index := r.PathValue("index")
		if index == "" {
			http.Error(w, "Index parameter is required", http.StatusBadRequest)
			return
		}

		idx, err := strconv.Atoi(index)
		if err != nil || idx < 0 || idx >= len(*commands) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		*commands = append((*commands)[:idx], (*commands)[idx+1:]...)
		w.WriteHeader(http.StatusNoContent)
	}
}

func bazelCommandsPut(commands *[][]string, mtx *sync.Mutex) http.HandlerFunc {
	l := common.NewLoggerWithPrefixAndColor("restful_server.ServeApiV1BazelCommands: ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		l.Printf("Received request to update Bazel command list")
		mtx.Lock()
		defer mtx.Unlock()

		index := r.PathValue("index")
		if index == "" {
			http.Error(w, "Index parameter is required", http.StatusBadRequest)
			return
		}

		idx, err := strconv.Atoi(index)
		if err != nil || idx < 0 || idx >= len(*commands) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		var updatedCommand []string
		if err := json.NewDecoder(r.Body).Decode(&updatedCommand); err != nil {
			l.Printf("Error decoding request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		(*commands)[idx] = updatedCommand

		w.WriteHeader(http.StatusNoContent)
	}
}

func bazelCommandsNew(commands *[][]string, mtx *sync.Mutex) http.HandlerFunc {
	l := common.NewLoggerWithPrefixAndColor("restful_server.ServeApiV1BazelCommands: ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		l.Printf("Received request to add new Bazel command")
		mtx.Lock()
		defer mtx.Unlock()

		var newCommand []string
		if err := json.NewDecoder(r.Body).Decode(&newCommand); err != nil {
			l.Printf("Error decoding request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// check for duplicates
		for _, cmd := range *commands {
			if len(cmd) == len(newCommand) {
				for i, j := range cmd {
					if j != newCommand[i] {
						break
					}
					if i == len(cmd)-1 {
						l.Printf("Duplicate command found")
						force := r.URL.Query().Get("f")
						if force == "" {
							l.Printf("Duplicate command found, but not forced to add")
							http.Error(w, "Duplicate command", http.StatusBadRequest)
							return
						} else {
							l.Printf("Duplicate command found, but forced to add")
						}
					}
				}
			}
		}

		*commands = append(*commands, newCommand)

		w.WriteHeader(http.StatusNoContent)
	}
}
