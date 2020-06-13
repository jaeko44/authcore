package languages

import (
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

// CheckAvailableLanguages returns the availablility for the given locale string
func CheckAvailableLanguages(lang string) bool {
	availableLanguages := viper.GetStringSlice("available_languages")
	for _, availableLang := range availableLanguages {
		if availableLang == lang {
			return true
		}
	}
	return false
}

// LanguageTagFromAvailableLanguages return Tag from available languages given locale string
func LanguageTagFromAvailableLanguages(lang string) language.Tag {
	availableLanguages := viper.GetStringSlice("available_languages")
	for _, availableLang := range availableLanguages {
		if availableLang == lang {
			return language.Make(lang)
		}
	}
	return language.Make(viper.GetString("default_language"))
}
