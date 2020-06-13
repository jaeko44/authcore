package languages

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestCheckAvailableLanguages(t *testing.T) {
	viper.SetDefault("available_languages", []string{
		"en",
		"zh-HK",
	})

	assert.True(t, CheckAvailableLanguages("en"))
	assert.True(t, CheckAvailableLanguages("zh-HK"))
	assert.False(t, CheckAvailableLanguages(""))
}

func TestLanguageTagFromAvailableLanguages(t *testing.T) {
	viper.SetDefault("available_languages", []string{
		"en",
		"zh-HK",
	})

	viper.SetDefault("default_language", viper.GetStringSlice("available_languages")[0])

	assert.Equal(t, LanguageTagFromAvailableLanguages("en"), language.Make("en"))
	assert.Equal(t, LanguageTagFromAvailableLanguages("zh-HK"), language.Make("zh-HK"))
	assert.Equal(t, LanguageTagFromAvailableLanguages(""), language.Make("en"))
}
