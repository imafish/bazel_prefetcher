
# PLAN

## step 1

1. server
    1. [x] init (load config)
    2. [x] fetch and update src
       1. [ ] hardcoded
       2. [x] analyzer by regex
       3. [ ] analyzer by regex and anchor
    3. [x] download
    4. [x] provide a file server
    5. [x] repeat/schedule (in period of time)
    6. [x] provide file server
2. client
    1. [ ] analyze (from config) download request
    2. [ ] fetch from server, and put it in cache

## step 2

1. server
   1. [ ] find dependencies from bazel output
   2. [ ] analyzer that support params
   3. [ ] downloader API

# Minor improvements

1. calculate hash of downloaded file
2. internal/db/item.go: remove explicit sqlite query
3. internal/db/item.go: the table should have more constraints -- e.g. Url should be unique.
4. better logger
