# bov-2

`bov-2` is a starter Viam module for **arbitrary, push-style metrics capture**.

It is designed for cases where your process produces event metrics (for example, command counts, loop counters, or latency samples) and you want to upload them to Viam on demand instead of relying on fixed polling intervals.

## What this module does

- Accepts metric events from service logic through `Emit(...)`.
- Buffers metrics in a bounded in-memory queue.
- Supports manual upload through `DoCommand` (`metrics_flush_now`).
- Exposes queue and upload health via `DoCommand` (`metrics_stats`).
- Converts metric events into tabular payloads for Viam Data Client upload.

## Current status

This repository is a scaffold with working core metrics service logic and tests.

- Queueing, flushing, stats, and command handling are implemented.
- Real startup wiring to Viam Data Client is implemented through `NewModuleServiceFromEnv(...)`.

## Module layout

- `cmd/metricscapturemodule/main.go` - module entrypoint (`main`).
- `metricscapture/service.go` - top-level module service wrapper.
- `metricscapture/startup.go` - startup wiring for API key auth + uploader construction.
- `metricscapture/commands.go` - `DoCommand` routing.
- `metricscapture/metrics/types.go` - core metric/result/stat types.
- `metricscapture/metrics/queue.go` - bounded queue with drop accounting.
- `metricscapture/metrics/service.go` - emit/flush/stats service behavior.
- `metricscapture/metrics/docommand.go` - metrics commands (`metrics_flush_now`, `metrics_stats`).
- `metricscapture/metrics/uploader.go` - adapter interfaces and tabular request mapping.
- `internal/dataclient/` - concrete Viam data client adapter + mapping + retry.
- `test/metrics_queue_test.go` - flush success/failure behavior tests.

## How to use

### 0) Wire startup and credentials

Set required environment variables:

```bash
export VIAM_API_KEY="your-api-key"
export VIAM_API_KEY_ID="your-api-key-id"
```

Then initialize service startup with `NewModuleServiceFromEnv(...)`:

```go
logger := logging.NewDebugLogger("metricscapturemodule")
svc, err := metricscapture.NewModuleServiceFromEnv(context.Background(), logger, metricscapture.StartupConfig{
    PartID:        "YOUR_PART_ID",
    ComponentType: "rdk:component:sensor",
    ComponentName: "metricscapture",
    MethodName:    "Readings",
    Tags:          []string{"metricscapture"},
    MaxQueueSize:  1000,
    BatchSize:     100,
})
if err != nil {
    return err
}
defer svc.Close(context.Background())
```

### 1) Emit metrics from your service logic

Call `Emit(...)` (or `EmitMetric(...)` from `ModuleService`) whenever an event happens:

- command executed
- loop iteration completed
- operation failed
- latency measured

Example metric names:

- `metricscapture.do_command_total`
- `metricscapture.auto_mode_loop_total`
- `metricscapture.move_cancelled_total`

### 2) Flush metrics on demand with DoCommand

Send this command to upload queued metrics now:

```json
{"command":"metrics_flush_now"}
```

Expected response keys:

- `status`
- `uploaded_count`
- `file_ids` (if returned by uploader)

### 3) Inspect queue and upload health

Send:

```json
{"command":"metrics_stats"}
```

Response includes:

- `queue_depth`
- `dropped_count`
- `uploaded_count`
- `upload_fail_count`
- `last_upload_at`
- `last_upload_error`
- `max_queue_size`
- `batch_size`

## Example in Viam

Use this as a simple end-to-end flow once your module binary and model registration are wired.

### 1) Add the service to your machine config

In your Viam machine config, add this module and service (replace placeholders with your real values):

```json
{
  "modules": [
    {
      "name": "metricscapture-module",
      "executable_path": "/path/to/metricscapturemodule"
    }
  ],
  "services": [
    {
      "name": "metricscapture",
      "api": "rdk:service:generic",
      "model": "bov:metricscapture:service",
      "attributes": {
        "part_id": "YOUR_PART_ID",
        "component_type": "rdk:component:sensor",
        "component_name": "metricscapture",
        "method_name": "Readings",
        "tags": ["metricscapture"],
        "max_queue_size": 1000,
        "batch_size": 100,
        "retry_max_attempts": 3,
        "retry_base_delay_ms": 300
      }
    }
  ]
}
```

Set these environment variables on the machine that runs the module process:

```bash
export VIAM_API_KEY="your-api-key"
export VIAM_API_KEY_ID="your-api-key-id"
```

### 2) Emit metrics from your service logic

Where your module handles real events, emit metrics like:

```go
svc.EmitMetric("metricscapture.do_command_total", 1, map[string]string{"command": "start"})
svc.EmitMetric("metricscapture.auto_mode_loop_total", 1, nil)
svc.EmitMetric("metricscapture.move_cancelled_total", 1, map[string]string{"reason": "bumper"})
```

### 3) Trigger on-demand upload

From your Viam client code (or any path that can call `DoCommand`), send:

```json
{"command":"metrics_flush_now"}
```

Expected success response:

```json
{
  "status": "ok",
  "uploaded_count": 3,
  "file_ids": ["..."]
}
```

### 4) Check health/status

Send:

```json
{"command":"metrics_stats"}
```

You should see queue depth and upload counters update after flush.

## Running tests

From repo root:

```bash
go test ./...
```

Useful options:

- `go test -v ./...` for verbose output
- `go test ./test -run TestFlushNowSuccessEmptiesQueue -v` for one test

## Reference

- [TabularDataCaptureUpload docs](https://docs.viam.com/reference/apis/data-client/#tabulardatacaptureupload)
