import cloneDeep from 'lodash-es/cloneDeep'

import {
  AUTHENTICATE_HANDLE_STARTED,
  AUTHENTICATE_HANDLE_COMPLETED,
  AUTHENTICATE_HANDLE_FAILED,
  AUTHENTICATE_FACTOR_INIT,
  AUTHENTICATE_FACTOR_STARTED,
  AUTHENTICATE_FACTOR_COMPLETED,
  AUTHENTICATE_FACTOR_FAILED,
  RESET_PASSWORD_STARTED,
  RESET_PASSWORD_COMPLETED,
  RESET_PASSWORD_FAILED,

  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
    challenges: [],
    selectedChallengeMethod: undefined,
    authorizationToken: undefined,
    authenticateHandleDone: false,
    authenticateFactorDone: false,
    resetPasswordDone: false,
    loading: false,
    error: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [AUTHENTICATE_HANDLE_STARTED] (state) {
      state.loading = true
      state.authenticateHandleDone = false
      state.error = undefined
    },
    [AUTHENTICATE_HANDLE_COMPLETED] (state) {
      state.loading = false
      state.authenticateHandleDone = true
    },
    [AUTHENTICATE_HANDLE_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
      if (err.response !== undefined) {
        // TODO: Change to corresponding error message. Invalid password as only SPAKE2 is used currently
        switch (err.response.status) {
          case 404:
            state.error = 'reset_password.input.error.no_contact'
            break
          case 429:
            // Error contains many if limit is reached
            if (/many/.test(err.response.body.error)) {
              state.error = 'reset_password.text.error.reach_limit'
            }
            break
        }
      }
    },
    [AUTHENTICATE_FACTOR_INIT] (state, challenges) {
      // Store the login challenges from the response
      state.loading = false
      state.challenges = cloneDeep(challenges)
      state.selectedChallengeMethod = cloneDeep(challenges[0])
    },
    [AUTHENTICATE_FACTOR_STARTED] (state) {
      state.loading = true
      state.authenticateFactorDone = false
      state.error = undefined
    },
    [AUTHENTICATE_FACTOR_COMPLETED] (state, authorizationToken) {
      state.loading = false
      state.authorizationToken = authorizationToken
      state.authenticateFactorDone = true
    },
    [AUTHENTICATE_FACTOR_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
      if (err.response !== undefined) {
        switch (err.response.status) {
          case 404:
            state.error = 'reset_password.text.error.invalid_reset_password'
            break
        }
      }
    },
    [RESET_PASSWORD_STARTED] (state) {
      state.loading = true
      state.resetPasswordDone = false
      state.error = undefined
    },
    [RESET_PASSWORD_COMPLETED] (state) {
      state.loading = false
      state.resetPasswordDone = true
    },
    [RESET_PASSWORD_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async _authenticateInit ({ state, commit, rootState: { client = undefined } }, challenges) {
      commit(AUTHENTICATE_FACTOR_INIT, challenges)
      // Start SMS authentication flow if necessary
      if (state.selectedChallengeMethod === 'SMS_CODE') {
        await client.authcoreClient.startAuthenticateSMS()
      }
    },
    async authenticateHandle ({ commit, dispatch, rootState: { client = undefined } }, handle) {
      try {
        commit(AUTHENTICATE_HANDLE_STARTED)
        const res = await client.authcoreClient.startResetPasswordAuthentication(handle)
        commit(AUTHENTICATE_HANDLE_COMPLETED)
        await dispatch('_authenticateInit', res.challenges)
      } catch (err) {
        commit(AUTHENTICATE_HANDLE_FAILED, err)
      }
    },
    async authenticateWithContact ({ commit, dispatch, rootState: { client = undefined } }, contactToken) {
      try {
        commit(AUTHENTICATE_FACTOR_STARTED)
        const updateAuthenticationResponse = await client.authcoreClient.authenticateResetPasswordWithContact(contactToken)
        if (updateAuthenticationResponse.authenticated) {
          // Authenticated
          const authorizationToken = updateAuthenticationResponse.reset_password_token
          commit(AUTHENTICATE_FACTOR_COMPLETED, authorizationToken)
        } else {
          await dispatch('_authenticateInit', updateAuthenticationResponse.challenges)
        }
      } catch (err) {
        commit(AUTHENTICATE_FACTOR_FAILED, err)
      }
    },
    async resetPassword ({ state, commit, rootState: { client = undefined } }, { password, confirmPassword }) {
      try {
        commit(RESET_PASSWORD_STARTED)
        if (password !== confirmPassword) {
          throw new Error('passwords do not match')
        }
        await client.authcoreClient.resetPassword(state.authorizationToken, password)
        commit(RESET_PASSWORD_COMPLETED)
      } catch (err) {
        commit(RESET_PASSWORD_FAILED, err)
      }
    }
  }
}
