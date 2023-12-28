package routes

import (
	"errors"
	"strconv"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/controller"

	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)

type (
	_user_ struct {
		controller.Controller
	}
)

func (c *_user_) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageUser
	page.Metatags.Description = "Users Page"
	page.Pager = controller.NewPager(ctx, 4)
	page.Data = c.fetchUsers(ctx, &page.Pager)

	return c.RenderPage(ctx, page)
}

func (c *_user_) Put(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil {
		return c.Fail(err, "invalid user id")
	}
	req := struct {
		Permission string `form:"permission"`
	}{}
	if err := ctx.Bind(&req); err != nil {
		return c.Fail(err, "unable to bind form")
	}
	_, err = c.Container.ORM.User.
		UpdateOneID(id).
		SetPermission(req.Permission).
		Save(ctx.Request().Context())
	if err != nil {
		return c.Fail(err, "unable to save user")
	}

	// set method to get before redirecting
	ctx.Request().Method = "GET"

	return c.routeToUserPage(ctx, page)
}

func (c *_user_) fetchUsers(ctx echo.Context, pager *controller.Pager) []ent.User {

	total, u := getUsers(c.Controller, ctx, pager)
	pager.SetItems(total)
	if pager.Page < 1 {
		pager.Page = 1
	}
	users := make([]ent.User, len(u))
	for k, v := range u {
		if len(v.Name) > 30 {
			v.Name = v.Name[:30] + "..."
		}
		if len(v.Email) > 80 {
			v.Email = v.Email[:80] + "..."
		}
		users[k] = ent.User{
			ID:    v.ID,
			Name: v.Name,
			Email:  v.Email,
			Permission: v.Permission,
		}
	}
	return users
}

func getUsers(c controller.Controller, ctx echo.Context, pager *controller.Pager) (int, []*ent.User) {
	total, err := c.Container.ORM.User.
		Query().
		Count(ctx.Request().Context())
   if err != nil {
		return 0, nil
   }
	users, err := c.Container.ORM.User.
		Query().
		Offset(pager.GetOffset()).
		Limit(pager.ItemsPerPage).
		Order(ent.Desc("created_at")).
		All(ctx.Request().Context())
	if err != nil {
		return 0, nil
	}

	return total, users
}

func (c *_user_) routeToUserPage(ctx echo.Context, p int) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageUser
	page.Metatags.Description = "Users Page"
	page.Pager = controller.NewPager(ctx, 4)
	page.Pager.Page = p
	page.Data = c.fetchUsers(ctx, &page.Pager)

	return c.RenderPage(ctx, page)
}

