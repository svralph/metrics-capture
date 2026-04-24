# Metrics package

This package is responsible for:

- collecting push-style metric events from the service,
- buffering events in a bounded queue,
- uploading batches to Viam data APIs, and
- exposing `doCommand` handlers for manual flush and stats.
