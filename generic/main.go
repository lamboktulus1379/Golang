package main

import (
	"context"
	"log"
	"net/http"

	"github.com/lamboktulus1379/oauth-skeleton/providers"
	"github.com/lamboktulus1379/oauth-skeleton/providers/facebook"
	"github.com/lamboktulus1379/oauth-skeleton/providers/google"
	"github.com/lamboktulus1379/oauth-skeleton/server"
	"github.com/lamboktulus1379/oauth-skeleton/service"
)

func main() {
	googleConfigs := map[string]providers.Config{
		"": {
			ClientID:     "GOOGLE_CLIENT_ID",
			ClientSecret: "GOOGLE_CLIENT_SECRET",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"email", "profile"},
		},
		"acme": {
			ClientID:     "ACME_GOOGLE_CLIENT_ID",
			ClientSecret: "ACME_GOOGLE_CLIENT_SECRET",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"email", "profile"},
		},
	}

	facebookConfigs := map[string]providers.Config{
		"": {
			ClientID:     "FACEBOOK_APP_ID",
			ClientSecret: "FACEBOOK_APP_SECRET",
			RedirectURL:  "http://localhost:8080/auth/facebook/callback",
			Scopes:       []string{"public_profile", "email"},
		},
		"acme": {
			ClientID:     "ACME_FACEBOOK_APP_ID",
			ClientSecret: "ACME_FACEBOOK_APP_SECRET",
			RedirectURL:  "http://localhost:8080/auth/facebook/callback",
			Scopes:       []string{"public_profile", "email"},
		},
	}

	configLoader := providers.NewStaticConfigLoader(map[string]map[string]providers.Config{
		"google":   googleConfigs,
		"facebook": facebookConfigs,
	})

	tenantService := service.NewTenantService()

	providerFactories := map[string]func(providers.Config) service.Provider{
		"google":   func(cfg providers.Config) service.Provider { return google.New(cfg) },
		"facebook": func(cfg providers.Config) service.Provider { return facebook.New(cfg) },
	}

	records, err := configLoader.Load(context.Background())
	if err != nil {
		log.Fatalf("failed to load provider configs: %v", err)
	}

	for _, record := range records {
		factory, ok := providerFactories[record.Provider]
		if !ok {
			log.Printf("skipping unrecognized provider %q", record.Provider)
			continue
		}

		cfg := record.Config
		tenant := record.Tenant
		providerName := record.Provider

		if err := tenantService.RegisterTenantProviderFactory(tenant, providerName, func(_ context.Context, _ string) (service.Provider, error) {
			return factory(cfg), nil
		}); err != nil {
			log.Fatalf("failed to register %s provider for tenant %q: %v", providerName, tenant, err)
		}
	}

	httpServer := server.New(tenantService, nil)
	addr := ":8080"

	log.Printf("OAuth skeleton listening on %s", addr)
	log.Printf("Use GET /auth/google/login?tenant=acme&state=STATE (omit tenant to use shared config)")
	log.Printf("Handle the callback at GET /auth/google/callback?tenant=acme&code=AUTH_CODE")

	if err := http.ListenAndServe(addr, httpServer); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
