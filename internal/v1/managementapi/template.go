package managementapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"
	"authcore.io/authcore/pkg/slice"

	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/text/language"
)

// ListTemplates lists the templates.
func (s *Service) ListTemplates(ctx context.Context, in *managementapi.ListTemplatesRequest) (*managementapi.ListTemplatesResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListTemplatesPermission)
	if err != nil {
		return nil, err
	}

	templateType, ok := parseTemplateTypeFromString(in.Type)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	templates, err := s.TemplateStore.FindAllTemplates(ctx)
	if err != nil {
		return nil, err
	}

	pbLanguages := []*managementapi.TemplateLanguage{}
	for _, language := range template.GetAvailableLanguages() {
		pbLanguage := marshalLanguage(language)
		pbLanguages = append(pbLanguages, pbLanguage)
	}

	pbTemplates, err := listTemplates(templates, templateType)
	if err != nil {
		return nil, err
	}

	return &managementapi.ListTemplatesResponse{
		Templates: pbTemplates,
		Languages: pbLanguages,
	}, nil
}

// GetTemplate gets a template.
func (s *Service) GetTemplate(ctx context.Context, in *managementapi.GetTemplateRequest) (*managementapi.Template, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, GetTemplatePermission)
	if err != nil {
		return nil, err
	}

	languageTag, err := language.Parse(in.Language)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	templateType, ok := parseTemplateTypeFromString(in.Type)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	baseName := in.Name
	switch templateType {
	case template.TemplateEMAIL:
		if !slice.Contains(template.EmailTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		subjectName := fmt.Sprintf("%s.subj", baseName)
		subjectTemplate, err := s.TemplateStore.GetTemplate(ctx, subjectName, languageTag)
		if err != nil {
			return nil, err
		}
		htmlTemplateName := fmt.Sprintf("%s.html", baseName)
		htmlTemplate, err := s.TemplateStore.GetTemplate(ctx, htmlTemplateName, languageTag)
		if err != nil {
			return nil, err
		}
		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		textTemplate, err := s.TemplateStore.GetTemplate(ctx, textTemplateName, languageTag)
		if err != nil {
			return nil, err
		}

		updatedAt := template.GetUpdatedAtFromStrings(subjectTemplate, htmlTemplate, textTemplate)
		var pbUpdatedAt *timestamp.Timestamp
		if updatedAt != nil {
			pbUpdatedAt = &timestamp.Timestamp{
				Seconds: updatedAt.Unix(),
				Nanos:   int32(updatedAt.Nanosecond()),
			}
		}

		return &managementapi.Template{
			Language: languageTag.String(),
			Name:     baseName,
			Template: &managementapi.Template_EmailTemplate{
				EmailTemplate: &managementapi.EmailTemplate{
					Subject:      subjectTemplate.Template,
					HtmlTemplate: htmlTemplate.Template,
					TextTemplate: textTemplate.Template,
				},
			},
			UpdatedAt: pbUpdatedAt,
		}, nil
	case template.TemplateSMS:
		if !slice.Contains(template.SMSTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		textTemplate, err := s.TemplateStore.GetTemplate(ctx, textTemplateName, languageTag)
		if err != nil {
			return nil, err
		}

		updatedAt := template.GetUpdatedAtFromStrings(textTemplate)
		var pbUpdatedAt *timestamp.Timestamp
		if updatedAt != nil {
			pbUpdatedAt = &timestamp.Timestamp{
				Seconds: updatedAt.Unix(),
				Nanos:   int32(updatedAt.Nanosecond()),
			}
		}

		return &managementapi.Template{
			Language: languageTag.String(),
			Name:     baseName,
			Template: &managementapi.Template_SmsTemplate{
				SmsTemplate: &managementapi.SMSTemplate{
					Template: textTemplate.Template,
				},
			},
			UpdatedAt: pbUpdatedAt,
		}, nil

	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
}

// CreateTemplate creates a new template.
func (s *Service) CreateTemplate(ctx context.Context, in *managementapi.CreateTemplateRequest) (*managementapi.CreateTemplateResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, CreateTemplatePermission)
	if err != nil {
		return nil, err
	}

	inTemplate := in.Template

	baseName := inTemplate.Name
	languageTag, err := language.Parse(inTemplate.Language)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	switch templateContent := inTemplate.Template.(type) {
	case *managementapi.Template_EmailTemplate:
		emailTemplateContent := templateContent.EmailTemplate
		if !slice.Contains(template.EmailTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		subjectName := fmt.Sprintf("%s.subj", baseName)
		subjectContent := emailTemplateContent.Subject
		err = s.TemplateStore.UpdateOrCreateTemplate(ctx, subjectName, languageTag, subjectContent)
		if err != nil {
			return nil, err
		}

		htmlTemplateName := fmt.Sprintf("%s.html", baseName)
		htmlTemplateContent := emailTemplateContent.HtmlTemplate
		err = s.TemplateStore.UpdateOrCreateTemplate(ctx, htmlTemplateName, languageTag, htmlTemplateContent)
		if err != nil {
			return nil, err
		}

		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		textTemplateContent := emailTemplateContent.TextTemplate
		err = s.TemplateStore.UpdateOrCreateTemplate(ctx, textTemplateName, languageTag, textTemplateContent)
		if err != nil {
			return nil, err
		}
	case *managementapi.Template_SmsTemplate:
		smsTemplateContent := templateContent.SmsTemplate
		if !slice.Contains(template.SMSTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		textTemplateContent := smsTemplateContent.Template
		err = s.TemplateStore.UpdateOrCreateTemplate(ctx, textTemplateName, languageTag, textTemplateContent)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	return &managementapi.CreateTemplateResponse{}, nil
}

// ResetTemplate resets a template to its default setting.
func (s *Service) ResetTemplate(ctx context.Context, in *managementapi.ResetTemplateRequest) (*managementapi.ResetTemplateResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ResetTemplatePermission)
	if err != nil {
		return nil, err
	}

	languageTag, err := language.Parse(in.Language)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	templateType, ok := parseTemplateTypeFromString(in.Type)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	baseName := in.Name
	switch templateType {
	case template.TemplateEMAIL:
		if !slice.Contains(template.EmailTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		subjectName := fmt.Sprintf("%s.subj", baseName)
		err = s.TemplateStore.DeleteTemplateByNameAndLanguage(ctx, subjectName, languageTag)
		if err != nil {
			return nil, err
		}
		htmlTemplateName := fmt.Sprintf("%s.html", baseName)
		err = s.TemplateStore.DeleteTemplateByNameAndLanguage(ctx, htmlTemplateName, languageTag)
		if err != nil {
			return nil, err
		}
		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		err = s.TemplateStore.DeleteTemplateByNameAndLanguage(ctx, textTemplateName, languageTag)
		if err != nil {
			return nil, err
		}
	case template.TemplateSMS:
		if !slice.Contains(template.SMSTemplates, baseName) {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}

		textTemplateName := fmt.Sprintf("%s.txt", baseName)
		err = s.TemplateStore.DeleteTemplateByNameAndLanguage(ctx, textTemplateName, languageTag)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	return &managementapi.ResetTemplateResponse{}, nil
}

func parseTemplateTypeFromString(in string) (template.Type, bool) {
	switch in {
	case "sms":
		return template.TemplateSMS, true
	case "email":
		return template.TemplateEMAIL, true
	}
	return template.TemplateSMS, false
}

func listTemplates(templates *[]template.Template, templateType template.Type) ([]*managementapi.Template, error) {
	var pbTemplates []*managementapi.Template

	var availableTemplates []string
	switch templateType {
	case template.TemplateEMAIL:
		availableTemplates = template.EmailTemplates
	case template.TemplateSMS:
		availableTemplates = template.SMSTemplates
	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	for _, language := range template.GetAvailableLanguages() {
		for _, templateName := range availableTemplates {
			pbTemplates = append(pbTemplates, &managementapi.Template{
				Language:  language,
				Name:      templateName,
				UpdatedAt: getUpdatedAt(templates, language, templateName),
			})
		}
	}
	return pbTemplates, nil
}

func getUpdatedAt(templates *[]template.Template, language, templateNamePrefix string) *timestamp.Timestamp {
	updatedAt := time.Unix(0, 0)
	for _, template := range *templates {
		if language == template.Language && strings.HasPrefix(template.Name, templateNamePrefix) {
			updatedAt = template.UpdatedAt
		}
	}
	if updatedAt == time.Unix(0, 0) {
		return nil
	}
	return &timestamp.Timestamp{
		Seconds: updatedAt.Unix(),
		Nanos:   int32(updatedAt.Nanosecond()),
	}
}

func marshalLanguage(language string) *managementapi.TemplateLanguage {
	return &managementapi.TemplateLanguage{
		Language: language,
	}
}
