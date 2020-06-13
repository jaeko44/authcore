package settings

import (
	"fmt"
	"net/http"
	"strings"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// APIv2 returns a function that registers settings related API 2.0 endpoints with an Echo instance.
func APIv2() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{}

		g := e.Group("/api/v2")
		g.GET("/preferences", h.Preferences)
	}
}

type handler struct {
}

func (h *handler) Preferences(c echo.Context) error {
	clientID := c.QueryParam("clientId")

	clientApp, err := clientapp.GetByClientID(clientID)
	if clientApp == nil {
		return errors.New(errors.ErrorInvalidArgument, "no client app is associated with client id")
	}
	if err != nil {
		return err
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

	settings := JSONSettings{
		// Settings
		AnalyticsToken: viper.GetString("analytics_token"),
		// Application settings
		AppHosts:              clientApp.AppDomains,
		MattersUnlinkDisabled: viper.GetBool("matters_unlink_disabled"),
		SignUpEnabled:         viper.GetBool("sign_up_enabled"),
		Preferences: JSONPreferences{
			Company: clientApp.Name,
			Logo:    clientApp.Logo,
			IDPList: clientApp.IDPList,
		},
		RedirectFallbackURL: redirectFallbackURL,
	}

	return c.JSON(http.StatusOK, settings)
}

// JSONSettings represents settings in API
type JSONSettings struct {
	AnalyticsToken        string          `json:"analytics_token"`
	AppHosts              []string        `json:"app_hosts"`
	MattersUnlinkDisabled bool            `json:"matters_unlink_disabled"`
	SignUpEnabled         bool            `json:"sign_up_enabled"`
	Preferences           JSONPreferences `json:"preferences"`
	RedirectFallbackURL   string          `json:"redirect_fallback_url"`
}

// JSONPreferences represents preferences for the application in API
type JSONPreferences struct {
	Company string   `json:"company"`
	Logo    string   `json:"logo"`
	IDPList []string `json:"idp_list"`
}
