<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".assets/playtomic-logo-dark.png">
    <source media="(prefers-color-scheme: light)" srcset=".assets/playtomic-logo-light.png">
    <img alt="Playtomic Logo" width="400">
  </picture>
</div>


<h1 align="center">Traefik Add Path Group Middleware</h1>

This repository contains a custom Traefik middleware plugin that extracts the path group (normalized path with IDs replaced by labels) into a request header before forwarding it to the upstream service. ID segments (UUIDs, numeric IDs, alphanumeric slugs) are replaced with labels to create a normalized path group. Useful for grouping requests by path pattern rather than specific IDs.

<div align="center">
  <picture>
    <source srcset=".assets/plugin-logo.png">
    <img alt="Plugin Logo" width="250">
  </picture>
</div>


## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `headerName` | `string` | `x-path-group` | Name of the request header to set |

## Example

The following paths will be normalized to the following path group and added to the `x-path-group` header:

```
/api/v1/users/550e8400-e29b-41d4-a716-446655440000/profile -> /api/v1/users/uuid/profile
/api/v1/courts/42/bookings -> /api/v1/courts/numeric_id/bookings
/api/v1/bookings/booking-abc-99/details -> /api/v1/bookings/slug/details
/api/v1/users/user_42/profile -> /api/v1/users/slug/profile
/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000/courts/42/bookings/booking-abc-99 -> /api/v1/tenants/uuid/courts/numeric_id/bookings/slug
```



## Usage in Traefik

If you manage Traefik via a `helm_release`, plugins are registered in the `experimental.plugins` block of the Helm values, and then activated per-route using a `Middleware` CRD.

### 1. Enable the plugin in the module

Add the plugin to the `experimental.plugins` block with the desired plugin and
version.

```hcl
experimental = {
  plugins = {
    addPathHeader = {
      moduleName = "github.com/syltek/traefik-add-path-group-middleware"
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
        addPathGroup = {
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
