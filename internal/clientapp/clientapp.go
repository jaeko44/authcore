package clientapp

import (
	"net/url"
	"regexp"
	"strings"
	"sync"

	"authcore.io/authcore/internal/errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AdminPortalClientID is the client ID of the Admin Portal app.
const AdminPortalClientID string = "_authcore_admin_portal_"

var loadConfigOnce sync.Once
var clientAppsMap map[string]ClientApp

// ClientApp contains application specific attributes
type ClientApp struct {
	ID                  string
	Name                string
	Logo                string
	AppDomains          []string `mapstructure:"app_domains"`
	AllowedCallbackURLs []string `mapstructure:"allowed_callback_urls"`
	IDPList             []string `mapstructure:"idp_list"`
}

// GetByClientID retrieve ClientApp from viper configs
func GetByClientID(clientID string) (*ClientApp, error) {
	// Treat authcore.io as empty default client ID to fallback for old react-native SDK change.
	// See: https://gitlab.com/blocksq/authcore/issues/821 for details
	if clientID == "authcore.io" {
		clientID = ""
	}
	if !validateClientIDFormat(clientID) {
		return nil, errors.New(errors.ErrorUnknown, "invalid client id")
	}
	// default client ID fallback
	if clientID == "" {
		clientID = viper.GetString("default_client_id")
	}

	loadConfigOnce.Do(func() {
		var err error
		clientAppsMap, err = LoadClientApps()
		if err != nil {
			log.Errorf("error loading applications config: %v", err)
			clientAppsMap = make(map[string]ClientApp)
		}
	})

	clientApp, ok := clientAppsMap[strings.ToLower(clientID)]
	if !ok {
		return nil, errors.Errorf(errors.ErrorUnknown, "invalid client id %v", clientID)
	}
	return &clientApp, nil
}

// GetAdminPortalClientApp returns a ClientApp instance that represents the built-in Authcore portal
// app.
func GetAdminPortalClientApp() (ClientApp, error) {
	// set default config that depends on base_url after read in config
	baseURL, err := url.Parse(viper.GetString("base_url"))
	if err != nil {
		return ClientApp{}, err
	}
	logoURL, err := baseURL.Parse("/widgets/favicon.png")
	if err != nil {
		return ClientApp{}, err
	}
	webURL, err := baseURL.Parse("/web/")
	if err != nil {
		return ClientApp{}, err
	}

	return ClientApp{
		ID:                  AdminPortalClientID,
		Name:                "Authcore",
		Logo:                logoURL.String(),
		AppDomains:          []string{baseURL.Host},
		AllowedCallbackURLs: []string{webURL.String()},
	}, nil
}

// LoadClientApps loads the client apps from config.
func LoadClientApps() (map[string]ClientApp, error) {
	rawMap := make(map[string]ClientApp)
	err := viper.UnmarshalKey("applications", &rawMap)

	adminPortal, err := GetAdminPortalClientApp()
	if err != nil {
		return nil, errors.Errorf(errors.ErrorUnknown, "error reading applications config: %w", err)
	}
	rawMap[AdminPortalClientID] = adminPortal

	clientApps := make(map[string]ClientApp)
	for k, app := range rawMap {
		app.ID = strings.ToLower(k)
		if app.IDPList == nil {
			app.IDPList = viper.GetStringSlice("default_idp_list")
		}
		clientApps[app.ID] = app
	}
	return clientApps, nil
}

func validateClientIDFormat(clientID string) bool {
	// check if clientID contains only alphanumeric, underscore and hyphen
	return regexp.MustCompile("^[A-Za-z0-9_\\-]*$").MatchString(clientID)
}
