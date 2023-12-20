package routes

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)

type (
	home struct {
		controller.Controller
	}

	_post_ struct {
		Title string
		Body  string
	}
)

func (c *home) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageHome
	page.Metatags.Description = "Welcome to the homepage."
	page.Metatags.Keywords = []string{"Go", "MVC", "Web", "Software"}
	page.Pager = controller.NewPager(ctx, 4)
	page.Data = c.fetchPosts_(ctx, &page.Pager)

	return c.RenderPage(ctx, page)
}

func (c *home) fetchPosts_(ctx echo.Context, pager *controller.Pager) []_post_ {

	total, p := getPosts(c.Controller, ctx, pager)
	pager.SetItems(total)
	posts := make([]_post_, total)
	for k, v := range p {
		if len(v.Title) > 30 {
			v.Title = v.Title[:30] + "..."
		}
		if len(v.Body) > 80 {
			v.Body = v.Body[:80] + "..."
		}
		posts[k] = _post_{
			Title: fmt.Sprintf("%s", v.Title),
			Body:  fmt.Sprintf("%s", v.Body),
		}
	}
	return posts[:pager.ItemsPerPage]
}
