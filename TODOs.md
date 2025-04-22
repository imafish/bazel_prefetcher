
# PLAN 1

## step 1

### server

1. [x] init (load config)
2. [x] fetch and update src
    1. [x] hardcoded
    2. [x] analyzer by regex
    3. [x] analyzer by regex and anchor
    4. [ ] analyzer by params
3. [x] download
4. [x] provide a file server
5. [x] repeat/schedule (in period of time)
6. [x] provide file server
  
### client

1. [x] analyze (from config) download request
2. [x] fetch from server, and put it in cache

## step 2

### server

1. [ ] find dependencies from bazel output
2. [ ] downloader API

# PLAN 2

## step 1

1. [x] server_brutal
2. [x] client_brutal

## step 2
     
### server_brutal

1. [ ] more targets
    1. [ ] coverage
    2. [ ] coverity
    3. [ ] tests
    4. [ ] clang-tidy
    5. [ ] git-hook
2. [x] clean (cache items not used for xx days)
3. [x] configurable bazel targets.
4. [x] restapi to manage bazel targets.

### client_brutal

1. [ ] clean (cache items not used for xx days)
2. [ ] two-ways sync

## step 3

### server_brutal

1. [ ] single file download
    1. [ ] save cache information to db
        1. [ ] initial update
        2. [ ] periodically update
    2. [ ] api/v1/files gets all files (from db)
    3. [ ] api/v1/query_url query if server has the cache item for certain URL, and it's status.
        1. save downloading status to DB
    4. [ ] api/v1/request_download request the server to download the file
        1. [ ] warn if URL is not valid
        1. [ ] use downloaders to download the requested URL

### client_brutal

1. [ ] request the server to download a file
    1. [ ] download if server has the file (and put in bazel cache)
    2. [ ] request the server to download if it doesn't have it, and is not downloading.
        1. [ ] if --wait, wait and poll the server until download finishes.
        2. [ ] if --retry=N, retry N times if server fails
      

# PLAN 3

1. bazel-remote, content-addressable cache

# Minor improvements

1. [x] calculate hash of downloaded file
2. [x] single line separator
3. [x] improved logging (for HTML server)
4. [x] improved logging (others)
4. [ ] internal/db/item.go: remove sqlite specific query
5. [ ] internal/db/item.go: the table should have more constraints -- e.g. Url should be unique, indexes, etc.
