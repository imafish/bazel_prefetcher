package httpserver

import (
	"fmt"
	"internal/common"
	"log"
	"net/http"
)

func StartServer(config *common.ServerConfig) {

	// serve `/file` requests
	serveFiles(config)

	// serve `/restapi/v1` request
	serveApiV1(config)

	// serve `/` requests
	// http.HandleFunc("/", http.NotFound)

	// Start server
	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Server started on http://localhost" + addr)
	log.Printf("Access files at http://localhost" + addr + "/files/path/to/your/file")
	http.ListenAndServe(addr, nil)
}
