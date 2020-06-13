import { i18n } from '@/i18n-setup'

import client from '@/client'

export default {
  namespaced: true,

  state () {
    return {
      secondFactors: [],
      loading: false,
      error: undefined,
      totpSecret: '',
      totpCode: ''
    }
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_SECOND_FACTORS_LIST (state, secondFactors) {
      state.loading = false
      if (secondFactors) {
        secondFactors = secondFactors.sort(function (a, b) {
          // sort by last used at
          if (a.last_used_at > b.last_used_at) {
            return -1
          }
          if (a.last_used_at < b.last_used_at) {
            return 1
          }
          return 0
        })
        state.secondFactors = secondFactors
      } else {
        state.secondFactors = []
      }
    },
    SET_TOTP_SECRET (state, secret) {
      state.totpSecret = secret
    },
    SET_TOTP_CODE (state, code) {
      state.totpCode = code
    },
    SET_ERROR (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    RESET (state) {
      Object.assign(state, {
        secondFactors: [],
        loading: false,
        error: undefined,
        totpSecret: '',
        totpCode: ''
      })
    }
  },
  actions: {
    async list ({ commit }) {
      try {
        commit('SET_LOADING')
        const secondFactors = await client.client.listCurrentUserMFA()
        commit('SET_SECOND_FACTORS_LIST', secondFactors.results)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async delete ({ commit }, factorId) {
      try {
        commit('SET_LOADING')
        await client.client.deleteCurrentUserMFA(factorId)
        commit('UNSET_LOADING')
        commit('alert/SET_MESSAGE', {
          type: 'success',
          message: 'manage_authenticator_app.message.success_remove_authenticator_app',
          pending: true
        }, { root: true })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async generateTOTPSecret ({ commit }) {
      const totpSecret = await client.utils.generateTOTPSecret()
      commit('SET_TOTP_SECRET', totpSecret)
    },
    async createTOTP ({ commit, state }) {
      try {
        commit('SET_LOADING')
        await client.client.createCurrentUserMFA({
          type: 'totp',
          secret: state.totpSecret,
          verifier: btoa(state.totpCode)
        })
      } catch (err) {
        let error
        if (err.response !== undefined) {
          switch (err.response.status) {
            case 400:
              error = i18n.t('mfa_totp_create.input.error.invalid_verification_code')
              break
          }
        } else {
          error = err
        }
        commit('SET_ERROR', error)
      }
    }
  }
}
