# OAuth Skeleton

This repository contains a minimal skeleton for an OAuth aggregation service written in Go. The goal is to demonstrate how to wire a main entry point, an authentication service layer, and provider-specific implementations (Google, Facebook) while leaving the HTTP handlers and real token exchange logic for later.

## Project Layout

```
.
├── main.go                     # Bootstraps providers and starts the HTTP server
├── server/
│   └── server.go               # net/http handlers exposing the OAuth flow
├── providers/
│   ├── config.go               # Shared provider configuration struct
│   ├── facebook/
│   │   └── facebook.go         # Facebook provider skeleton
│   └── google/
│       └── google.go           # Google provider skeleton
└── service/
    ├── service.go              # Service interface and AuthService implementation
    ├── tenant_service.go       # Optional tenant-aware wrapper delegating to AuthService
    └── types.go                # Token and UserProfile domain types
```

## Getting Started

1. Update `go.mod` with your module path (e.g. `github.com/yourname/oauth-skeleton`).
2. Replace the placeholder client IDs, secrets, redirect URIs, and scopes in `main.go` with real configuration.
3. Implement the `Exchange` and `FetchUserProfile` methods inside each provider package using the identity provider’s API endpoints.
4. Start the development server:

```bash
go run .
```

5. Visit or `curl` the endpoints shown below to exercise the flow.

### Service interface

The `service` package exposes a `Service` interface so callers (including the HTTP server) can depend on abstraction-friendly methods:

```go
type Service interface {
    Register(provider Provider) error
    RegisterProviderFactory(name string, builder ProviderBuilder) error
    ProviderSession(ctx context.Context, providerName string) (ProviderSession, error)
    LoginURL(ctx context.Context, providerName string, params LoginParams) (string, error)
    ExchangeCode(ctx context.Context, providerName string, params ExchangeParams) (*Token, error)
    UserProfile(ctx context.Context, providerName string, token *Token) (*UserProfile, error)
}

type LoginParams struct {
    State               string
    CodeChallenge       string
    CodeChallengeMethod string
}

type ExchangeParams struct {
    Code         string
    CodeVerifier string
}
```


### Provider sessions

When you need to call multiple provider-specific operations within the same request, obtain a `ProviderSession` once and reuse it. Pass whatever tenant identifier your upstream service uses (an empty string works for single-tenant deployments if your factory accepts it):

```go
session, err := svc.ProviderSession(ctx, "google")
if err != nil {
    // handle err
}

state := generateRandomState()
codeVerifier := generatePKCEVerifier()

url, _ := session.LoginURL(service.LoginParams{
    State:               state,
    CodeChallenge:       pkceChallenge(codeVerifier),
    CodeChallengeMethod: "S256",
})
token, _ := session.ExchangeCode(ctx, service.ExchangeParams{
    Code:         code,
    CodeVerifier: codeVerifier,
})
profile, _ := session.UserProfile(ctx, token)
```

Internally this avoids repeating the provider lookup while giving you a clear, provider-scoped API. `AuthService` is the default implementation, but you can supply your own (for testing or alternative backends) anywhere the interface is accepted, such as the `server` package.

### Tenant routing (optional)

If you need tenant-specific provider configuration, compose an `AuthService` with the `TenantService` wrapper. It keeps the core service tenant-agnostic while still letting callers insist on a tenant value at the edge (e.g. HTTP handlers):

```go
tenantSvc := service.NewTenantService()
tenantSvc.RegisterTenantProviderFactory("acme", "google", func(ctx context.Context, tenant string) (service.Provider, error) {
    cfg := lookupConfig(tenant)
    return google.New(cfg), nil
})

session, err := tenantSvc.ProviderSession(ctx, "acme", "google")
```

The server in this skeleton depends on the tenant-aware interface so that routes still require a `tenant` query parameter, but you can wire whichever layer matches your deployment.

#### Single-tenant usage

If you don't need tenant isolation, work with `AuthService` directly. Register your providers once and call the service methods without a tenant parameter:

```go
svc := service.NewAuthService()

svc.RegisterProviderFactory("google", func(ctx context.Context) (service.Provider, error) {
    return google.New(google.Config{ /* single-tenant config */ }), nil
})

loginURL, _ := svc.LoginURL(ctx, "google", service.LoginParams{
    State:               "state-value",
    CodeChallenge:       pkceChallenge("verifier"),
    CodeChallengeMethod: "S256",
})
token, _ := svc.ExchangeCode(ctx, "google", service.ExchangeParams{
    Code:         "code-from-callback",
    CodeVerifier: "verifier",
})
profile, _ := svc.UserProfile(ctx, "google", token)
```

