# traefik-middlewares

This repository contains custom Traefik middleware plugins. Each subfolder is a standalone plugin that can be loaded into Traefik via its [plugin system](https://doc.traefik.io/traefik/plugins/).

## Plugins

| Plugin | Description |
|--------|-------------|
| [`add-path-header`](./add-path-header/) | Extracts the path group (normalized path with IDs replaced) into a configurable request header |

---

## add-path-header

Extracts the path group (normalized path with IDs replaced by `*`) into a request header before forwarding it to the upstream service. ID segments (UUIDs, numeric IDs, alphanumeric slugs) are replaced with `*` to create a normalized path group. Useful for grouping requests by path pattern rather than specific IDs.

**Configuration:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `headerName` | `string` | `x-path-group` | Name of the request header to set |

---

## Usage in terraform-aws-eks-cluster

The `terraform-aws-eks-cluster` module manages Traefik via a `helm_release`. Plugins are registered in the `experimental.plugins` block of the Helm values, and then activated per-route using a `Middleware` CRD.

### 1. Enable the plugin in the module

In the `ingress` config of your cluster module call, add the plugin to the `experimental.plugins` block. In `k8s_ingress.tf`, extend the `experimental` value passed to the Helm release:

```hcl
experimental = {
  plugins = {
    addPathHeader = {
      moduleName = "github.com/syltek/traefik-middlewares/add-path-header"
      version    = "v0.1.0"
    }
  }
}
```

The full context inside `helm_release.traefik_ingress` values looks like:

```hcl
values = [yamlencode(merge(
  {
    # ... existing values ...
    experimental = {
      plugins = {
        addPathHeader = {
          moduleName = "github.com/syltek/traefik-middlewares/add-path-header"
          version    = "v0.1.0"
        }
      }
    }
  }
))]
```

### 2. Create the Middleware resource

Add a `Middleware` CRD to the `extraObjects` list in the same Helm release, or deploy it separately as a Kubernetes manifest:

```hcl
extraObjects = [
  {
    apiVersion = "traefik.io/v1alpha1"
    kind       = "Middleware"
    metadata = {
      name      = "add-path-header"
      namespace = "traefik"
    }
    spec = {
      plugin = {
        addPathHeader = {
          headerName = "x-path-group"
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
  namespace: anemone
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`my-service.example.com`)
      kind: Rule
      middlewares:
        - name: add-path-header
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

The GitHub Actions workflow (`.github/workflows/test.yaml`) runs on every pull request and push to `main`. For each plugin it runs:

- `gofmt` — fails if code is not properly formatted
- `go vet` — static analysis
- `go mod tidy` check — fails if `go.mod`/`go.sum` are not up to date
- `yaegi test` — runs tests using the Yaegi interpreter (the same runtime Traefik uses to execute plugins)
