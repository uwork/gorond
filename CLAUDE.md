# CLAUDE.md — AI Assistant Guide for gorond

## Project Overview

**gorond** is a Go-based cron daemon that reimplements traditional Unix cron with enhanced features:
- Error notifications via multiple channels (email, Slack, SNS, Fluentd, stdout)
- Sequential and parallel multi-command execution in a single job definition
- Hot config reloading via file system watching
- REST Web API for external monitoring of job statuses

**Version**: 1.0.1
**Language**: Go
**Module path**: `github.com/uwork/gorond`
**License**: MIT

---

## Repository Structure

```
gorond/
├── main.go              # Entry point: CLI flags, startup orchestration
├── main_test.go         # Integration tests for daemon lifecycle
├── Makefile             # Build targets (test, build, deps, rpm, clean, run)
├── .travis.yml          # CI: vet → test → build → rpm → GitHub release
├── README.md            # User-facing documentation (Japanese)
│
├── goron/               # Core daemon logic
│   ├── goron.go         # Cron scheduling, job execution, signal handling, auto-reload
│   ├── goron_test.go    # Tests for scheduling and reload
│   ├── notify.go        # Notification dispatcher (routes to notify/* channels)
│   └── notify_test.go
│
├── config/              # Configuration parsing
│   ├── config.go        # Config struct definitions and LoadConfig()
│   ├── config_parser.go # INI-style parser for goron.conf
│   ├── job_parser.go    # Cron-syntax job file parser (handles sequential/parallel jobs)
│   ├── config_test.go
│   ├── config_parser_test.go
│   ├── job_parser_test.go
│   └── config_test.d/   # Fixture configs for tests
│
├── logging/             # Structured logger
│   ├── logger.go        # Logger with levels: DEBUG, INFO, WARN, ERROR, FATAL
│   └── logger_test.go
│
├── webapi/              # REST API server
│   ├── webapi.go        # HTTP server, /statuses and /jobs endpoints
│   ├── webapi_test.go
│   └── config_test.d/
│
├── fswatch/             # File system watcher for config hot-reload
│   ├── fswatch.go
│   └── fswatch_test.go
│
├── notify/              # Notification channel implementations
│   ├── email.go         # SMTP email
│   ├── slack.go         # Slack webhook
│   ├── sns.go           # AWS SNS
│   ├── fluentd.go       # Fluentd HTTP
│   └── stdout.go        # Standard output
│
├── util/                # Shared utilities
│   ├── util.go          # File operations, string helpers, validation
│   ├── util_test.go
│   ├── pid.go           # PID file management
│   └── pid_test.go
│
└── rpmbuild/            # RPM packaging
    ├── gorond.spec
    ├── gorond.sysconfig
    ├── goron.initd
    └── goron.sample.conf
```

---

## Build and Development Commands

```bash
# Run all tests
make test
# equivalent: go test ./...

# Build binary to build/gorond
make build

# Fetch all dependencies
make deps

# Full cycle: clean → test → build (default target)
make all

# Build and run the daemon
make run

# Remove build artifacts
make clean

# Build RPM package (requires rpmbuild tool)
make rpm
```

The compiled binary is output to `build/gorond`.

---

## CLI Flags

```
gorond [-c config_file] [-d config_dir] [-p pid_file] [-t] [-v]

  -c string   Root config file (default: /etc/goron.conf)
  -d string   Job config directory (default: /etc/goron.d/)
  -p string   PID file path (default: /var/pid/gorond)
  -t          Validate config files and exit (test mode)
  -v          Print version and exit
```

---

## Configuration

### Main Config File (`/etc/goron.conf`)

INI format, parsed via `gopkg.in/gcfg.v1`:

```ini
[config]
webApi   = localhost:6777    # REST API listen address (omit to disable)
log      = /var/log/gorond/goron.log
cronLog  = /var/log/gorond/cron.log
apiLog   = /var/log/gorond/api.log
notifyType = stdout          # stdout | mail | slack | fluentd | sns
notifyWhen = onerror         # onerror | always
subject  = [gorond] @result  # @result is replaced with command result

[mail]
dest         = alert@example.com
from         = from@example.com
smtpHost     = localhost:25
smtpUser     = username
smtpPassword = password

[slack]
Channel    = my-channel       # Without '#'
WebhookUrl = https://hooks.slack.com/...
IconUrl    = https://...      # Optional bot icon

[fluentd]
url = http://endpoint:8888/tag

[sns]
region   = ap-northeast-1
topicArn = arn:aws:sns:ap-northeast-1:000000000000:topic-name
```

### Job Config Files (`/etc/goron.d/*.conf`)

Extended cron format:

```
# Standard cron: second minute hour day month weekday user command
0 0 4 * * THU  root  /usr/bin/my-script.sh
0 4 * * * *    user  /usr/bin/other-script.sh

# Predefined schedules
@daily  root  /usr/bin/daily-task.sh

# % can be used without escaping (unlike standard cron)
0 0 * * * *  root  date +%Y-%m-%d
```

**Sequential and parallel execution**: Indented jobs with a leading `-` run after their parent succeeds. Jobs at the same indent level run in parallel:

