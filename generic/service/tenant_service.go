package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// TenantProviderBuilder manufactures provider instances for a specific tenant.
type TenantProviderBuilder func(ctx context.Context, tenant string) (Provider, error)

// TenantService routes calls to tenant-scoped AuthService instances.
type TenantService struct {
	mu       sync.RWMutex
	services map[string]*AuthService
}

// NewTenantService constructs a TenantService with no registered tenants.
func NewTenantService() *TenantService {
	return &TenantService{
		services: make(map[string]*AuthService),
	}
}

// RegisterTenantProvider registers a concrete provider for the given tenant.
func (t *TenantService) RegisterTenantProvider(tenant string, provider Provider) error {
	svc := t.ensureService(tenant)
	return svc.Register(provider)
}

// RegisterTenantProviderFactory registers a factory for the given tenant that can build provider instances on demand.
func (t *TenantService) RegisterTenantProviderFactory(tenant, name string, builder TenantProviderBuilder) error {
	if builder == nil {
		return errors.New("tenant: nil provider factory")
	}
	svc := t.ensureService(tenant)
	return svc.RegisterProviderFactory(name, func(ctx context.Context) (Provider, error) {
		return builder(ctx, tenant)
	})
}

// ProviderSession resolves a tenant-specific provider session.
func (t *TenantService) ProviderSession(ctx context.Context, tenant, providerName string) (ProviderSession, error) {
	svc, err := t.serviceForTenant(tenant)
	if err != nil {
		return nil, err
	}
	return svc.ProviderSession(ctx, providerName)
}

// LoginURL delegates to the tenant's provider to build an auth URL.
func (t *TenantService) LoginURL(ctx context.Context, tenant, providerName string, params LoginParams) (string, error) {
	svc, err := t.serviceForTenant(tenant)
	if err != nil {
		return "", err
	}
	return svc.LoginURL(ctx, providerName, params)
}

// ExchangeCode performs the code exchange for the tenant's provider.
func (t *TenantService) ExchangeCode(ctx context.Context, tenant, providerName string, params ExchangeParams) (*Token, error) {
	svc, err := t.serviceForTenant(tenant)
	if err != nil {
		return nil, err
	}
	return svc.ExchangeCode(ctx, providerName, params)
}

// UserProfile fetches a profile from the tenant's provider using the provided token.
func (t *TenantService) UserProfile(ctx context.Context, tenant, providerName string, token *Token) (*UserProfile, error) {
	svc, err := t.serviceForTenant(tenant)
	if err != nil {
		return nil, err
	}
	return svc.UserProfile(ctx, providerName, token)
}

func (t *TenantService) ensureService(tenant string) *AuthService {
	t.mu.Lock()
	defer t.mu.Unlock()
	svc, ok := t.services[tenant]
	if !ok {
		svc = NewAuthService()
		t.services[tenant] = svc
	}
	return svc
}

func (t *TenantService) serviceForTenant(tenant string) (*AuthService, error) {
	t.mu.RLock()
	svc, ok := t.services[tenant]
	t.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("tenant: %s not registered", tenant)
	}
	return svc, nil
}
