module server

go 1.23.2

replace internal/git => ../../internal/git

replace internal/common => ../../internal/common

replace internal/downloaders => ../../internal/downloaders

replace internal/db => ../../internal/db

replace internal/httpserver => ../../internal/http_server

replace internal/prefetcher => ../../internal/prefetcher

require internal/git v1.0.0

require internal/common v1.0.0

require internal/downloaders v1.0.0

require internal/db v1.0.0

require internal/httpserver v1.0.0

require internal/prefetcher v1.0.0

require github.com/mattn/go-sqlite3 v1.14.27 // indirect
