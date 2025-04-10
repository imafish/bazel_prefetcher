
# PLAN

## step 1

1. server
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
2. client
    1. [x] analyze (from config) download request
    2. [x] fetch from server, and put it in cache

## step 2

1. server
   1. [ ] find dependencies from bazel output
   2. [ ] downloader API

# Minor improvements

1. [ ] calculate hash of downloaded file
2. [ ] single line separator
3. [ ] improved logging
4. [ ] internal/db/item.go: remove sqlite specific query
5. [ ] internal/db/item.go: the table should have more constraints -- e.g. Url should be unique.
