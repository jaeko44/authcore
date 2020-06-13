import assign from 'lodash-es/assign'

import {
  GET_CLIENT_COMPLETED,
  GET_CALLBACKS_COMPLETED,
  GET_CONTAINER_PARAMETER_COMPLETED,

  SET_LAYOUT_STATE,
  SET_OAUTH_WIDGET_STATE,
  SET_AUTHENTICATED_STATE
} from '@/store/types'

import { loadLanguageAsync } from '@/i18n-setup'

import { inMobile, camelCaseObject } from '@/utils/util.js'

const allowedButtonSize = [
  'normal',
  'large'
]

const allowedSocialLoginPaneOption = [
  'list',
  'grid'
]
const allowedSocialLoginPaneStyle = [
  'top',
  'bottom'
]

const defaultButtonSize = 'large'
const defaultSocialLoginPaneOption = 'grid'
const defaultSocialLoginPaneStyle = 'bottom'

export default {
  namespaced: true,
  modules: {},

  state: {
    authcoreClient: undefined,
    callbacks: {},
    logo: undefined,
    company: undefined,
    containerId: undefined,
    internal: undefined,
    primary: undefined,
    success: undefined,
    danger: undefined,
    privacyLink: undefined,
    privacyCheckbox: undefined,
    successRedirectUrl: '',
    authenticated: false,
    requireUsername: undefined,
    showAvatar: undefined,
    prefillContact: undefined,
    oAuthWidget: false,
    responseType: '',
    clientId: '',
    redirectUri: '',
    scope: '',
    oAuthState: '',
    codeChallenge: '',
    codeChallengeMethod: '',
    inMobile: undefined,
    buttonSize: undefined,
    socialLoginPaneOption: undefined,
    socialLoginPaneStyle: undefined,
    widgetsSettings: undefined,

    layoutStateSet: false,

    loading: false,
    done: false,
    error: undefined
  },
  getters: {
    accessToken: (state) => {
      if (state.authcoreClient !== undefined) {
        return state.authcoreClient.getAccessToken()
      }
      return undefined
    },
    isReady (state) {
      return state.widgetsSettings !== undefined
    }
  },

  mutations: {
    [GET_CLIENT_COMPLETED] (state, client) {
      state.authcoreClient = client
      // Set the state in mobile or desktop case
      // Using the recommendation from https://developer.mozilla.org/en-US/docs/Web/HTTP/Browser_detection_using_the_user_agent for mobile checking
      state.inMobile = inMobile()

      // Add event listener for postMessage
      window.addEventListener('message', async e => {
        // Upon receiving a message of type 'AuthCore_*', a callback function will be called.
        // For example, if AuthCore_getCurrentUser is received, `getCurrentUser` will be called.
        if (typeof e.data !== 'object') return
        const { type, data } = e.data
        if (typeof type !== 'string' || !(type.startsWith('AuthCore_'))) return
        const callbackName = type.substr(9)
        if (typeof state.callbacks[callbackName] !== 'function') return
        await state.callbacks[callbackName](data)
      })
    },
    [GET_CALLBACKS_COMPLETED] (state, callbacks) {
      state.callbacks = assign(state.callbacks, callbacks)
    },
    [GET_CONTAINER_PARAMETER_COMPLETED] (state, payload) {
      const {
        clientId,
        containerId,
        internal,
        primary,
        success,
        danger,
        privacyLink,
        privacyCheckbox,
        successRedirectUrl,
        verification,
        requireUsername,
        showAvatar,
        fixedContact,
        prefillContact
      } = payload
      state.clientId = clientId
      state.containerId = containerId
      state.privacyLink = privacyLink
      state.prefillContact = prefillContact
      state.fixedContact = fixedContact === 'true'
      state.internal = internal === 'true'
      state.verification = verification === 'true'
      state.privacyCheckbox = privacyCheckbox === 'true'
      state.requireUsername = requireUsername === 'true'
      state.successRedirectUrl = successRedirectUrl
      state.showAvatar = showAvatar === 'true'
      if (primary !== undefined) {
        state.primary = primary
      }
      if (success !== undefined) {
        state.success = success
      }
      if (danger !== undefined) {
        state.danger = danger
      }
    },
    [SET_LAYOUT_STATE] (state, payload) {
      const {
        logo,
        company,
        buttonSize,
        socialLoginPaneOption,
        socialLoginPaneStyle,
        language
      } = payload
      if (logo === 'undefined') {
        state.logo = undefined
      } else {
        state.logo = logo
      }
      if (company === 'undefined') {
        state.company = undefined
      } else {
        state.company = company
      }

      if (!allowedButtonSize.includes(buttonSize)) {
        // Fallback to default one if the value is invalid
        state.buttonSize = defaultButtonSize
        if (buttonSize !== '') {
          console.warn('buttonSize parameter is invalid')
        }
      } else {
        state.buttonSize = buttonSize
      }
      if (!allowedSocialLoginPaneOption.includes(socialLoginPaneOption)) {
        // Fallback to default one if the value is invalid
        state.socialLoginPaneOption = defaultSocialLoginPaneOption
        if (socialLoginPaneOption !== '') {
          console.warn('socialLoginPaneOption parameter is invalid')
        }
      } else {
        state.socialLoginPaneOption = socialLoginPaneOption
      }
      if (!allowedSocialLoginPaneStyle.includes(socialLoginPaneStyle)) {
        // Fallback to default one if the value is invalid
        state.socialLoginPaneStyle = defaultSocialLoginPaneStyle
        if (socialLoginPaneStyle !== '') {
          console.warn('socialLoginPaneStyle parameter is invalid')
        }
      } else {
        state.socialLoginPaneStyle = socialLoginPaneStyle
      }

      if (language !== undefined) {
        loadLanguageAsync(language)
      }
      state.layoutStateSet = true
    },
    [SET_OAUTH_WIDGET_STATE] (state, payload) {
      const { responseType, clientId, redirectUri, scope, state: oAuthState, codeChallenge, codeChallengeMethod } = payload
      state.oAuthWidget = true
      state.responseType = responseType
      state.clientId = clientId
      state.redirectUri = redirectUri
      state.scope = scope
      state.oAuthState = oAuthState
      state.codeChallenge = codeChallenge
      state.codeChallengeMethod = codeChallengeMethod
    },
    [SET_AUTHENTICATED_STATE] (state, authenticated) {
      state.authenticated = authenticated
    },
    SET_WIDGETS_SETTINGS (state, payload) {
      if (payload.preferences) {
        payload.preferences = camelCaseObject(payload.preferences)
      }
      state.widgetsSettings = camelCaseObject(payload)
    },
    GET_WIDGETS_SETTINGS_FAILED (state, err) {
      console.error('Invalid Client ID / widgets setting')
      console.error(err)
      state.error = err
    }
  }
}
