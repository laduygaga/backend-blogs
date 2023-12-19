package routes

import (
	"fmt"
	"strings"

	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)


type (
	Post struct {
		controller.Controller
	}

	postForm struct {
		Title string `form:"title" validate:"required"`
		Body  string `form:"body" validate:"required"`
		Submission controller.FormSubmission
	}
)

func (c *Post ) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutAuth
	page.Name = templates.PagePost
	page.Title = "Post"
	page.Form = postForm{}

	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*postForm)
	}

	return c.RenderPage(ctx, page)
}

func (c *Post ) Post(ctx echo.Context) error {
	var form postForm
	ctx.Set(context.FormKey, &form)

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		return c.Fail(err, "unable to parse post form")
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return c.Fail(err, "unable to process form submission")
	}

	if form.Submission.HasErrors() {
		return c.Get(ctx)
	}

	_, err := c.Container.ORM.Post.Create().
		SetTitle(strings.TrimSpace(form.Title)).
		SetBody(strings.TrimSpace(form.Body)).
		SetAuthor("duynn").
		Save(ctx.Request().Context())

	if err != nil {
		return c.Fail(err, "unable to save post")
	}

	return c.Redirect(ctx, routeNameHome)
}

// fetchPosts is an mock example of fetching posts to illustrate how paging works
func (c *Post) FetchPosts(pager *controller.Pager) []post {
	pager.SetItems(20)
	posts := make([]post, 20)

	for k := range posts {
		posts[k] = post{
			Title: fmt.Sprintf("Post example #%d", k+1),
			Body:  fmt.Sprintf("Lorem ipsum example #%d ddolor sit amet, consectetur adipiscing elit. Nam elementum vulputate tristique.", k+1),
		}
	}
	return posts[pager.GetOffset() : pager.GetOffset()+pager.ItemsPerPage]
}
