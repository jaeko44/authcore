// Reference: http://kazupon.github.io/vue-i18n/guide/lazy-loading.html

import Vue from 'vue'
import VueI18n from 'vue-i18n'
import messages from '@/lang/en'

Vue.use(VueI18n)

const allowedLanguage = [
  'en',
  'zh-HK'
]

export const i18n = new VueI18n({
  locale: 'en',
  fallbackLocale: 'en',
  messages
})

const loadedLanguages = ['en']

function setI18nLanguage (lang) {
  i18n.locale = lang
  document.querySelector('html').setAttribute('lang', lang)
  return lang
}

export function loadLanguageAsync (lang) {
  if (!allowedLanguage.includes(lang)) {
    // Warn the client. Language is default to be English if not loaded
    console.warn('language parameter is invalid, the required language is not loaded')
    return
  }
  // If the same language
  if (i18n.locale === lang) {
    return Promise.resolve(setI18nLanguage(lang))
  }

  // If the language was already loaded
  if (loadedLanguages.includes(lang)) {
    return Promise.resolve(setI18nLanguage(lang))
  }

  // If the language hasn't been loaded yet
  return import(/* webpackChunkName: "lang-[request]" */ `@/lang/${lang}.js`).then(
    messages => {
      i18n.setLocaleMessage(lang, messages.default[lang])
      loadedLanguages.push(lang)
      return setI18nLanguage(lang)
    }
  )
}
