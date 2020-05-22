# Gitlab Logger
This allows to disable tail based Gitlab logging (by providing a no-op tail) and instead use a sidecar logger agent that tail files and output (semi-)structured logs (JSON) on stdout or, eventually, use it as a tail dropin replacement.

Contribution is more than welcomed!

[![GitHub license](https://img.shields.io/github/license/jbguerraz/gitlab-logger.svg?c)](https://github.com/jbguerraz/gitlab-logger/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jbguerraz/gitlab-logger?status.svg)](https://pkg.go.dev/github.com/jbguerraz/gitlab-logger?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbguerraz/gitlab-logger)](https://goreportcard.com/report/github.com/jbguerraz/gitlab-logger)
[![GitHub issues](https://img.shields.io/github/issues/jbguerraz/gitlab-logger.svg)](https://github.com/jbguerraz/gitlab-logger/issues)

## Configuration
Configuration is flag based. When used as a tail dropin replacement, use a wrapper script if you want to configure it:

for example, in `/usr/local/bin/tail`
```
#!/usr/bin/env sh
exec /usr/local/bin/gitlab-logger --poll=true --exclude="sasl|config|lock|@|gzip|tgz|gz|production.log" --minlevel=3 $@
```

### poll
Set it to true to use polling instead of inotify for directory watching (used when a directory is passed in the list of files to tail. e.g: /var/log/gitlab)
#### default
`false`
### json
Set it to true to keep only (json) structured event logs
#### default
`false`
### minlevel
Minimum log level, used to filter out logs:
- 1 debug
- 2 info
- 3 notice
- 4 warn
- 5 err
- 6 fatal
- 100 unknown (level is keyword detection based. some won't match)
#### default
`0`
### exclude
blacklist files based on their full path (used when watching a directory). list of keywords, using a pipe (`|`) as separator.
#### default
`sasl|config|lock|@|gzip|tgz|gz`

## How to give it a try
### Sidecar
`docker-compose -f docker-compose.sidecar.yml build && docker-compose -f docker-compose.sidecar.yml up -d && docker logs -f gitlab-logger_logger_1`

### Tail dropin replacement
`docker-compose -f docker-compose.dropin.yml build && docker-compose -f docker-compose.dropin.yml up -d && docker logs -f gitlab-logger_web_1`

## Example output
```
$ docker logs -f gitlab-logger_logger_1 | jq
{
  "date": "2020-05-22T00:47:17Z",
  "component": "gitaly",
  "subcomponent": "current",
  "level": "error",
  "file": "/var/log/gitlab/gitaly/current",
  "message": {
    "error": "open /var/opt/gitlab/gitaly/gitaly.pid: no such file or directory",
    "level": "error",
    "msg": "find gitaly",
    "time": "2020-05-22T00:47:17Z",
    "wrapper": 466
  }
}
{
  "date": "2020-05-22T00:47:17Z",
  "component": "gitaly",
  "subcomponent": "current",
  "level": "warning",
  "file": "/var/log/gitlab/gitaly/current",
  "message": "time=\"2020-05-22T00:47:17Z\" level=warning msg=\"git path not configured. Using default path resolution\" resolvedPath=/opt/gitlab/embedded/bin/git"
}
{
  "date": "2020-05-22T00:47:18Z",
  "component": "gitaly",
  "subcomponent": "current",
  "level": "warning",
  "file": "/var/log/gitlab/gitaly/current",
  "message": {
    "level": "warning",
    "msg": "spawned",
    "supervisor.args": [
      "bundle",
      "exec",
      "bin/ruby-cd",
      "/var/opt/gitlab/gitaly",
      "/opt/gitlab/embedded/service/gitaly-ruby/bin/gitaly-ruby",
      "479",
      "/var/opt/gitlab/gitaly/internal_sockets/ruby.0"
    ],
    "supervisor.name": "gitaly-ruby.0",
    "supervisor.pid": 498,
    "time": "2020-05-22T00:47:17.676Z"
  }
}
{
  "date": "2020-05-22T00:47:18Z",
  "component": "gitaly",
  "subcomponent": "current",
  "level": "warning",
  "file": "/var/log/gitlab/gitaly/current",
  "message": {
    "level": "warning",
    "msg": "spawned",
    "supervisor.args": [
      "bundle",
      "exec",
      "bin/ruby-cd",
      "/var/opt/gitlab/gitaly",
      "/opt/gitlab/embedded/service/gitaly-ruby/bin/gitaly-ruby",
      "479",
      "/var/opt/gitlab/gitaly/internal_sockets/ruby.1"
    ],
    "supervisor.name": "gitaly-ruby.1",
    "supervisor.pid": 500,
    "time": "2020-05-22T00:47:17.676Z"
  }
}
{
  "date": "2020-05-22T00:47:42Z",
  "component": "gitlab-rails",
  "subcomponent": "gitlab-rails-db-migrate-2020-05-22-00-47-26",
  "level": "notice",
  "file": "/var/log/gitlab/gitlab-rails/gitlab-rails-db-migrate-2020-05-22-00-47-26.log",
  "message": "psql:/opt/gitlab/embedded/service/gitlab-rails/db/structure.sql:3: NOTICE:  extension \"plpgsql\" already exists, skipping"
}
{
  "date": "2020-05-22T00:47:42Z",
  "component": "gitlab-rails",
  "subcomponent": "gitlab-rails-db-migrate-2020-05-22-00-47-26",
  "level": "notice",
  "file": "/var/log/gitlab/gitlab-rails/gitlab-rails-db-migrate-2020-05-22-00-47-26.log",
  "message": "psql:/opt/gitlab/embedded/service/gitlab-rails/db/structure.sql:5: NOTICE:  extension \"pg_trgm\" already exists, skipping"
}
```

## Watch it

[![asciicast](https://asciinema.org/a/lx8RDAWAAaMFXZNbiaseBRkDL.svg)](https://asciinema.org/a/lx8RDAWAAaMFXZNbiaseBRkDL)
