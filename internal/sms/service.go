package sms

import (
	"context"
	"fmt"
	"net/url"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/template"
)

// Service provides services for delivering transactional SMS
type Service struct {
	templateStore *template.Store
}

// NewService initialize a new Service.
func NewService(templateStore *template.Store) *Service {
	return &Service{
		templateStore: templateStore,
	}
}

// SendVerificationSMS sends a verification SMS
func (s *Service) SendVerificationSMS(ctx context.Context, displayName, phone, code string) error {
	langTag, err := defaultLanguage()
	if err != nil {
		return err
	}
	messageTemplate, err := s.templateStore.GetTemplate(ctx, "VerificationSMS.txt", langTag)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	m := map[string]string{
		"code":         code,
		"display_name": displayName,
	}

	return sendSMS(messageTemplate, m, phone)
}

// SendAuthenticationSMS sends an authentication SMS
func (s *Service) SendAuthenticationSMS(ctx context.Context, displayName, phone, code string) error {
	langTag, err := defaultLanguage()
	if err != nil {
		return err
	}
	messageTemplate, err := s.templateStore.GetTemplate(ctx, "AuthenticationSMS.txt", langTag)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	m := map[string]string{
		"code":         code,
		"display_name": displayName,
	}

	return sendSMS(messageTemplate, m, phone)
}

// SendResetPasswordAuthenticationSMS sends an authentication SMS for reset password
func (s *Service) SendResetPasswordAuthenticationSMS(ctx context.Context, hostname, displayName, phone, token, resetPasswordRedirectLink string) error {
	langTag, err := defaultLanguage()
	if err != nil {
		return err
	}
	messageTemplate, err := s.templateStore.GetTemplate(ctx, "ResetPasswordAuthenticationSMS.txt", langTag)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	m := map[string]string{
		"reset_password_link": fmt.Sprintf("%s/widgets/reset-password/contact/%s?redirect_uri=%s&identifier=%s", hostname, token, url.QueryEscape(resetPasswordRedirectLink), url.QueryEscape(phone)),
		"display_name":        displayName,
	}

	return sendSMS(messageTemplate, m, phone)
}

// SendResetLinkV2 sends an reset password link SMS.
func (s *Service) SendResetLinkV2(ctx context.Context, resetLink, phone string) error {
	langTag, err := defaultLanguage()
	if err != nil {
		return err
	}
	messageTemplate, err := s.templateStore.GetTemplate(ctx, "ResetPasswordAuthenticationSMS.txt", langTag)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	m := map[string]string{
		"reset_password_link": resetLink,
		"display_name":        "",
	}

	return sendSMS(messageTemplate, m, phone)
}