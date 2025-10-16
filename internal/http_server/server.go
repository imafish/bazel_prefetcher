package httpserver

import (
	"fmt"
	"internal/common"
	"net/http"
	"sync"
	"time"
)

type HttpServerBuilder struct {
	config   *common.ServerConfig
	serveMux *http.ServeMux
}

func NewHttpServerBuilder(config *common.ServerConfig) *HttpServerBuilder {
	return &HttpServerBuilder{
		config:   config,
		serveMux: http.NewServeMux(),
	}
}

func (b *HttpServerBuilder) ServeFiles() *HttpServerBuilder {
	b.serveMux.HandleFunc("/files/", serveFiles(b.config))
	return b
}

func (b *HttpServerBuilder) ServeApiV1Files() *HttpServerBuilder {
	b.serveMux.HandleFunc("/restapi/v1/files", serveApiV1GetAllFiles(b.config))
	return b
}

func (b *HttpServerBuilder) ServeApiV1BazelCommands(bazelCommands *[][]string, mtx *sync.Mutex) *HttpServerBuilder {
	b.serveMux.HandleFunc("GET /restapi/v1/bazelcommands", bazelCommandsGetList(bazelCommands, mtx))
	b.serveMux.HandleFunc("GET /restapi/v1/bazelcommands/{index}", bazelCommandsGetOne(bazelCommands, mtx))
	b.serveMux.HandleFunc("DELETE /restapi/v1/bazelcommands", bazelCommandsDelete(bazelCommands, mtx))
	b.serveMux.HandleFunc("PUT /restapi/v1/bazelcommands", bazelCommandsPut(bazelCommands, mtx))
	b.serveMux.HandleFunc("POST /restapi/v1/bazelcommands", bazelCommandsNew(bazelCommands, mtx))
	return b
}

func (b *HttpServerBuilder) Build() *http.Server {
	return &http.Server{
		Addr:           fmt.Sprintf(":%d", b.config.Server.Port),
		Handler:        b.serveMux,
		IdleTimeout:    600 * time.Second, // b.config.Server.IdleTimeout,
		ReadTimeout:    600 * time.Second, // b.config.Server.ReadTimeout,
		WriteTimeout:   600 * time.Second, // b.config.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,           // b.config.Server.MaxHeaderBytes,
	}
}
