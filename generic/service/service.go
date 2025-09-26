package service

import (
	"context"
	"errors"
	"fmt"
)

// Provider is the contract each third-party OAuth provider implementation must satisfy.
type Provider interface {
	Name() string
	AuthCodeURL(params LoginParams) (string, error)
	Exchange(ctx context.Context, params ExchangeParams) (*Token, error)
	FetchUserProfile(ctx context.Context, token *Token) (*UserProfile, error)
}

// ProviderSession exposes a provider-specific view used once per request flow.
type ProviderSession interface {
	LoginURL(params LoginParams) (string, error)
	ExchangeCode(ctx context.Context, params ExchangeParams) (*Token, error)
	UserProfile(ctx context.Context, token *Token) (*UserProfile, error)
}

// ProviderBuilder manufactures provider instances on demand.
type ProviderBuilder func(ctx context.Context) (Provider, error)

// Service is the public contract consumers should rely on.
type Service interface {
	Register(provider Provider) error
	RegisterProviderFactory(name string, builder ProviderBuilder) error
	ProviderSession(ctx context.Context, providerName string) (ProviderSession, error)
	LoginURL(ctx context.Context, providerName string, params LoginParams) (string, error)
	ExchangeCode(ctx context.Context, providerName string, params ExchangeParams) (*Token, error)
	UserProfile(ctx context.Context, providerName string, token *Token) (*UserProfile, error)
}

// AuthService coordinates registered OAuth providers and exposes high-level flow helpers.
type AuthService struct {
	factories map[string]ProviderBuilder
}

// NewAuthService bootstraps an AuthService with the given providers.
func NewAuthService() *AuthService {
	return &AuthService{factories: make(map[string]ProviderBuilder)}
}

// Register adds a provider to the service. It fails when a provider name already exists.
func (s *AuthService) Register(provider Provider) error {
	if provider == nil {
		return errors.New("provider: nil registration")
	}
	name := provider.Name()
	if name == "" {
		return errors.New("provider: empty name")
	}
	return s.RegisterProviderFactory(name, func(context.Context) (Provider, error) {
		return provider, nil
	})
}

// RegisterProviderFactory registers a provider builder.
func (s *AuthService) RegisterProviderFactory(name string, builder ProviderBuilder) error {
	if builder == nil {
		return errors.New("provider: nil factory")
	}
	if name == "" {
		return errors.New("provider: empty name")
	}
	if _, exists := s.factories[name]; exists {
		return fmt.Errorf("provider: %s already registered", name)
	}
	s.factories[name] = builder
	return nil
}

// ProviderSession returns a provider-specific session to avoid repeated lookups within a request lifecycle.
func (s *AuthService) ProviderSession(ctx context.Context, providerName string) (ProviderSession, error) {
	factory, ok := s.factories[providerName]
	if !ok {
		return nil, fmt.Errorf("provider: %s not found", providerName)
	}
	provider, err := factory(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider: factory for %s returned nil", providerName)
	}
	return providerSession{provider: provider}, nil
}

// LoginURL delegates to the named provider to build the authorization URL.
func (s *AuthService) LoginURL(ctx context.Context, providerName string, params LoginParams) (string, error) {
	session, err := s.ProviderSession(ctx, providerName)
	if err != nil {
		return "", err
	}
	return session.LoginURL(params)
}

// ExchangeCode handles the exchange step for the named provider.
func (s *AuthService) ExchangeCode(ctx context.Context, providerName string, params ExchangeParams) (*Token, error) {
	session, err := s.ProviderSession(ctx, providerName)
	if err != nil {
		return nil, err
	}
	return session.ExchangeCode(ctx, params)
}

// UserProfile fetches user information from the provider using the given token.
func (s *AuthService) UserProfile(ctx context.Context, providerName string, token *Token) (*UserProfile, error) {
	session, err := s.ProviderSession(ctx, providerName)
	if err != nil {
		return nil, err
	}
	return session.UserProfile(ctx, token)
}

type providerSession struct {
	provider Provider
}

func (p providerSession) LoginURL(params LoginParams) (string, error) {
	return p.provider.AuthCodeURL(params)
}

func (p providerSession) ExchangeCode(ctx context.Context, params ExchangeParams) (*Token, error) {
	return p.provider.Exchange(ctx, params)
}

func (p providerSession) UserProfile(ctx context.Context, token *Token) (*UserProfile, error) {
	return p.provider.FetchUserProfile(ctx, token)
}
