package routes

import (
	"strconv"
	"strings"

	"github.com/mikestefanello/pagoda/ent"
	post_ "github.com/mikestefanello/pagoda/ent/post"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)

type (
	post struct {
		controller.Controller
	}
	postForm struct {
		Title string `form:"title" validate:"required"`
		Body  string `form:"body" validate:"required"`
		Submission controller.FormSubmission
	}
)

func (c *post ) Get(ctx echo.Context) error {
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

func (c *post ) Post(ctx echo.Context) error {
	var form postForm
	ctx.Set(context.FormKey, &form)
	if err := ctx.Bind(&form); err != nil {
		return c.Fail(err, "unable to parse post form")
	}
	if err := form.Submission.Process(ctx, form); err != nil {
		return c.Fail(err, "unable to process form submission")
	}
	if form.Submission.HasErrors() {
		return c.Get(ctx)
	}
	user := ctx.Get("auth_user").(*ent.User)
	_, err := c.Container.ORM.Post.Create().
		SetTitle(strings.TrimSpace(form.Title)).
		SetBody(strings.TrimSpace(form.Body)).
		SetAuthor(user.Email).
		Save(ctx.Request().Context())
	if err != nil {
		return c.Fail(err, "unable to save post")
	}

	return c.Redirect(ctx, routeNameHome)
}

func (c *post ) Delete(ctx echo.Context) error {

	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		return c.Fail(nil, "unable to delete post")
	}
	if err := c.Container.ORM.Post.DeleteOneID(id).Exec(ctx.Request().Context()); err != nil {
		return c.Fail(err, "unable to delete post")
	}

	return c.Redirect(ctx, routeNameHome)
}

func getPosts(c controller.Controller, ctx echo.Context, pager *controller.Pager) (int, []*ent.Post) {
	total, err := c.Container.ORM.Post.
		Query().
		Count(ctx.Request().Context())
   if err != nil {
		return 0, nil
   }
	posts, err := c.Container.ORM.Post.
		Query().
		Offset(pager.GetOffset()).
		Limit(pager.ItemsPerPage).
		Order(ent.Desc(post_.FieldID)).
		All(ctx.Request().Context())
	if err != nil {
		return 0, nil
	}

	return total, posts
}

