package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/user"
	"github.com/mikestefanello/pagoda/pkg/context"
	cctx "context"
	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/pkg/msg"
	"github.com/mikestefanello/pagoda/templates"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/labstack/echo/v4"
)

type (
	login struct {
		controller.Controller
	}

	loginForm struct {
		Email      string `form:"email" validate:"required,email"`
		Password   string `form:"password" validate:"required"`
		Submission controller.FormSubmission
	}
)

func (c *login) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutAuth
	page.Name = templates.PageLogin
	page.Title = "Log in"
	page.Form = loginForm{}

	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*loginForm)
	}

	return c.RenderPage(ctx, page)
}

func (c *login) Post(ctx echo.Context) error {
	var form loginForm
	ctx.Set(context.FormKey, &form)

	authFailed := func() error {
		form.Submission.SetFieldError("Email", "")
		form.Submission.SetFieldError("Password", "")
		msg.Danger(ctx, "Invalid credentials. Please try again.")
		return c.Get(ctx)
	}

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		return c.Fail(err, "unable to parse login form")
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return c.Fail(err, "unable to process form submission")
	}

	if form.Submission.HasErrors() {
		return c.Get(ctx)
	}

	// Attempt to load the user
	u, err := c.Container.ORM.User.
		Query().
		Where(user.Email(strings.ToLower(form.Email))).
		Only(ctx.Request().Context())

	switch err.(type) {
	case *ent.NotFoundError:
		return authFailed()
	case nil:
	default:
		return c.Fail(err, "error querying user during login")
	}

	// Check if the password is correct
	err = c.Container.Auth.CheckPassword(form.Password, u.Password)
	if err != nil {
		return authFailed()
	}

	// Log the user in
	err = c.Container.Auth.Login(ctx, u.ID)
	if err != nil {
		return c.Fail(err, "unable to log in user")
	}

	msg.Success(ctx, fmt.Sprintf("Welcome back, <strong>%s</strong>. You are now logged in.", u.Name))
	return c.Redirect(ctx, routeNameHome)
}

func (c *login) LoginWithGoogle(ctx echo.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	googleOAuthConfig := oauth2.Config{
	ClientID:     cfg.Google.ClientID,
	ClientSecret: cfg.Google.ClientSecret,
	RedirectURL:  cfg.Google.RedirectURL,
	Scopes:       []string{
	 "https://www.googleapis.com/auth/userinfo.profile", 
	 "https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}
	authURL := googleOAuthConfig.AuthCodeURL("test", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(ctx.Response(), ctx.Request(), authURL, http.StatusFound)
	return nil
}

func (c *login) GetCallback(ctx echo.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	googleOAuthConfig := oauth2.Config{
	ClientID:     cfg.Google.ClientID,
	ClientSecret: cfg.Google.ClientSecret,
	RedirectURL:  cfg.Google.RedirectURL,
	Scopes:       []string{
	 "https://www.googleapis.com/auth/userinfo.profile", 
	 "https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}
	code := ctx.QueryParam("code")
	tok, err := googleOAuthConfig.Exchange(ctx.Request().Context(), code)
	if err != nil {
		panic(fmt.Sprintf("failed to exchange token: %v", err))
	}

	httpClient := googleOAuthConfig.Client(ctx.Request().Context(), tok)

	userInfo, err := getUserInfo(httpClient)
	if err != nil {
		return err
	}
	// if !strings.HasSuffix(userInfo.Email, "@appota.com") {
	// 	msg.Danger(ctx, "You must use Appota email to login")
	// 	return c.Redirect(ctx, routeNameHome)
	// }

	u, err := c.Container.ORM.User.
		Query().
		Where(user.Email(strings.ToLower(userInfo.Email))).
		Only(ctx.Request().Context())
	
	switch err.(type) {
	case *ent.NotFoundError:
		// create user
		password, _ := c.Container.Auth.HashPassword(fmt.Sprintf("%s%s", userInfo.Id, userInfo.Email))
		u, err = c.Container.ORM.User.
			Create().
			SetName(userInfo.Name).
			SetEmail(userInfo.Email).
			SetPassword(password).
			SetVerified(true).
			Save(ctx.Request().Context())
		if err != nil {
			return err
		}
	case nil:
	default:
		return c.Fail(err, "error querying user during login")
	}

	// Log the user in
	err = c.Container.Auth.Login(ctx, u.ID)
	if err != nil {
		return c.Fail(err, "unable to log in user")
	}

	msg.Success(ctx, fmt.Sprintf("Welcome back, <strong>%s</strong>. You are now logged in.", u.Name))
	return c.Redirect(ctx, routeNameHome)
}

func getUserInfo(httpClient *http.Client) (*googleoauth2.Userinfo, error) {
	oauth2Service, err := googleoauth2.NewService(cctx.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	oauth2UserInfo, err := oauth2Service.Userinfo.V2.Me.Get().Do()
	if err != nil {
		return nil, err
	}

	return oauth2UserInfo, nil
}

