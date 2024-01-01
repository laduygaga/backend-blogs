package routes

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mikestefanello/pagoda/ent"
	post_ "github.com/mikestefanello/pagoda/ent/post"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/controller"

	// "github.com/mikestefanello/pagoda/pkg/msg"
	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)

type (
	post struct {
		controller.Controller
	}
	postForm struct {
		ID    int    `form:"id"`
		Title string `form:"title" validate:"required"`
		Body  string `form:"body" validate:"required"`
		Page  int    `form:"page"`
		Submission controller.FormSubmission
	}
)

func (c *post ) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PagePost
	page.Title = "Post"
	page.Form = postForm{}
	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*postForm)
	}

	return c.RenderPage(ctx, page)
}

func (c *post ) GetUpdate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return c.Fail(err, "unable to update post")
	}
	post, err := c.Container.ORM.Post.Get(ctx.Request().Context(), id)
	if err != nil {
		return c.Fail(err, "unable to update post")
	}

	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PagePostUpdate
	page.Title = "Update"
	page.Form = postForm{
		ID:    post.ID,
		Title: post.Title,
		Body:  post.Body,
		Page:  page.Pager.Page,
	}
	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*postForm)
	}

	return c.RenderPage(ctx, page)
}

func (c *post ) Post(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}
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

func (c *post ) Update(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return c.Fail(err, "unable to update post")
	}

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
	_, err = c.Container.ORM.Post.UpdateOneID(id).
		SetTitle(strings.TrimSpace(form.Title)).
		SetBody(strings.TrimSpace(form.Body)).
		SetAuthor(user.Email).
		Save(ctx.Request().Context())
	if err != nil {
		return c.Fail(err, "unable to save post")
	}

	return routeToPostPage(c.Controller, ctx, form.Page)
}

func (c *post ) Delete(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}

	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		return c.Fail(nil, "unable to delete post")
	}
	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil {
		return c.Fail(err, "invalid page number")
	}
	if err := c.Container.ORM.Post.DeleteOneID(id).Exec(ctx.Request().Context()); err != nil {
		return c.Fail(err, "unable to delete post")
	}

	// msg.Info(ctx, fmt.Sprintf("Post %d deleted", id))
	return routeToPostPage(c.Controller, ctx, page)
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

// upload file to static/uploads
func (c *post ) Upload(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}

	file, err := ctx.FormFile("upload")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"Status": err.Error()})
	}

	// Open the file from the temporary location
	src, err := file.Open()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"Status": err.Error()})
	}
	defer src.Close()

	// Create a new file in the destination
	// check if donot have folder static/uploads then create it
	if _, err := os.Stat("./static/uploads"); os.IsNotExist(err) {
		os.Mkdir("./static/uploads", 0755)
	}
	dst, err := os.Create("./static/uploads/" + file.Filename)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"Status": err.Error()})
	}
	defer dst.Close()

	// Copy the file content to the destination
	if _, err = io.Copy(dst, src); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"Status": err.Error()})
	}

	// Return success response
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"uploaded":  1,
		"fileName":  file.Filename,
		"url":       "/admin/static/uploads/" + file.Filename,
	})
}

func (c *post ) GetPosts(ctx echo.Context) error {
	from := ctx.QueryParam("from")
	to := ctx.QueryParam("to")
	if from != "" && to != "" {
		return c.GetPostsByDate(ctx)
	}
	page := controller.NewPage(ctx)
	page.Pager = controller.NewPager(ctx, 4)
	total, posts := getPosts(c.Controller, ctx, &page.Pager)

	page.Pager.SetItems(total)
	if page.Pager.Page < 1 {
		page.Pager.Page = 1
	}
	// return posts as json
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"posts": posts,
		"pager": page.Pager,
	})
}

func (c *post ) GetPostsByDate(ctx echo.Context) error {
	from := ctx.QueryParam("from")
	to := ctx.QueryParam("to")
	page := controller.NewPage(ctx)
	page.Pager = controller.NewPager(ctx, 4)
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return c.Fail(err, "invalid from date")
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return c.Fail(err, "invalid to date")
	}
	total, posts := getPostsByDate(c.Controller, ctx, &page.Pager, fromDate, toDate)

	page.Pager.SetItems(total)
	if page.Pager.Page < 1 {
		page.Pager.Page = 1
	}
	// return posts as json
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"posts": posts,
		"pager": page.Pager,
	})
}

func getPostsByDate(c controller.Controller, ctx echo.Context, pager *controller.Pager, from, to time.Time) (int, []*ent.Post) {
	to = to.AddDate(0, 0, 1)
	total, err := c.Container.ORM.Post.
		Query().
		Where(post_.CreatedAtGTE(from)).
		Where(post_.CreatedAtLTE(to)).
		Count(ctx.Request().Context())
   if err != nil {
		return 0, nil
   }
	posts, err := c.Container.ORM.Post.
		Query().
		Where(post_.CreatedAtGTE(from)).
		Where(post_.CreatedAtLTE(to)).
		Offset(pager.GetOffset()).
		Limit(pager.ItemsPerPage).
		Order(ent.Desc(post_.FieldID)).
		All(ctx.Request().Context())
	if err != nil {
		return 0, nil
	}

	return total, posts
}
