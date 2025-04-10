package httpserver

import (
	"internal/common"
	"net/http"
)

func serveApiV1(config *common.ServerConfig) {
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {

	})
}
