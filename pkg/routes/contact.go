package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/controller"
	"github.com/mikestefanello/pagoda/templates"

	"github.com/labstack/echo/v4"
)

type (
	contact struct {
		controller.Controller
	}

	contactForm struct {
		Type	   string `form:"type" validate:"required"`
		Link	   string `form:"link" validate:"required"`
		Email	   string `form:"email" validate:"required,email"`
		Message    string `form:"message" validate:"required"`
		Submission controller.FormSubmission
	}
)

func (c *contact) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageContact
	page.Title = "Recent Contact"
	page.Pager = controller.NewPager(ctx, 4)
	page.Data = fetchContacts(c.Controller, ctx, &page.Pager)

	return c.RenderPage(ctx, page)
}

func fetchContacts(c controller.Controller, ctx echo.Context, pager *controller.Pager) []ent.Contact {
	total, p := getContacts(c, ctx, pager)
	pager.SetItems(total)
	if pager.Page < 1 {
		pager.Page = 1
	}
	contacts := make([]ent.Contact, len(p))
	for k, v := range p {
		contacts[k] = ent.Contact{
			ID: v.ID,
			Email: v.Email,
			Link: v.Link,
			Type: v.Type,
			Message: v.Message,
		}
	}
	return contacts
}

func (c *contact) Delete(ctx echo.Context) error {
	if ctx.Get("auth_user").(*ent.User).Permission != "Editor" {
		return c.Fail(errors.New("Permission Error"), "do not have permission to delete contact")
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return c.Fail(err, "unable to delete contact")
	}
	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil {
		return c.Fail(err, "invalid page number")
	}
	contact, err := c.Container.ORM.Contact.Get(ctx.Request().Context(), id)
	if err != nil {
		return c.Fail(err, "unable to delete contact")
	}
	if err := c.Container.ORM.Contact.DeleteOne(contact).Exec(ctx.Request().Context()); err != nil {
		return c.Fail(err, "unable to delete contact")
	}
	return routeToContactPage(c.Controller, ctx, page)
}

func routeToContactPage(c controller.Controller, ctx echo.Context, p int) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageContact
	page.Metatags.Description = "Contact page."
	page.Pager = controller.NewPager(ctx, 4)
	page.Pager.Page = p
	page.Data = fetchContacts(c, ctx, &page.Pager)

	return c.RenderPage(ctx, page)
}

func getContacts(c controller.Controller, ctx echo.Context, pager *controller.Pager) (int, []*ent.Contact) {
	contacts, err := c.Container.ORM.Contact.Query().
		Order(ent.Desc("created_at")).
		Limit(pager.ItemsPerPage).
		Offset(pager.GetOffset()).
		All(ctx.Request().Context())
	if err != nil {
		return 0, nil
	}
	total, err := c.Container.ORM.Contact.Query().
		Count(ctx.Request().Context())
	if err != nil {
		return 0, nil
	}
	return total, contacts
}

func (c *contact) Post(ctx echo.Context) error {
	var form contactForm
	ctx.Set(context.FormKey, &form)

	if err := ctx.Bind(&form); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if form.Submission.HasErrors() {
		return ctx.JSON(http.StatusBadRequest, form.Submission)
	}
	contact, err := c.Container.ORM.Contact.Create().
		SetEmail(form.Email).
		SetLink(form.Link).
		SetType(form.Type).
		SetMessage(form.Message).
		Save(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusCreated, contact)
}
