# BAZEL_PREFETCHER

## Prerequisites

1. git
2. src cloned at workspace
3. proper bazel version
4. a downloader (e.g. aria2c)

## PLAN

### step 1

1. server
    1. [ ] init (load config)
    2. [ ] fetch and update src
    3. [ ] analyze (hardcoded) download request
    4. [ ] download
    5. [ ] provide downloader API
    6. [ ] repeat/schedule (in period of time)
2. client
    1. [ ] analyze (hardcoded) download request
    2. [ ] fetch from server, and put it in cache
