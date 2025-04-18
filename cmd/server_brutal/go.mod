module serverbrutal

go 1.24.2

replace internal/git => ../../internal/git

replace internal/common => ../../internal/common

replace internal/downloaders => ../../internal/downloaders

replace internal/db => ../../internal/db

replace internal/httpserver => ../../internal/http_server

replace internal/prefetcher => ../../internal/prefetcher

replace internal/cleanup => ../../internal/cleanup

require internal/git v1.0.0

require internal/common v1.0.0

require internal/db v1.0.0 // indirect

require internal/httpserver v1.0.0

require internal/cleanup v1.0.0

require github.com/mattn/go-sqlite3 v1.14.27 // indirect
