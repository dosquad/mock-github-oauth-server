package mockghauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/na4ma4/config"
)

const (
	defaultTimeout = 10 * time.Second
)

type Server struct {
	baseURL *url.URL
	clients *Clients
	codes   *Codes
	tokens  *Tokens
	g       *gin.Engine
}

//nolint:forbidigo // panic error.
func NewServer(baseURL *url.URL, cfg config.Conf) *Server {
	g := gin.Default() // listen on 0.0.0.0:8080
	codes := &Codes{}
	clients := NewClients()
	tokens := &Tokens{}

	s := &Server{
		baseURL: baseURL,
		g:       g,
		codes:   codes,
		tokens:  tokens,
		clients: clients,
	}

	if filename := cfg.GetString("load.code-file"); filename != "" {
		if err := codes.ReadFile(filename); err != nil {
			fmt.Printf("unable to load file[%s]: %s\n", filename, err)
			panic(err)
		}
	}

	if filename := cfg.GetString("load.clients-file"); filename != "" {
		if err := clients.ReadFile(filename); err != nil {
			fmt.Printf("unable to load file[%s]: %s\n", filename, err)
			panic(err)
		}
	}

	if filename := cfg.GetString("load.tokens-file"); filename != "" {
		if err := tokens.ReadFile(filename); err != nil {
			fmt.Printf("unable to load file[%s]: %s\n", filename, err)
			panic(err)
		}
	}

	g.GET("/login/oauth/authorize", s.loginOauthAuthorize)
	g.POST("/login/oauth/access_token", s.loginOauthAccessToken)
	g.GET("/api/v3/user", s.apiV3User)

	return s
}

func (s *Server) loginOauthAuthorize(c *gin.Context) {
	clientID, clientIDExists := c.GetQuery("client_id")
	if !clientIDExists || !s.clients.HasID(clientID) {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	redirectURI, redirectURIExists := c.GetQuery("redirect_uri")
	if !redirectURIExists {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	code := s.codes.New()

	c.Redirect(http.StatusFound, fmt.Sprintf("%s?code=%s", redirectURI, code))
}

func (s *Server) loginOauthAccessToken(c *gin.Context) {
	var oauthReq GitHubOAuth

	if err := c.ShouldBind(&oauthReq); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !s.clients.HasID(oauthReq.ClientID) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !s.codes.Exists(oauthReq.Code) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	token := s.tokens.New()

	resp := &GitHubOAuthResponse{
		AccessToken: token,
		Scope:       "repo,admin",
		TokenType:   "bearer",
	}

	c.JSON(http.StatusOK, resp)
}

//nolint:mnd // get everything after first space in Authorization header.
func (s *Server) checkAuthIsValid(c *gin.Context) bool {
	if authHeader := c.Request.Header.Get("Authorization"); authHeader != "" {
		spHeader := strings.SplitN(authHeader, " ", 2)
		if len(spHeader) != 2 {
			return false
		}
		if !s.tokens.Exists(spHeader[1]) {
			return false
		}
	} else {
		return false
	}

	return true
}

func (s *Server) apiV3User(c *gin.Context) {
	if !s.checkAuthIsValid(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, UnauthorizedGitHubAPIError())
		return
	}

	user, err := DefaultGitHubAPIUser(s.baseURL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UnauthorizedGitHubAPIError(err))
		return
	}

	c.JSON(http.StatusOK, user)
}

func (s *Server) AddClient(id, secret string) {
	s.clients.Add(id, secret)
}

func (s *Server) Reaper(ts time.Time) {
	s.tokens.Reaper(ts)
}

func (s *Server) Run(ctx context.Context) error {
	address := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		address = ":" + port
	}

	srv := &http.Server{
		Addr:              address,
		Handler:           s.g.Handler(),
		ReadTimeout:       defaultTimeout,
		ReadHeaderTimeout: defaultTimeout,
		WriteTimeout:      defaultTimeout,
		IdleTimeout:       defaultTimeout,
	}

	doneCh := make(chan error)
	go func() {
		doneCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-doneCh:
		return err
	}
}
