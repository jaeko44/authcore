import client from '@/client'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_DONE (state) {
      state.loading = false
      state.done = true
    },
    SET_ERROR (state, error) {
      console.error(state)
      state.loading = false
      state.error = error
    },

    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async changePassword ({ commit, rootState: { management } }, payload) {
      try {
        const { id: userId } = management.userDetails.user
        const { password, confirmPassword } = payload
        // const { newPassword, confirmNewPassword } = payload
        commit('SET_LOADING')
        if (password === '') throw new Error('the new password should not be empty')
        // TODO: password meter checking
        if (password !== confirmPassword) throw new Error('passwords are not equal')
        const passwordVerifier = await client.utils.createPasswordVerifier(password)
        await client.client.updateUserPassword(userId, passwordVerifier)
        commit('SET_DONE')
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_details_security.message.change_password_successfully',
          pending: false
        }, { root: true })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
