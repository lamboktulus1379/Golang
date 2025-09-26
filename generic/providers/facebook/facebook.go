package facebook

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/lamboktulus1379/oauth-skeleton/providers"
	"github.com/lamboktulus1379/oauth-skeleton/service"
)

const providerName = "facebook"

// Provider implements the service.Provider contract for Facebook OAuth.
type Provider struct {
	config providers.Config
}

// New constructs a Facebook OAuth provider skeleton.
func New(config providers.Config) *Provider {
	return &Provider{config: config}
}

// Name returns the canonical provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// AuthCodeURL builds the Facebook authorization URL.
func (p *Provider) AuthCodeURL(params service.LoginParams) (string, error) {
	if err := p.validate(); err != nil {
		return "", err
	}

	query := url.Values{
		"client_id":     {p.config.ClientID},
		"redirect_uri":  {p.config.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(p.config.Scopes, ",")},
		"state":         {params.State},
	}

	if params.CodeChallenge != "" {
		query.Set("code_challenge", params.CodeChallenge)
		if params.CodeChallengeMethod != "" {
			query.Set("code_challenge_method", params.CodeChallengeMethod)
		}
	}

	return fmt.Sprintf("https://www.facebook.com/v18.0/dialog/oauth?%s", query.Encode()), nil
}

// Exchange exchanges the authorization code for a token. The implementation is left as a TODO for consumers.
func (p *Provider) Exchange(ctx context.Context, params service.ExchangeParams) (*service.Token, error) {
	return nil, fmt.Errorf("facebook: exchange not implemented: received code %q", params.Code)
}

// FetchUserProfile retrieves the Facebook user profile given a token. The implementation is left as a TODO for consumers.
func (p *Provider) FetchUserProfile(ctx context.Context, token *service.Token) (*service.UserProfile, error) {
	return nil, fmt.Errorf("facebook: fetch user profile not implemented")
}

func (p *Provider) validate() error {
	if p.config.ClientID == "" {
		return fmt.Errorf("facebook: missing client ID")
	}
	if p.config.ClientSecret == "" {
		return fmt.Errorf("facebook: missing client secret")
	}
	if p.config.RedirectURL == "" {
		return fmt.Errorf("facebook: missing redirect URL")
	}
	if len(p.config.Scopes) == 0 {
		return fmt.Errorf("facebook: at least one scope is required")
	}
	return nil
}
