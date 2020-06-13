import client from '@/client'

function getDefaultState () {
  return {
    oauthFactor: {},

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
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_OAUTH_FACTOR (state, oauthFactor) {
      state.oauthFactor = oauthFactor
    },
    SET_DONE (state) {
      state.done = true
    },
    SET_ERROR (state, error) {
      console.error(state)
      state.error = error
    },

    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async unlink ({ commit, rootState: { management } }, service) {
      try {
        commit('SET_LOADING')
        const { id: userId } = management.userDetails.user
        await client.client.deleteUserIDP(userId, service)
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
