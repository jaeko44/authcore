package email

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"authcore.io/authcore/internal/languages"
	"authcore.io/authcore/internal/template"

	"github.com/spf13/viper"
)

// Service provides services related to template management.
type Service struct {
	templateStore *template.Store
}

// NewService initialize a new Service.
func NewService(templateStore *template.Store) *Service {
	return &Service{
		templateStore: templateStore,
	}
}

func (s *Service) getEmailTemplate(ctx context.Context, name string, lang string) (emailTemplate, error) {
	langTag := languages.LanguageTagFromAvailableLanguages(lang)
	list := []string{".subj", ".html", ".txt"}
	var content []template.String
	for _, ext := range list {
		fullName := fmt.Sprintf("%s%s", name, ext)
		template, err := s.templateStore.GetTemplate(ctx, fullName, langTag)
		if err != nil {
			return emailTemplate{}, err
		}
		content = append(content, template)
	}
	return emailTemplate{
		Subject:      content[0],
		HTMLTemplate: content[1],
		RawTemplate:  content[2],
	}, nil
}

// SendVerificationMail sends a verification mail
func (s *Service) SendVerificationMail(ctx context.Context, hostname, displayName, emailAddress, lang, code, token string) error {
	emailTemplate, err := s.getEmailTemplate(ctx, "VerificationMail", lang)
	if err != nil {
		return err
	}
	m := map[string]string{
		"code":         code,
		"display_name": displayName,
	}
	return sendMail(emailTemplate, m, displayName, emailAddress, viper.GetString("verification_email_sender_name"), viper.GetString("verification_email_sender_address"))
}

// SendResetPasswordAuthenticationMail sends an authentication mail for reset password
func (s *Service) SendResetPasswordAuthenticationMail(ctx context.Context, hostname, displayName, emailAddress, lang, token, resetPasswordRedirectLink string) error {
	emailTemplate, err := s.getEmailTemplate(ctx, "ResetPasswordAuthenticationMail", lang)
	if err != nil {
		return err
	}
	resetPasswordLinkURL, err := url.Parse(hostname)
	if err != nil {
		return err
	}
	resetPasswordLinkURL.Path = fmt.Sprintf("widgets/reset-password/contact/%s", token)
	query := url.Values{}
	query.Add("redirect_uri", url.QueryEscape(resetPasswordRedirectLink))
	query.Add("identifier", url.QueryEscape(emailAddress))
	query.Add("company", url.QueryEscape(viper.GetString("application_name")))
	query.Add("logo", url.QueryEscape(viper.GetString("application_logo")))
	resetPasswordLinkURL.RawQuery = query.Encode()
	m := map[string]string{
		"reset_password_link": resetPasswordLinkURL.String(),
		"display_name":        displayName,
	}
	return sendMail(emailTemplate, m, displayName, emailAddress, viper.GetString("reset_password_authentication_email_sender_name"), viper.GetString("reset_password_authentication_email_sender_address"))
}

// SendResetLinkV2 sends a reset link email.
func (s *Service) SendResetLinkV2(ctx context.Context, resetLink, emailAddress, lang string) error {
	emailTemplate, err := s.getEmailTemplate(ctx, "ResetPasswordAuthenticationMail", lang)
	if err != nil {
		return err
	}
	m := map[string]string{
		"reset_password_link": resetLink,
		"display_name":        "",
	}
	return sendMail(emailTemplate, m, "", emailAddress, viper.GetString("reset_password_authentication_email_sender_name"), viper.GetString("reset_password_authentication_email_sender_address"))
}

func sendMail(emailTemplate emailTemplate, emailContentMap map[string]string, displayName, emailAddress, senderName, senderAddress string) error {
	if strings.HasSuffix(os.Args[0], ".test") {
		return nil
	}
	if strings.Index(emailAddress, "example.com") > -1 {
		return nil
	}
	emailContentMap["application_name"] = viper.GetString("application_name")

	email := emailTemplate.Execute(emailContentMap)
	from := People{
		Name:  senderName,
		Email: senderAddress,
	}
	to := People{
		Name:  displayName,
		Email: emailAddress,
	}

	emailProvider, err := getProvider()
	if err != nil {
		return err
	}

	return emailProvider.Send(from, to, email.Subject, email.Raw, email.HTML)
}
