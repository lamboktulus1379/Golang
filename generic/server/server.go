package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lamboktulus1379/oauth-skeleton/service"
)

// TenantAuthService captures the tenant-aware operations the HTTP layer relies on.
type TenantAuthService interface {
	ProviderSession(ctx context.Context, tenant, providerName string) (service.ProviderSession, error)
}

// Server exposes HTTP handlers for the OAuth service flow.
type Server struct {
	auth     TenantAuthService
	mux      *http.ServeMux
	sessions LoginSessionStore
}

// New constructs a Server for the provided AuthService.
func New(auth TenantAuthService, store LoginSessionStore) *Server {
	if store == nil {
		store = NewMemorySessionStore(defaultSessionTTL)
	}
	s := &Server{
		auth:     auth,
		mux:      http.NewServeMux(),
		sessions: store,
	}

	s.registerRoutes()

	return s
}

// ServeHTTP satisfies http.Handler by delegating to the internal mux.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	s.mux.HandleFunc("/auth/", s.authHandler)
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segments) != 3 || segments[0] != "auth" {
		http.NotFound(w, r)
		return
	}

	provider := segments[1]
	action := segments[2]

	switch action {
	case "login":
		s.handleLogin(w, r, provider)
	case "callback":
		s.handleCallback(w, r, provider)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request, provider string) {
	state := r.URL.Query().Get("state")
	if state == "" {
		generated, err := generateState()
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}
		state = generated
	}

	tenant := r.URL.Query().Get("tenant")

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}
	codeChallenge := codeChallengeS256(codeVerifier)

	session, err := s.auth.ProviderSession(r.Context(), tenant, provider)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err)
		return
	}

	loginURL, err := session.LoginURL(service.LoginParams{
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethodS256,
	})
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err)
		return
	}

	handle, err := s.sessions.Create(LoginSession{
		Tenant:       tenant,
		Provider:     provider,
		State:        state,
		CodeVerifier: codeVerifier,
	})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthSessionCookie,
		Value:    handle.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  handle.ExpiresAt,
	})

	s.writeJSON(w, http.StatusOK, map[string]string{
		"tenant":                tenant,
		"provider":              provider,
		"redirect_url":          loginURL,
		"state":                 state,
		"code_challenge_method": codeChallengeMethodS256,
	})
}

func (s *Server) readSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(oauthSessionCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request, provider string) {
	code := r.URL.Query().Get("code")
	if code == "" {
		s.writeError(w, http.StatusBadRequest, errors.New("missing code parameter"))
		return
	}

	reqState := r.URL.Query().Get("state")
	if reqState == "" {
		s.writeError(w, http.StatusBadRequest, errors.New("missing state parameter"))
		return
	}

	sessionID, err := s.readSessionID(r)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, errors.New("missing oauth session cookie"))
		return
	}

	savedSession, ok := s.sessions.Consume(sessionID)
	if !ok {
		s.writeError(w, http.StatusBadRequest, errors.New("oauth session expired or not found"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	if !subtleConstantTimeCompare(reqState, savedSession.State) {
		s.writeError(w, http.StatusBadRequest, errors.New("state parameter mismatch"))
		return
	}

	queryTenant := r.URL.Query().Get("tenant")
	tenant := savedSession.Tenant
	if queryTenant != "" {
		if tenant != "" && tenant != queryTenant {
			s.writeError(w, http.StatusBadRequest, errors.New("tenant mismatch"))
			return
		}
		tenant = queryTenant
	}

	session, err := s.auth.ProviderSession(r.Context(), tenant, provider)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err)
		return
	}

	token, err := session.ExchangeCode(r.Context(), service.ExchangeParams{
		Code:         code,
		CodeVerifier: savedSession.CodeVerifier,
	})
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err)
		return
	}

	profile, err := session.UserProfile(r.Context(), token)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"tenant":   tenant,
		"provider": provider,
		"profile":  profile,
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}
