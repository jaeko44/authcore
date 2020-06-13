package config

import (
	"encoding/json"
	"sync"

	"authcore.io/authcore/pkg/secret"

	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var initOnce sync.Once

// Config is an alias to *viper.Viper
type Config *viper.Viper

// SecretConfigKey contains a list of secret config keys that are masked.
var SecretConfigKey = []string{
	"database_url",
	"secret_key_base",
	"secret_key_base_old",
	"facebook_app_secret",
	"google_app_secret",
	"apple_app_private_key",
	"matters_app_secret",
	"twitter_consumer_secret",
	"sendgrid_api_key",
	"twilio_account_sid",
	"twilio_service_sid",
	"twilio_auth_token",
	"aws_ses_secret_access_key",
	"external_webhook_token",

	// auth
	"access_token_private_key",

	// secretdgateway
	"secrets.secretd_client_private_key",
}

// InitConfig set the defaults and read config from config files.
func InitConfig() {
	initEnv("authcore")
	readInConfig()
	InitSecretConfig(SecretConfigKey)
}

// InitSecretConfig convert secret config to secret.String
func InitSecretConfig(keys []string) {
	for _, key := range keys {
		val := viper.Get(key)
		_, ok := val.(secret.String)
		if !ok {
			viper.Set(key, secret.NewString(viper.GetString(key)))
		}
	}
}

// PrintConfig prints all settings to stdout.
func PrintConfig() {
	allSettings, err := json.MarshalIndent(viper.AllSettings(), "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("resolved settings: %s", allSettings)
}

// InitDefaults initializes default configs.
func InitDefaults() {
	// template
	basePath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot get executable directory: %v", err)
	}
	viper.SetDefault("base_path", basePath)

	// config file
	viper.SetDefault("config", "authcore.toml")
	viper.SetDefault("include_configs", []string{})

	// http
	viper.SetDefault("base_url", "https://authcore.localhost/")
	viper.SetDefault("static_cache_ttl", 2592000) // Default is 30 days. Set to 0 to disable cache.
	viper.SetDefault("grpc_listen", "0.0.0.0:7000")
	viper.SetDefault("http_listen", "0.0.0.0:80")
	viper.SetDefault("https_listen", "0.0.0.0:443")
	viper.SetDefault("https_enabled", false)
	viper.SetDefault("docs_enabled", true)
	viper.SetDefault("apiv1_enabled", false)

	// server
	viper.SetDefault("redis_address", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("redis_sentinel_enabled", false)
	viper.SetDefault("redis_sentinel_addresses", []string{})
	viper.SetDefault("redis_sentinel_master_name", "mymaster")
	viper.SetDefault("migration_dir", "./db/migrations")

	// language
	viper.SetDefault("available_languages", []string{
		"en",
		"zh-HK",
	})

	// authn
	viper.SetDefault("sign_up_enabled", true)
	viper.SetDefault("spake2_time_limit", "10m")
	viper.SetDefault("contact_rate_limit_interval", "1m")
	viper.SetDefault("contact_rate_limit_count", "1")
	viper.SetDefault("reset_link_rate_limit_interval", "1m")
	viper.SetDefault("reset_link_rate_limit_count", "1")
	viper.SetDefault("second_factor_rate_limit_interval", "10m")
	viper.SetDefault("second_factor_rate_limit_count", "10")
	viper.SetDefault("authentication_rate_limit_interval", "3h")
	viper.SetDefault("authentication_rate_limit_count", "10")
	viper.SetDefault("pow_challenge_time_limit", "10m")
	viper.SetDefault("authentication_time_limit", "15m")
	viper.SetDefault("reset_password_count_limit", 5)
	viper.SetDefault("authenticate_reset_password_time_limit", "504h") // 3 weeks.
	viper.SetDefault("authorization_token_expires_in", "10m")
	viper.SetDefault("pow_challenge_difficulty", "65536")
	viper.SetDefault("access_token_expires_in", "8h")
	viper.SetDefault("access_token_private_key", "")
	viper.SetDefault("session_expires_in", "720h") // 30 days.
	viper.SetDefault("default_client_id", "")
	viper.SetDefault("sms_code_length", "6")
	viper.SetDefault("sms_code_expiry", "5m")
	viper.SetDefault("reset_link_expiry", "5m")
	viper.SetDefault("reset_password_redirect_link", "%s/web/sign-in")
	viper.SetDefault("default_idp_list", []string{})

	viper.SetDefault("create_oauth_factor_state_expires_in", "10m")

	viper.SetDefault("matters_unlink_disabled", false)
	viper.SetDefault("analytics_token", "")

	// email & sms
	viper.SetDefault("default_language", viper.GetStringSlice("available_languages")[0])
	viper.SetDefault("application_name", "Authcore")

	// email
	viper.SetDefault("application_logo", "")
	viper.SetDefault("reset_password_authentication_email_sender_name", "Authcore")
	viper.SetDefault("reset_password_authentication_email_sender_address", "noreply@authcore.io")
	viper.SetDefault("verification_email_sender_name", "Authcore")
	viper.SetDefault("verification_email_sender_address", "noreply@authcore.io")

	// identity
	viper.SetDefault("require_user_email_or_phone", true)
	viper.SetDefault("require_user_phone", false)
	viper.SetDefault("require_user_email", false)
	viper.SetDefault("require_user_username", false)

	// email
	viper.SetDefault("reset_password_authentication_email_sender", "noreply@authcore.io")
	viper.SetDefault("verification_email_sender", "noreply@authcore.io")

	// Closed loop configurations
	viper.SetDefault("closed_loop_max_attempts", "5")
	viper.SetDefault("closed_loop_code_length", "6")
	viper.SetDefault("closed_loop_verification_request_duration", "1m")
	viper.SetDefault("closed_loop_authentication_request_duration", "1m")
	viper.SetDefault("contact_verification_expiry_for_email", "24h")
	viper.SetDefault("contact_verification_expiry_for_phone", "10m")
	viper.SetDefault("contact_authentication_expiry_for_phone", "10m")
	viper.SetDefault("contact_reset_password_authentication_expiry", "10m")

	// session
	viper.SetDefault("session_expires_in", "720h") // 30 days.
	viper.SetDefault("access_token_expires_in", "8h")

	// integration
	viper.SetDefault("matters_url", "https://server.matters.news")

	// secretdgateway
	viper.SetDefault("secretdgateway_enabled", false)
	viper.SetDefault("secretd_address", "127.0.0.1:9000")
}

// Reset resets the config to initial state.
func Reset() {
	viper.Reset()
}

func initEnv(envPrefix string) {
	viper.BindEnv("database_url", "DATABASE_URL")
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetDefault("config", "configs/authcore.yaml")
}

func readInConfig() {
	viper.SetConfigType("yaml")

	mainConfigFile := viper.GetString("config") //default config file
	log.Printf("read in config file: %s", mainConfigFile)
	viper.SetConfigName(mainConfigFile)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		_, configNotFound := err.(viper.ConfigFileNotFoundError)
		if configNotFound {
			log.Printf("config file not found: %v", err)
		} else {
			log.Fatalf("failed to read config file: %v", err)
		}
	}

	// read in multiple config file
	configFiles := viper.GetStringSlice("include_configs")

	for _, config := range configFiles {
		log.Printf("read in config file: %s", config)
		viper.SetConfigName(config)
		viper.AddConfigPath(viper.GetString("base_path"))
		// first config file
		if err := viper.MergeInConfig(); err != nil {
			_, configNotFound := err.(viper.ConfigFileNotFoundError)
			if configNotFound {
				log.Printf("config file not found: %v", err)
			} else {
				log.Fatalf("failed to read config file: %v", err)
			}
		}
	}
}
