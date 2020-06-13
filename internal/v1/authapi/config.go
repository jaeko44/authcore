package authapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"

	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/nulls"
)

// GetWidgetsSettings returns settings for widgets
func (s Service) GetWidgetsSettings(ctx context.Context, in *authapi.GetWidgetsSettingsRequest) (*authapi.GetWidgetsSettingsResponse, error) {
	clientID := in.ClientId
	setting := map[string]interface{}{
		"analytics_token": viper.GetString("analytics_token"),
	}

	// return global settings only if default clientID is empty
	clientApp, err := clientapp.GetByClientID(clientID)
	if clientApp == nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	// TODO: reset_password_redirect_link can be changed in generic
	redirectFallbackURL := viper.GetString("reset_password_redirect_link")
	// Only replace the link with origin when it contains %s at the beginning.
	// The path redirected shall be Authcore management web sign in page by default,
	// which is hosted in the same origin. Frontend could use relative path
	// for redirection.

	// If client set to redirect back to its own application, the path should be in absolute
	// format. There will be no replacement in that case.
	if strings.HasPrefix(redirectFallbackURL, "%s") {
		redirectFallbackURL = fmt.Sprintf(redirectFallbackURL, "")
	}

	applicationSettings := map[string]interface{}{
		"matters_unlink_disabled": viper.GetBool("matters_unlink_disabled"),
		"app_hosts":               clientApp.AppDomains,
		"redirect_fallback_url":   redirectFallbackURL,
	}

	// merge two config without overwrite
	for k, v := range applicationSettings {
		if _, ok := setting[k]; !ok {
			setting[k] = v
		}
	}

	settingString, err := nulls.NewJSON(setting).String()

	if err != nil {
		return nil, errors.New(errors.ErrorUnknown, "")
	}

	return &authapi.GetWidgetsSettingsResponse{WidgetsSettings: settingString}, nil
}