```
0 0 4 * * *  root  first_command
             - root  second_command1    # runs in parallel after first_command
             - root  second_command2    # runs in parallel after first_command
               - root  third_command   # runs after both second commands succeed
```

**Per-file config override**: A `.conf` file in `goron.d/` may start with a `[config]` section to override global settings for its jobs only:

```ini
[config]
notifyType = mail

[mail]
dest = team@example.com

[job]
0 0 4 * * *  root  /usr/bin/app-task.sh
```

---

## Architecture and Key Concepts

### Startup Flow (`main.go`)

1. Parse CLI flags
2. `config.LoadConfig()` — loads root config + all `*.conf` from include dir
3. `goron.NewGorond(config)` — initializes cron scheduler and loggers
4. `grn.Start()` — starts the cron scheduler goroutine
5. Optionally start `webapi.NewWebApiServer()` if `webApi` is configured
6. Write PID file, then block on OS signals

### Signal Handling

| Signal | Behavior |
|--------|----------|
| `SIGTERM` | Graceful shutdown (exit code 15) |
| `SIGINT` | Ctrl-C shutdown (exit code 2) |
| `SIGHUP` | Logged but no action (reload triggered separately via fswatch) |

### Config Hot-Reload

`goron.StartAutoReload()` uses `fswatch` to watch the config file and `goron.d/` directory. On any `.conf` file change:
1. Stop the current cron scheduler
2. Re-parse all config files
3. Recreate and start the scheduler with new jobs

### Job Execution

Each job runs via `su - <user> -c <command>` for user context switching. `SystemCommand` is a package-level variable (a function), enabling easy mocking in tests.

Job status lifecycle: `WAITING` → `RUNNING` → `WAITING` (success) or `FAILED` (non-zero exit)

### Notification Flow

`goron/notify.go` dispatches to `notify/<channel>.go` based on `notifyType`. Triggered when:
- A command exits with non-zero status (always)
- `notifyWhen = always` (on every completion)

---

## Web API

When `webApi` is configured, an HTTP server exposes two endpoints:

### `GET /statuses`
Returns current execution status of all jobs:
```json
{
  "app1.conf": {
    "root first_command": "waiting",
    "root second_command1": "running"
  }
}
```
Possible statuses: `waiting`, `running`, `failed`

### `GET /jobs`
Returns raw job definitions from all `.conf` files:
```json
{
  "app1.conf": [
    "0 0 4 * * * root first_command",
    "            - root second_command1"
  ]
}
```

All other paths return HTTP 404.

---

## Testing

Every module has a corresponding `*_test.go` file. The test suite uses only the Go standard library (`testing` package) — no third-party test frameworks.

```bash
# Run all tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./config/...
go test ./goron/...
go test ./webapi/...
```

**Test conventions**:
- `SystemCommand` in `goron/goron.go` is a replaceable function variable — tests assign a mock to simulate command outcomes without shell execution
- Config test fixtures live in `config/config_test.d/` and `webapi/config_test.d/`
- Integration tests in `main_test.go` exercise the full daemon startup/shutdown lifecycle

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/robfig/cron` | Cron expression parsing and scheduling (6-field: sec min hr dom mon dow) |
| `gopkg.in/gcfg.v1` | INI-style config file parsing |
| `github.com/aws/aws-sdk-go/service/sns` | AWS SNS notifications |
| Go standard library | HTTP server, exec, signal handling, file I/O |

Install dependencies:
```bash
make deps
# or: go get -d -t -v ./...
```

---

## CI/CD

Travis CI (`.travis.yml`) runs on every branch and tag:

1. `make deps` — fetch dependencies
2. `go vet ./...` — static analysis
3. `go test -v ./...` — run tests
4. `make build` — compile binary
5. `make rpm` — build RPM package (before deploy)
6. Deploy RPM to GitHub Releases on tagged commits

---

## Code Conventions

- **Comments**: Source code comments are written in Japanese; English is acceptable for new code
- **Logging**: Use the `logging.Logger` wrapper (not `log` package directly) for structured level-based logging
- **Error handling**: Errors from job execution propagate to the notification system; fatal config errors exit the process
- **User switching**: Jobs always run through `su -` — do not bypass this for security reasons
- **Package-level vars for testing**: Functions like `SystemCommand` are declared as `var` at package scope to allow test mocking without interfaces
- **No global state in tests**: Tests construct their own configs via fixture files in `*_test.d/` directories

---

## Common Tasks

**Validate config without starting the daemon:**
```bash
./build/gorond -t -c /etc/goron.conf -d /etc/goron.d/
```

**Run with custom config paths:**
```bash
./build/gorond -c ./my.conf -d ./my-jobs/ -p /tmp/gorond.pid
```

**Check job statuses (when Web API is enabled):**
```bash
curl http://localhost:6777/statuses
curl http://localhost:6777/jobs
```

**Build and install as an RPM:**
```bash
make rpm
sudo rpm -ivh /home/travis/rpmbuild/RPMS/noarch/gorond-1.0.1-1.noarch.rpm
```
