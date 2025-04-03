module prefetcher

go 1.23.2

replace internal/common => ../../internal/common

replace internal/db => ../../internal/db

require internal/common v1.0.0

require (
	github.com/mattn/go-sqlite3 v1.14.27 // indirect
	internal/db v1.0.0 // indirect
)
