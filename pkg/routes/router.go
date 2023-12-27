package routes

import (
	"net/http"

	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/services"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

const (
	routeNameForgotPassword       = "forgot_password"
	routeNameForgotPasswordSubmit = "forgot_password.submit"
	routeNameLogin                = "login"
	routeNameLoginWithGoogle      = "login_with_google"
	routeNameLoginGoogleCallback  = "login_with_google_callback"
	routeNameLoginSubmit          = "login.submit"
	routeNameLogout               = "logout"
	routeNameRegister             = "register"
	routeNameRegisterSubmit       = "register.submit"
	routeNameResetPassword        = "reset_password"
	routeNameResetPasswordSubmit  = "reset_password.submit"
	routeNameVerifyEmail          = "verify_email"
	routeNameContact              = "contact"
	routeNameContactSubmit        = "contact.submit"
	routeNameContactDelete        = "contact.delete"
	routeNameAbout                = "about"
	routeNameHome                 = "home"
	routeNameSearch               = "search"
	routeNamePost                 = "post"
	routeNamePostSubmit           = "post.submit"
	routeNamePostUpdate           = "post.update"
	routeNamePostDelete           = "post.delete"
	routeNamePostUpload           = "post.upload"
	routeNameUser                 = "user"
	routeNameUserUpdate           = "user.update"
	routeNameUserDelete           = "user.delete"
)

// BuildRouter builds the router
func BuildRouter(c *services.Container) {
	// Static files with proper cache control
	// funcmap.File() should be used in templates to append a cache key to the URL in order to break cache
	// after each server restart

	c.Web.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
		AllowCredentials: true,
	}))
	c.Web.Group("", middleware.CacheControl(c.Config.Cache.Expiration.StaticFile)).
		Static(config.StaticPrefix, config.StaticDir)

	// Non static file route group
	g := c.Web.Group("")

	// Force HTTPS, if enabled
	if c.Config.HTTP.TLS.Enabled {
		g.Use(echomw.HTTPSRedirect())
	}

	g.Use(
		echomw.RemoveTrailingSlashWithConfig(echomw.TrailingSlashConfig{
			RedirectCode: http.StatusMovedPermanently,
		}),
		echomw.Recover(),
		echomw.Secure(),
		echomw.RequestID(),
		echomw.Gzip(),
		echomw.Logger(),
		middleware.LogRequestID(),
		echomw.TimeoutWithConfig(echomw.TimeoutConfig{
			Timeout: c.Config.App.Timeout,
		}),
		session.Middleware(sessions.NewCookieStore([]byte(c.Config.App.EncryptionKey))),
		middleware.LoadAuthenticatedUser(c.Auth),
		middleware.ServeCachedPage(c.Cache),
		echomw.CSRFWithConfig(echomw.CSRFConfig{
			// TokenLookup: "form:csrf",
			Skipper: func(ctx echo.Context) bool {
				return true
			},
		}),
	)

	// Base controller
	ctr := controller.NewController(c)

	// Error handler
	err := errorHandler{Controller: ctr}
	c.Web.HTTPErrorHandler = err.Get

	// Example routes
	navRoutes(c, g, ctr)
	userRoutes(c, g, ctr)
	postRoutes(c, g, ctr)
}

func navRoutes(c *services.Container, g *echo.Group, ctr controller.Controller) {
	home := home{Controller: ctr}
	g.Static("/static", config.StaticDir)
	g.GET("/", home.Get).Name = routeNameHome

	search := search{Controller: ctr}
	g.GET("/search", search.Get).Name = routeNameSearch

	about := about{Controller: ctr}
	g.GET("/about", about.Get).Name = routeNameAbout

	contact := contact{Controller: ctr}
	g.GET("/contact", contact.Get).Name = routeNameContact
	g.POST("/contact", contact.Post).Name = routeNameContactSubmit
	g.DELETE("/contact/delete/:id", contact.Delete).Name = routeNameContactDelete
}

func userRoutes(c *services.Container, g *echo.Group, ctr controller.Controller) {
	logout := logout{Controller: ctr}
	g.GET("/logout", logout.Get, middleware.RequireAuthentication()).Name = routeNameLogout

	verifyEmail := verifyEmail{Controller: ctr}
	g.GET("/email/verify/:token", verifyEmail.Get).Name = routeNameVerifyEmail

	noAuth := g.Group("/user", middleware.RequireNoAuthentication())
	login := login{Controller: ctr}
	noAuth.GET("/login_with_google", login.LoginWithGoogle).Name = routeNameLoginWithGoogle
	g.GET("/api/v1/google/callback", login.GetCallback).Name = routeNameLoginGoogleCallback
	noAuth.GET("/login", login.Get).Name = routeNameLogin
	noAuth.POST("/login", login.Post).Name = routeNameLoginSubmit

	Auth := g.Group("/users", middleware.RequireAuthentication())
	user := _user_{Controller: ctr}
	Auth.GET("", user.Get).Name = routeNameUser
	Auth.PUT("/update/:id", user.Put).Name = routeNameUserUpdate

	register := register{Controller: ctr}
	noAuth.GET("/register", register.Get).Name = routeNameRegister
	noAuth.POST("/register", register.Post).Name = routeNameRegisterSubmit

	forgot := forgotPassword{Controller: ctr}
	noAuth.GET("/password", forgot.Get).Name = routeNameForgotPassword
	noAuth.POST("/password", forgot.Post).Name = routeNameForgotPasswordSubmit

	resetGroup := noAuth.Group("/password/reset",
		middleware.LoadUser(c.ORM),
		middleware.LoadValidPasswordToken(c.Auth),
	)
	reset := resetPassword{Controller: ctr}
	resetGroup.GET("/token/:user/:password_token/:token", reset.Get).Name = routeNameResetPassword
	resetGroup.POST("/token/:user/:password_token/:token", reset.Post).Name = routeNameResetPasswordSubmit
}

func postRoutes(c *services.Container, g *echo.Group, ctr controller.Controller) {
	noAuth := g.Group("/api/v1", middleware.RequireNoAuthentication())

	post := post{Controller: ctr}
	noAuth.GET("/posts", post.GetPosts)

	Auth := g.Group("/post", middleware.RequireAuthentication())
	Auth.GET("/create", post.Get).Name = routeNamePost
	Auth.POST("/create", post.Post).Name = routeNamePostSubmit
	Auth.GET("/edit/:id", post.GetUpdate).Name = routeNamePostSubmit
	Auth.PUT("/edit/:id", post.Update).Name = routeNamePostUpdate
	Auth.DELETE("/delete/:id", post.Delete).Name = routeNamePostDelete
	Auth.POST("/upload", post.Upload).Name = routeNamePostUpload

}
