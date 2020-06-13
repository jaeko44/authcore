import { i18n } from '@/i18n-setup'
import client from '@/client'

export default {
  namespaced: true,
  state () {
    return {
      loading: false,
      error: null,

      password: '',
      passwordConfirmation: '',
      passwordScore: -1
    }
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
      state.error = null
    },
    SET_ERROR (state, error) {
      console.error(error)
      state.loading = false
      state.error = error
    },
    SET_PASSWORD (state, value) {
      state.password = value
    },
    SET_PASSWORD_CONFIRMATION (state, value) {
      state.passwordConfirmation = value
    },
    SET_PASSWORD_SCORE (state, value) {
      state.passwordScore = value
    },
    RESET (state) {
      state.loading = false
      state.error = null
      state.password = ''
      state.passwordConfirmation = ''
      state.passwordScore = -1
    }
  },

  actions: {
    async changePassword ({ commit, dispatch, state, rootState: { authn } }) {
      if (!state.password) {
        const err = i18n.t('change_password.input.error.requires_better_password_strength')
        commit('SET_ERROR', err)
        return
      }
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
      try {
        commit('SET_LOADING')
        const passwordVerifier = await client.utils.createPasswordVerifier(state.password)
        await client.client.updateCurrentUserPassword(passwordVerifier)
        commit('RESET')
        commit('alert/SET_MESSAGE', {
          type: 'success',
          message: 'settings_home.message.success_change_password',
          pending: true
        }, { root: true })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
