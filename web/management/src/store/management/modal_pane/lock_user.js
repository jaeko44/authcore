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
    async lock ({ commit, rootState: { management } }) {
      try {
        commit('SET_LOADING')
        const { id: userId } = management.userDetails.user
        // Support infinity lock without description
        const updatedUser = await client.client.updateUser(userId, { is_locked: true })
        commit('SET_DONE')
        commit('management/userDetails/SET_USER', updatedUser, { root: true })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async unlock ({ commit, rootState: { management } }) {
      try {
        commit('SET_LOADING')
        const { id: userId } = management.userDetails.user
        const updatedUser = await client.client.updateUser(userId, { is_locked: false })
        commit('SET_DONE')
        commit('management/userDetails/SET_USER', updatedUser, { root: true })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
