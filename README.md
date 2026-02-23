<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/playtomic-logo-dark.png">
    <source media="(prefers-color-scheme: light)" srcset="docs/playtomic-logo-light.png">
    <img alt="Playtomic Logo" width="400">
  </picture>
</div>


<h1 align="center">Traefik Middlewares</h1>

This repository contains custom Traefik middleware plugins. Each subfolder is a standalone plugin that can be loaded into Traefik via its [plugin system](https://doc.traefik.io/traefik/plugins/).

## Plugins List

| Plugin | Description |
|--------|-------------|
| [`add-path-header`](./add-path-header/) | Extracts the path group (normalized path with IDs replaced) into a configurable request header |

---

### add-path-header

Extracts the path group (normalized path with IDs replaced by `*`) into a request header before forwarding it to the upstream service. ID segments (UUIDs, numeric IDs, alphanumeric slugs) are replaced with `*` to create a normalized path group. Useful for grouping requests by path pattern rather than specific IDs.

**Configuration:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `headerName` | `string` | `x-path-group` | Name of the request header to set |

---

## Usage in Traefik

If you manage Traefik via a `helm_release`, plugins are registered in the `experimental.plugins` block of the Helm values, and then activated per-route using a `Middleware` CRD.

### 1. Enable the plugin in the module

Add the plugin to the `experimental.plugins` block with the desired plugin and
version.

```hcl
experimental = {
  plugins = {
    addPathHeader = {
      moduleName = "github.com/syltek/traefik-middlewares/<plugin-name>"
      version    = "<plugin-version>"
    }
  }
}
```

### 2. Create the Middleware resource

Add a `Middleware` CRD to the `extraObjects` list in the same Helm release, or deploy it separately as a Kubernetes manifest:

```hcl
extraObjects = [
  {
    apiVersion = "traefik.io/v1alpha1"
    kind       = "Middleware"
    metadata = {
      name      = "<middleware-name>"
      namespace = "traefik"
    }
    spec = {
      plugin = {
        <plugin-name> = {
          <plugin-config> = <plugin-config-value>
        }
      }
    }
  }
]
```

### 3. Attach the middleware to a route

Reference the middleware in an `IngressRoute` or `Ingress` annotation:

```yaml
# IngressRoute example
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-service
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`my-service.example.com`)
      kind: Rule
      middlewares:
        - name: <middleware-name>
          namespace: traefik
      services:
        - name: my-service
          port: 8080
```

---

## Adding a new plugin

1. Create a new subfolder with the plugin name (e.g. `my-plugin/`)
2. Add the required files:
   - `go.mod` — Go module with path `github.com/syltek/traefik-middlewares/<plugin-name>`
   - `.traefik.yml` — plugin manifest
   - `<plugin_name>.go` — plugin implementation
   - `<plugin_name>_test.go` — tests compatible with `yaegi test`
3. The CI workflow automatically discovers and tests all plugin directories containing a `go.mod`

## CI

The GitHub Actions workflow (`.github/workflows/test.yaml`) runs on every pull request for changed plugins. For each plugin it runs:

- `gofmt` — fails if code is not properly formatted
- `go vet` — static analysis
- `go mod tidy` check — fails if `go.mod`/`go.sum` are not up to date
- `yaegi test` — runs tests using the Yaegi interpreter (the same runtime Traefik uses to execute plugins)

When a plugin is pushed to `main`, the GitHub Actions workflow (`.github/workflows/push.yaml`) runs and:

- creates a new tag
- pushes the tag to the repository
