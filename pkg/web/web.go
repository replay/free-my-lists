package web

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/replay/free-my-lists/pkg/config"
	"github.com/replay/free-my-lists/pkg/provider"
	"github.com/replay/free-my-lists/pkg/web/token"
)

const tokenCookie = "access_token"

type Web struct {
	cfg     config.Config
	store   sessions.CookieStore
	router  *gin.Engine
	private *gin.RouterGroup
}

func New(cfg config.Config) Web {
	w := Web{
		cfg:   cfg,
		store: sessions.NewCookieStore([]byte("secret")),
	}

	w.router = gin.New()
	w.router.Use(gin.Recovery())
	w.router.Use(sessions.Sessions("goquestsession", w.store))
	w.router.LoadHTMLGlob(cfg.Templates)

	// Logged out area.
	w.router.GET("/", w.indexHandler)
	w.router.GET("/login", w.loginHandler)
	w.router.GET("/logout", w.logoutHandler)
	w.router.GET("/auth/google", w.authGoogleHandler)
	w.router.GET("/auth/spotify", w.authSpotifyHandler)

	// Logged in area.
	w.private = w.router.Group("/members")
	w.private.Use(w.requireLogin())
	w.private.GET("/", w.mainHandler)
	w.private.GET("/lists", w.listsHandler)

	w.router.Run(":8080")

	return w
}

func (w *Web) requireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		tok := session.Get(tokenCookie)
		if tok == nil {
			c.Abort()
			c.Redirect(http.StatusFound, w.cfg.Domain)
			return
		}
		c.Next()
	}
}

func (w *Web) Shutdown() {
	fmt.Println("shutting down")
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (w *Web) loginHandler(c *gin.Context) {
	state := randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"loginSpotify": w.cfg.OauthProviders.Spotify.AuthCodeURL(state),
		"loginGoogle":  w.cfg.OauthProviders.Google.AuthCodeURL(state),
	})
}

func (w *Web) logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(tokenCookie)
	session.Save()
	c.Redirect(http.StatusFound, w.cfg.Domain)
}

func (w *Web) authGoogleHandler(c *gin.Context) {
	w.authHandler(c, token.Google)
}

func (w *Web) authSpotifyHandler(c *gin.Context) {
	w.authHandler(c, token.Spotify)
}

func (w *Web) authHandler(c *gin.Context, providerType token.Type) {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	tok, err := provider.Config(w.cfg, providerType).Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	serialized, err := token.NewToken(tok, providerType).Serialize()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session.Set(tokenCookie, serialized)
	session.Save()
	c.Redirect(http.StatusFound, w.cfg.Domain+"/members")
}

func (w *Web) indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Main website",
	})
}

func (w *Web) mainHandler(c *gin.Context) {
	client, err := w.authProvider(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	userInfo, err := client.UserInfo(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.HTML(http.StatusOK, "main.tmpl", gin.H{
		"userInfo": userInfo,
	})
}

func (w *Web) listsHandler(c *gin.Context) {
	p, err := w.listsProvider(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	lists, err := p.Lists(c.Request.Context())
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.HTML(http.StatusOK, "lists.tmpl", gin.H{
		"lists": lists,
	})
}

func (w *Web) authProvider(c *gin.Context) (provider.AuthProvider, error) {
	session := sessions.Default(c)
	tokSerialized := session.Get(tokenCookie).([]byte)
	t, err := token.Deserialize(tokSerialized)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return nil, fmt.Errorf("failed to deserialize token: %v", err)
	}

	return provider.NewAuthProvider(c.Request.Context(), w.cfg, t)
}

func (w *Web) listsProvider(c *gin.Context) (provider.ListsProvider, error) {
	authProvider, err := w.authProvider(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return nil, fmt.Errorf("failed to get auth provider: %v", err)
	}

	return authProvider.ListsProvider()
}
