import client from '@/client'

function getDefaultState () {
  return {
    secondFactor: {},

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
    SET_SECOND_FACTOR (state, secondFactor) {
      state.secondFactor = secondFactor
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
    async unlink ({ commit }, id) {
      try {
        commit('SET_LOADING')
        await client.client.deleteMFA(id)
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
