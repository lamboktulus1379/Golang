package providers

// Config captures the minimal configuration required to bootstrap an OAuth provider implementation.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}
