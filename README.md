# tacacs-exporter

A Prometheus exporter similar to blackbox_exporter but for TACACS servers.


## Usage
```
Usage of tacacs-exporter:
  -config string
        Path to the config file. (default "config.yml")
  -log.level string
        The log level. (default "info")
  -version
        Show the version
  -web.listen-address string
        HTTP server listen address (default ":9949")
  -web.telemetry-path string
        Path under which to expose metrics (default "/metrics")
```

## Configuration

```yaml
---
modules:
  module_name:
    # Required
    username: tacacs-test
    # Required
    password: tacacs-test
    # TACACS shared secret. Required.
    secret: s3cret
    # Use TACACS single connect mode. Default: false
    single_connect: false
    # Legacy single connect mode. For devices that don't set the single-connection header flag.
    # Default: false
    legacy_single_connect: false
    # TACACS timeout in seconds.
    # Default: 5
    timeout: 5
    # Privilege level to send to the TACACS server
    # Default: 0
    privilege_level: 0
    # Device port to send to the TACACS server
    port: probe
```
