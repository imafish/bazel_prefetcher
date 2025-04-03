
# PLAN 1

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

# PLAN 2

## step 1

  1. [x] server_brutal
  2. [x] client_brutal

## step 2
  1. server_brutal
    1. [ ] more targets
      1. [ ] coverage
      2. [ ] coverity
      3. [ ] tests
      4. [ ] clang-tidy
      5. [ ] git-hook
    2. [ ] clean (cache items not used for xx days)
    3. [ ] configurable bazel targets.
    4. [ ] restapi to manage bazel targets.
  2. client_brutal
    1. [ ] clean (cache items not used for xx days)

# PLAN 3

1. bazel-remote, content-addressable cache

# Minor improvements

1. [x] calculate hash of downloaded file
2. [x] single line separator
3. [x] improved logging (for HTML server)
4. [ ] improved logging (others)
4. [ ] internal/db/item.go: remove sqlite specific query
5. [ ] internal/db/item.go: the table should have more constraints -- e.g. Url should be unique.