When using the provided server, omit the `tenant` query parameter (or pass it empty) and it will fall back to the shared configuration you registered.

### State and PKCE handling

- The `/auth/{provider}/login` handler now generates a high-entropy state value and PKCE code verifier whenever you don’t supply them. Both values are stored in an HTTP-only cookie scoped to your domain.
- The returned JSON includes the state so browser-based clients can keep it alongside their own routing data if desired. The PKCE code challenge is already embedded in the redirect URL.
- On `/auth/{provider}/callback`, the server verifies the `state` query parameter matches the saved cookie and automatically forwards the stored PKCE verifier when calling `ExchangeCode`.
- If you operate a custom frontend, make sure it preserves cookies between the login call and the provider redirect so the callback can validate and finish the flow.
- Prefer HTTPS in production and mark the cookie as secure; the skeleton toggles the `Secure` flag automatically when TLS is present.

### Swap persistence and configuration sources

- Session storage is abstracted behind `server.LoginSessionStore`. The default `server.New` call accepts either your implementation or falls back to the included in-memory `server.NewMemorySessionStore`. Inject a Redis-backed store during wiring to survive multi-instance deployments.
- Provider configuration is loaded through the `providers.ConfigLoader` interface. The repository ships with `providers.NewStaticConfigLoader` for map-based configs, but you can implement the interface to pull settings from Vault, SQL, or any other source before registering providers with `TenantService`.
- Both abstractions make it easy to stub dependencies in tests—swap in ephemeral stores or deterministic config loaders to simulate edge cases without touching production wiring.

### Diagram view

```mermaid
flowchart LR
    subgraph Client
        A[HTTP Request]
    end
    subgraph Server Layer
        B[server.Server]
        C[service.TenantService]
        D[service.AuthService]
    end
    subgraph Providers
        E[ProviderBuilder]
        F[Provider]
    end

    A -->|/auth/{provider}/...| B
    B -->|tenant, provider| C
    C -->|lookup tenant| D
    D -->|resolve builder| E
    E -->|create instance| F
    F -->|AuthCodeURL/Exchange/Profile| B
```

#### Previewing the diagram in VS Code

If the diagram shows only raw code, try these steps:

1. Open this file in VS Code and launch the Markdown preview with <kbd>⇧</kbd><kbd>⌘</kbd><kbd>V</kbd> (or use the command palette: **Markdown: Open Preview to the Side**).
2. When prompted, mark the workspace as **Trusted** so the preview can run the Mermaid renderer.
3. Ensure Mermaid support is enabled: search for `Markdown › Preview: Mermaid` in Settings and turn it on. On older VS Code versions, install the [Markdown Preview Mermaid Support](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid) extension and reload the window.
4. If the preview bar reports an error, reopen the preview or reload the window (<kbd>⌘</kbd><kbd>R</kbd>) to give the renderer a clean reset.

Prefer a CLI render instead? Install the Mermaid CLI and export an SVG:

```bash
npm install --global @mermaid-js/mermaid-cli
mmdc --input README.md --output diagram.svg
```

The generated `diagram.svg` can be embedded in other docs or wikis, and you can pass `--configFile` later if you need custom theming.

## HTTP Endpoints

| Method | Path                                   | Description |
| ------ | -------------------------------------- | ----------- |
| GET    | `/healthz`                             | Liveness check. |
| GET    | `/auth/{provider}/login?tenant={id}&state=STATE`   | Returns the provider-specific authorization URL. `state` is optional—if omitted the server emits one, sets it in a cookie, and includes it in the JSON response. PKCE (S256) is enabled automatically. |
| GET    | `/auth/{provider}/callback?tenant={id}&code=CODE&state=STATE`  | Validates the `state` against the login cookie, uses the stored PKCE verifier during exchange, and returns the user profile (stubbed). |

Example login call:

```bash
curl "http://localhost:8080/auth/google/login?tenant=acme&state=test-state"
```

Since the provider skeletons do not yet perform real OAuth exchanges, the callback endpoint will respond with placeholder errors until you flesh out `Exchange` and `FetchUserProfile`.

## Changing the Module Path

To change the module path throughout the project:

```bash
cd /path/to/repo
# Update the module path in go.mod
go mod edit -module=github.com/yourname/oauth-skeleton
# (Optional) tidy dependencies
go mod tidy
```

After editing `go.mod`, update import paths that reference the old module path (e.g. in `main.go`, provider packages, or tests). The skeleton already uses the current module path `github.com/lamboktulus1379/oauth-skeleton`.

## Next Steps

- Add persistent storage for sessions or user records once tokens are issued.
- Support additional providers by creating new packages under `providers/` and registering them in `main.go`.
- Wire HTTP routing and middleware to integrate the service into your application stack.
