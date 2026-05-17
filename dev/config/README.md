# Local Development Config

Files in this directory are for optional local developer settings.

## ngrok Tunnel

Create `dev/config/tunnel-url` to have Tilt start an ngrok tunnel and render the
tunnel values into Kubernetes config.

Format:

```text
<reserved-ngrok-domain>[:<local-target-port>]
```

Example:

```text
flowing-properly-moth.ngrok-free.app:443
```

If the port is omitted, Tilt uses `443`. A pasted `https://` URL is also
accepted; Tilt strips the scheme and any path before starting ngrok.

When the file exists and has a value, Tilt:

- creates a `dev-tunnel-config` Kubernetes ConfigMap
- sets API `MCP_BASE_URL` and `MCP_AUDIENCE` from that ConfigMap
- starts `ngrok http --url=<reserved-ngrok-domain> <local-target-port>`

When the file is absent or empty, Tilt skips the ConfigMap and ngrok resource,
and local MCP uses `https://local.4ks.io/mcp`.

`dev/config/tunnel-url` is ignored by git because the value is developer-local.
