package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/replay/free-my-lists/pkg/config"
	"golang.org/x/oauth2"
)

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

	w.router.GET("/", w.indexHandler)
	w.router.GET("/login", w.loginHandler)
	w.router.GET("/auth", w.authHandler)

	w.private = w.router.Group("/members")
	w.private.Use(w.requireLogin())
	w.private.GET("/", w.welcomeHandler)

	w.router.Run(":8080")

	return w
}

func (w *Web) requireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		tok := session.Get("token")
		if tok == nil {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: "))
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
	c.Writer.Write([]byte("<html><title>Golang Google</title> <body> <a href='" + w.cfg.OauthProviders.Google.AuthCodeURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
}

func (w *Web) authHandler(c *gin.Context) {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", retrievedState))
		return
	}

	tok, err := w.cfg.OauthProviders.Google.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tokSerialized, err := json.Marshal(tok)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session.Set("token", tokSerialized)
	session.Save()

	client := w.cfg.OauthProviders.Google.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()

	c.Redirect(http.StatusFound, w.cfg.Domain+"/members")
}

func (w *Web) indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Main website",
	})
}

func (w *Web) welcomeHandler(c *gin.Context) {
	session := sessions.Default(c)
	tokSerialized := session.Get("token").([]byte)
	var tok *oauth2.Token
	err := json.Unmarshal(tokSerialized, &tok)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := w.cfg.OauthProviders.Google.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	c.Writer.Write([]byte("<html><title>Golang Google</title> <body> User info: " + string(data) + "</body></html>"))
}
