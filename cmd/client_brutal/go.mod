module clientbrutal

go 1.23.2

replace internal/git => ../../internal/git

replace internal/common => ../../internal/common

replace internal/downloaders => ../../internal/downloaders

replace internal/db => ../../internal/db

replace internal/httpserver => ../../internal/http_server

replace internal/prefetcher => ../../internal/prefetcher

require internal/common v1.0.0
