import { isPossibleNumber } from 'libphonenumber-js'

import client from '@/client'
import { i18n } from '@/i18n-setup'

import {
  INPUT_HANDLE_NOT_FOUND_ERROR,
  INPUT_HANDLE_ALREADY_EXISTS_ERROR
} from '@/store/types'

export default {
  namespaced: true,

  state: {
    authnState: null,
    loading: false,
    error: null,
    handle: '',
    handleType: '',
    password: '',
    passwordConfirmation: '',
    onetimeCode: '',
    selectedMFA: '',

    // For sign up and password reset
    passwordScore: -1,

    // For signUp
    privacyChecked: false,
    signUpErrors: null,

    // For social login
    recoveryEmail: '',
    recoveryEmailError: null,

    // For password reset
    resetToken: ''
  },

  getters: {
    isAuthenticated () {
      return !!client.tokenManager.get('access_token')
    }
  },

  actions: {
    async start ({ commit, state }, { handle, redirectURI, codeChallenge, codeChallengeMethod, clientState }) {
      try {
        if (handle) {
          commit('SET_HANDLE', handle)
        }
        commit('SET_LOADING')
        const authnState = await client.authn.start(state.handle, redirectURI, { codeChallenge, codeChallengeMethod, clientState })
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        var error = err
        if (err.response) {
          if (err.response.status === 403) {
            error = i18n.t('sign_in.input.error.user_is_locked')
          } else if (err.response.status === 404) {
            error = INPUT_HANDLE_NOT_FOUND_ERROR
          }
        }
        commit('SET_ERROR', error)
      }
    },

    async verifyPassword ({ commit, state }, password) {
      try {
        if (password) {
          commit('SET_PASSWORD', password)
        }
        commit('SET_LOADING')
        const authnState = await client.authn.verifyPassword(state.authnState, state.password)
        switch (authnState.status) {
          case 'PRIMARY':
            commit('SET_ERROR', i18n.t('sign_in.input.error.incorrect_password'))
            break
          case 'SUCCESS':
          case 'MFA_REQUIRED':
            commit('SET_AUTHN_STATE', authnState)
            break
          default:
            commit('SET_ERROR', new Error('unexpected status ' + authnState.status))
        }
      } catch (err) {
        var error = err
        if (err.response) {
          if (err.response.status === 403) {
            error = i18n.t('sign_in.input.error.incorrect_password')
          }
        }
        commit('SET_ERROR', error)
      }
    },

    async requestMFA ({ commit, state }) {
      try {
        if (this.selectedMFA === 'sms_otp') {
          commit('SET_LOADING')
          return client.authn.requestSMSOTP(state.authnState)
        }
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async verifyMFA ({ commit, state }, code) {
      try {
        commit('SET_LOADING')
        var authnState
        switch (state.selectedMFA) {
          case 'sms_otp':
            authnState = await client.authn.verifySMSOTP(state.authnState, code)
            break
          case 'totp':
            authnState = await client.authn.verifyTOTP(state.authnState, code)
            break
          case 'backup_code':
            authnState = await client.authn.verifyBackupCode(state.authnState, code)
            break
          default:
            throw new Error('selected MFA is unknown: ' + state.selectedMFA)
        }
        switch (authnState.status) {
          case 'MFA_REQUIRED':
            commit('SET_MFA_CODE_ERROR')
            break
          case 'SUCCESS':
            commit('SET_AUTHN_STATE', authnState)
            break
          default:
            commit('SET_ERROR', new Error('unexpected status ' + authnState.status))
        }
      } catch (err) {
        if (err.response) {
          if (err.response.status === 403) {
            commit('SET_MFA_CODE_ERROR')
            return
          }
        }
        commit('SET_ERROR', err)
      }
    },

    async signUp ({ commit, state }, { redirectURI, privacyCheckbox }) {
      try {
        commit('SET_LOADING')
        commit('UNSET_SIGN_UP_ERRORS')
        const signUpErrors = {}
        if (privacyCheckbox && !state.privacyChecked) {
          signUpErrors.privacy = i18n.t('error.accept_privacy_policy')
        }
        if (!state.handle) {
          signUpErrors.handle = i18n.t('register.input.error.invalid_contact')
        }
        if (state.passwordScore < 2) {
          signUpErrors.password = i18n.t('register.input.error.requires_better_password_strength')
        }
        if (Object.keys(signUpErrors).length > 0) {
          commit('SET_SIGN_UP_ERRORS', signUpErrors)
          return
        }

        const userInfo = {
          password_verifier: await client.utils.createPasswordVerifier(state.password)
        }
        if (isPossibleNumber(state.handle)) {
          userInfo.phone = state.handle
          commit('SET_HANDLE_TYPE', 'phone')
        } else {
          userInfo.email = state.handle
          commit('SET_HANDLE_TYPE', 'email')
        }
        const authnState = await client.client.signUp(redirectURI, userInfo)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        if (err.response !== undefined) {
          const signUpErrors = {}
          switch (err.response.status) {
            case 409:
              signUpErrors.handle = INPUT_HANDLE_ALREADY_EXISTS_ERROR
              commit('SET_SIGN_UP_ERRORS', signUpErrors)
              break
            case 400:
              // Get all the error fields from the response
              var errorFields = err.response.body.details[0].field_violations.map(field => field.field)
              var errorEmail = errorFields.includes('email')
              var errorPhone = errorFields.includes('phone')
              if (errorEmail) {
                signUpErrors.handle = i18n.t('register.input.error.invalid_email')
              }
              if (errorPhone) {
                signUpErrors.handle = i18n.t('register.input.error.invalid_phone')
              }
              commit('SET_SIGN_UP_ERRORS', signUpErrors)
              break
            default:
              commit('SET_ERROR', err) // unknown error
          }
        } else {
          commit('SET_ERROR', err) // unknown error
        }
      }
    },

    async startIDP ({ commit }, { idp, redirectURI, codeChallenge, codeChallengeMethod, clientState }) {
      try {
        commit('SET_LOADING')
        const authnState = await client.client.startIDP(idp, redirectURI, { codeChallenge, codeChallengeMethod, clientState })
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async verifyIDP ({ commit }, { stateToken, code }) {
      try {
        commit('SET_LOADING')
        const authnState = await client.client.verifyIDP(stateToken, code)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async startIDPBinding ({ commit }, { idp, redirectURI }) {
      try {
        commit('SET_LOADING')
        const authnState = await client.client.startIDPBinding(idp, redirectURI)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async verifyIDPBinding ({ commit }, { stateToken, code }) {
      try {
        commit('SET_LOADING')
        const authnState = await client.client.verifyIDPBinding(stateToken, code)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async verifyPasswordStepUp ({ commit, state }, password) {
      try {
        if (password) {
          commit('SET_PASSWORD', password)
        }
        commit('SET_LOADING')
        const authnState = await client.authn.verifyPasswordStepUp(state.password)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        var error = err
        if (err.response) {
          if (err.response.status === 403) {
            error = i18n.t('sign_in.input.error.incorrect_password')
          }
        }
        commit('SET_ERROR', error)
      }
    },

    async startPasswordReset ({ commit, state }, { handle }) {
      try {
        if (handle) {
          commit('SET_HANDLE', handle)
        }
        commit('SET_LOADING')
        const authnState = await client.client.startPasswordReset(state.handle)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        var error = err
        if (err.response) {
          if (err.response.status === 403) {
            error = i18n.t('sign_in.input.error.user_is_locked')
          } else if (err.response.status === 404) {
            error = INPUT_HANDLE_NOT_FOUND_ERROR
          }
        }
        commit('SET_ERROR', error)
      }
    },

    async verifyPasswordReset ({ commit, state }, { stateToken, resetToken }) {
      try {
        commit('SET_LOADING')
        if (resetToken) {
          commit('SET_RESET_TOKEN', resetToken)
        }
        let passwordVerifier
        if (state.password) {
          if (state.passwordScore < 2) {
            const err = i18n.t('change_password.input.error.requires_better_password_strength')
            commit('SET_ERROR', err)
            return
          }
          if (state.password !== state.passwordConfirmation) {
            const err = i18n.t('change_password.input.error.invalid_confirm_password')
            commit('SET_ERROR', err)
            return
          }
          passwordVerifier = await client.utils.createPasswordVerifier(state.password)
        }
        const authnState = await client.client.verifyPasswordReset(stateToken, state.resetToken, passwordVerifier)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        var error = err
        if (err.response) {
          if (err.response.status === 403) {
            error = i18n.t('reset_password.input.error.reset_link_expired')
          }
        }
        commit('SET_ERROR', error)
        commit('SET_RESET_TOKEN', '')
      }
    },

    async getState ({ commit }, stateToken) {
      try {
        commit('SET_LOADING')
        const authnState = await client.client.getAuthnState(stateToken)
        commit('SET_AUTHN_STATE', authnState)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },

    async signOut ({ commit }) {
      try {
        client.tokenManager.clear()
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  },

  mutations: {
    SET_HANDLE (state, value) {
      if (value !== undefined) {
        state.handle = value.trim()
      }
    },

    SET_HANDLE_TYPE (state, value) {
      state.handleType = value
    },

    SET_PASSWORD (state, value) {
      state.password = value
    },

    SET_PASSWORD_CONFIRMATION (state, value) {
      state.passwordConfirmation = value
    },

    SET_ONETIME_CODE (state, value) {
      state.onetimeCode = value
    },

    SET_SELECTED_MFA (state, value) {
      state.selectedMFA = value
    },

    UNSET_SELECTED_MFA (state) {
      state.selectedMFA = ''
    },

    SET_PASSWORD_SCORE (state, value) {
      state.passwordScore = value
    },

    SET_PRIVACY_CHECKED (state, value) {
      state.privacyChecked = value
    },

    SET_SIGN_UP_ERRORS (state, value) {
      state.signUpErrors = value
      state.loading = false
    },

    UNSET_SIGN_UP_ERRORS (state) {
      state.signUpErrors = null
    },

    SET_RECOVERY_EMAIL (state, value) {
      state.recoveryEmail = value
    },

    SET_RESET_TOKEN (state, value) {
      state.resetToken = value
    },

    RESET (state) {
      state.loading = false
      state.error = null
      state.authnState = null
      state.handle = ''
      state.handleType = ''
      state.password = ''
      state.onetimeCode = ''
      state.selectedMFA = ''
      state.signUpErrors = null
      state.recoveryEmail = ''
      state.recoveryEmailError = null
    },

    SET_LOADING (state) {
      state.loading = true
      state.error = null
    },

    SET_ERROR (state, err) {
      // The err is the localized error message to be displayed. Passing the error object to this
      // mutation is a convenient way to handle unknown error. Other error handling should not go
      // into this method.
      if (typeof err === 'object') {
        state.error = i18n.t('error.unknown')
        console.error('authentication transaction error', err)
      } else {
        state.error = err
      }
      state.loading = false
    },

    SET_MFA_CODE_ERROR (state) {
      state.loading = false
      switch (state.selectedMFA) {
        case 'sms_otp':
          state.error = i18n.t('sign_in.input.error.invalid_sms_code')
          break
        case 'totp':
          state.error = i18n.t('sign_in.input.error.invalid_totp_pin')
          break
        case 'backup_code':
          state.error = i18n.t('sign_in.input.error.invalid_backup_code')
          break
        default:
          console.error('selected MFA is unknown: ' + state.selectedMFA)
          state.error = i18n.t('error.unknown')
      }
    },

    SET_AUTHN_STATE (state, authnState) {
      state.loading = false
      state.authnState = authnState
    }
  }
}
