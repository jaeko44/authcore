package template

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"authcore.io/authcore/internal/apiutil"
	"authcore.io/authcore/internal/errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/language"
)

// APIv2 returns a function that registers API 2.0 endpoints with an Echo instance.
func APIv2(store *Store) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{store: store}

		g := e.Group("/api/v2")
		g.GET("/templates", h.ListAvaliableLanguage)
		g.GET("/templates/:type/:lang", h.ListTemplates)
		g.GET("/templates/:type/:lang/:name", h.GetTemplate)
		g.POST("/templates/:type/:lang/:name", h.UpdateTemplate)
		g.DELETE("/templates/:type/:lang/:name", h.ResetTemplate)
	}
}

type handler struct {
	store *Store
}

func (h *handler) ListAvaliableLanguage(c echo.Context) error {
	languages := GetAvailableLanguages()
	resp := apiutil.NewListPagination(languages, nil)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) ListTemplates(c echo.Context) error {
	t := c.Param("type")
	ty, ok := parseTemplateTypeFromString(t)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	lang := c.Param("lang")
	if !contains(GetAvailableLanguages(), lang) {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	templateNames := GetTemplateNamesFromType(ty)
	ctx := c.Request().Context()

	templates, err := h.store.FindAllTemplates(ctx)
	if err != nil {
		return err
	}
	var jsonTemplates []JSONTemplate
	for _, templateName := range templateNames {
		jsonTemplates = append(jsonTemplates, JSONTemplate{
			Name:      templateName,
			Language:  lang,
			UpdatedAt: getUpdatedAt(templates, lang, templateName),
		})
	}

	resp := apiutil.NewListPagination(jsonTemplates, nil)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) GetTemplate(c echo.Context) error {
	t := c.Param("type")
	ty, ok := parseTemplateTypeFromString(t)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	lang := c.Param("lang")
	languageTag, err := language.Parse(lang)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	name := c.Param("name")

	ctx := c.Request().Context()
	r := &JSONTemplateString{}
	switch ty {
	case TemplateEMAIL:
		if !contains(EmailTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		subjectName := fmt.Sprintf("%s.subj", name)
		subject, err := h.store.GetTemplate(ctx, subjectName, languageTag)
		if err != nil {
			return err
		}
		htmlName := fmt.Sprintf("%s.html", name)
		html, err := h.store.GetTemplate(ctx, htmlName, languageTag)
		if err != nil {
			return err
		}
		textName := fmt.Sprintf("%s.txt", name)
		text, err := h.store.GetTemplate(ctx, textName, languageTag)
		if err != nil {
			return err
		}
		r.Subject = subject.Template
		r.HTML = html.Template
		r.Text = text.Template
	case TemplateSMS:
		if !contains(SMSTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		textName := fmt.Sprintf("%s.txt", name)
		text, err := h.store.GetTemplate(ctx, textName, languageTag)
		if err != nil {
			return err
		}
		r.Text = text.Template
	}
	return c.JSON(http.StatusOK, r)
}

func (h *handler) UpdateTemplate(c echo.Context) error {
	t := c.Param("type")
	ty, ok := parseTemplateTypeFromString(t)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	lang := c.Param("lang")
	languageTag, err := language.Parse(lang)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	name := c.Param("name")
	r := JSONTemplateString{}
	if err := c.Bind(&r); err != nil {
		return err
	}

	ctx := c.Request().Context()
	switch ty {
	case TemplateEMAIL:
		if !contains(EmailTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		subjectName := fmt.Sprintf("%s.subj", name)
		err := h.store.UpdateOrCreateTemplate(ctx, subjectName, languageTag, r.Subject)
		if err != nil {
			return err
		}
		htmlName := fmt.Sprintf("%s.html", name)
		err = h.store.UpdateOrCreateTemplate(ctx, htmlName, languageTag, r.HTML)
		if err != nil {
			return err
		}
		textName := fmt.Sprintf("%s.txt", name)
		err = h.store.UpdateOrCreateTemplate(ctx, textName, languageTag, r.Text)
		if err != nil {
			return err
		}

	case TemplateSMS:
		if !contains(SMSTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		textName := fmt.Sprintf("%s.txt", name)
		err = h.store.UpdateOrCreateTemplate(ctx, textName, languageTag, r.Text)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *handler) ResetTemplate(c echo.Context) error {
	t := c.Param("type")
	ty, ok := parseTemplateTypeFromString(t)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	lang := c.Param("lang")
	languageTag, err := language.Parse(lang)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	name := c.Param("name")

	ctx := c.Request().Context()
	switch ty {
	case TemplateEMAIL:
		if !contains(EmailTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		subjectName := fmt.Sprintf("%s.subj", name)
		err := h.store.DeleteTemplateByNameAndLanguage(ctx, subjectName, languageTag)
		if err != nil {
			return err
		}
		htmlName := fmt.Sprintf("%s.html", name)
		err = h.store.DeleteTemplateByNameAndLanguage(ctx, htmlName, languageTag)
		if err != nil {
			return err
		}
		textName := fmt.Sprintf("%s.txt", name)
		err = h.store.DeleteTemplateByNameAndLanguage(ctx, textName, languageTag)
		if err != nil {
			return err
		}

	case TemplateSMS:
		if !contains(SMSTemplates, name) {
			return errors.New(errors.ErrorInvalidArgument, "")
		}
		textName := fmt.Sprintf("%s.txt", name)
		err = h.store.DeleteTemplateByNameAndLanguage(ctx, textName, languageTag)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func parseTemplateTypeFromString(in string) (Type, bool) {
	switch in {
	case "sms":
		return TemplateSMS, true
	case "email":
		return TemplateEMAIL, true
	}
	return TemplateSMS, false
}

func contains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func getUpdatedAt(templates *[]Template, language, templateNamePrefix string) *time.Time {
	updatedAt := time.Unix(0, 0)
	for _, template := range *templates {
		if language == template.Language && strings.HasPrefix(template.Name, templateNamePrefix) {
			updatedAt = template.UpdatedAt
		}
	}
	if updatedAt == time.Unix(0, 0) {
		return nil
	}
	return &updatedAt
}

// JSONTemplate represent a template in management API.
type JSONTemplate struct {
	Name      string     `json:"name"`
	Language  string     `json:"language"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// JSONTemplateString represent a template in management API.
type JSONTemplateString struct {
	Subject string `json:"subject,omitempty"`
	HTML    string `json:"html,omitempty"`
	Text    string `json:"text"`
}
