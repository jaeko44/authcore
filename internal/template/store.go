package template

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"regexp"

	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

// Store manages templates for email and sms .
type Store struct {
	db *db.DB
}

// NewStore retrusn a new Store instance.
func NewStore(d *db.DB) *Store {
	return &Store{
		db: d,
	}
}

// FindAllTemplates lists the templates.
func (s *Store) FindAllTemplates(ctx context.Context) (*[]Template, error) {
	templates := &[]Template{}
	err := sqlx.SelectContext(ctx, s.db, templates, "SELECT * FROM templates")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return templates, nil
}

// FindTemplateByTypeAndLocale lookups a template record by type and locale.
func (s *Store) FindTemplateByTypeAndLocale(ctx context.Context, name string, langTag language.Tag) (*Template, error) {
	template := &Template{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM templates WHERE name = ? AND language = ?", name, langTag.String()).StructScan(template)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return template, nil
}

// UpdateOrCreateTemplate updates a specified template, or create one if it does not exist.
func (s *Store) UpdateOrCreateTemplate(ctx context.Context, name string, languageTag language.Tag, sTemplate string) error {
	template := &Template{
		Name:     name,
		Language: languageTag.String(),
		Template: sTemplate,
	}
	_, err := sqlx.NamedExecContext(
		ctx,
		s.db,
		"INSERT INTO templates (name, language, template) VALUES (:name, :language, :template) ON DUPLICATE KEY UPDATE template = :template",
		&template,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTemplateByNameAndLanguage deletes a template by name and language tag
func (s *Store) DeleteTemplateByNameAndLanguage(ctx context.Context, name string, languageTag language.Tag) error {
	_, err := s.FindTemplateByTypeAndLocale(ctx, name, languageTag)
	if err != nil {
		return errors.Wrap(err, errors.ErrorNotFound, "")
	}

	_, err = s.db.QueryxContext(ctx, "DELETE FROM templates WHERE language = ? AND name = ?", languageTag.String(), name)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// GetTemplate query the DB if it exists, or fallback to default template file.
func (s *Store) GetTemplate(ctx context.Context, name string, langTag language.Tag) (String, error) {
	re := regexp.MustCompile(`^\w+\.\w+$`)
	if !re.Match([]byte(name)) {
		return String{}, errors.New("illegal name", "")
	}

	template, err := s.FindTemplateByTypeAndLocale(ctx, name, langTag)
	if err == nil {
		return String{
			Template:  template.Template,
			UpdatedAt: template.UpdatedAt,
		}, nil
	}
	if !errors.IsKind(err, errors.ErrorNotFound) {
		return String{}, err
	}
	basePath := viper.GetString("base_path")
	filePath := fmt.Sprintf("%s/templates/%v/%v", basePath, langTag, name)
	var bytes []byte
	bytes, err = ioutil.ReadFile(filePath)
	if err != nil {
		return String{}, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return ToTemplateString(string(bytes)), nil
}
