package digipoauth

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	oautherrs "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	digiconfig "github.com/holyhope/digiposte-oauth/config"
)

const (
	// AuthorizePath is the path to the authorize endpoint.
	AuthorizePath = "/authorize"
	// TokenPath is the path to the token endpoint.
	TokenPath = "/token"

	// ReadTimeout is the timeout for reading the request.
	ReadTimeout = 5 * time.Second
	// WriteTimeout is the timeout for writing the response.
	WriteTimeout = 5 * time.Minute
)

// Server is a local web server for collecting auth.
type Server struct {
	server      *http.Server
	listener    net.Listener
	manager     *manage.Manager
	clientStore *store.ClientStore
}

// StartServer starts a local webserver to receive the auth.
func NewServer(setter digiconfig.Setter, addr string, loginMethod LoginMethod, logger *log.Logger) (*Server, error) {
	// client memory store
	clientStore := store.NewClientStore()

	manager := newManager(clientStore, setter, loginMethod)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &Server{
		server:      newServer(manager, listener, logger),
		listener:    listener,
		manager:     manager,
		clientStore: clientStore,
	}, nil
}

// RegisterUser adds a user to the server.
func (s *Server) RegisterUser(clientID, clientSecret, redirectURL string) error {
	if err := s.clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		UserID: clientID,
		Public: false,
		Domain: redirectURL,
	}); err != nil {
		return fmt.Errorf("set client: %w", err)
	}

	return nil
}

func newServer(manager oauth2.Manager, listener net.Listener, logger *log.Logger) *http.Server {
	mux := http.NewServeMux()
	oauthServer := server.NewDefaultServer(manager)

	mux.HandleFunc(AuthorizePath, func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleAuthorizeRequest(w, r); err != nil {
			logger.Printf("Failed to handle authorize request: %v", err)
		}
	})
	mux.HandleFunc(TokenPath, func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleTokenRequest(w, r); err != nil {
			logger.Printf("Failed to handle token request: %v", err)
		}
	})

	oauthServer.SetAllowGetAccessRequest(true)

	oauthServer.UserAuthorizationHandler = func(w http.ResponseWriter, r *http.Request) (string, error) {
		if err := r.ParseForm(); err != nil {
			return "", oautherrs.ErrInvalidRequest
		}

		return r.Form.Get("client_id"), nil
	}

	oauthServer.ClientInfoHandler = func(r *http.Request) (string, string, error) {
		if err := r.ParseForm(); err != nil {
			return "", "", oautherrs.ErrInvalidRequest
		}

		return r.Form.Get("client_id"), r.Form.Get("client_secret"), nil
	}

	httpServer := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           mux,
		ReadHeaderTimeout: ReadTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       WriteTimeout,

		TLSConfig:      nil,
		MaxHeaderBytes: 0,
		BaseContext:    nil,
		TLSNextProto:   nil,
		ConnState:      nil,
		ErrorLog:       nil,
		ConnContext:    nil,
	}

	return httpServer
}

func newManager(clientStore oauth2.ClientStore, setter digiconfig.Setter, loginMethod LoginMethod) *manage.Manager {
	manager := manage.NewDefaultManager()

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapClientStorage(clientStore)
	manager.MapAccessGenerate(&AccessGenerator{
		setter: setter,
		login:  loginMethod,
	})
	manager.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:     time.Hour,
		IsGenerateRefresh:  true,
		RefreshTokenExp:    0,
		IsResetRefreshTime: false,
		IsRemoveAccess:     false,
		IsRemoveRefreshing: true,
	})

	return manager
}

// Close closes the server.
func (s *Server) Close() error {
	return s.server.Close() //nolint:wrapcheck
}

// Start starts the server.
func (s *Server) Start() error {
	if err := s.server.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

// AuthorizeURL returns the URL to the authorize endpoint.
func (s *Server) AuthorizeURL() string {
	return "http://" + s.listener.Addr().String() + AuthorizePath
}

// TokenURL returns the URL to the token endpoint.
func (s *Server) TokenURL() string {
	return "http://" + s.listener.Addr().String() + TokenPath
}
