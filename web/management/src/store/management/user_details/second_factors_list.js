import client from '@/client'
import { marshalSecondFactor } from '@/utils/marshal'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined,

    secondFactors: []
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {
    firstlyCreatedFactor (state) {
      if (state.secondFactors.length > 0) {
        return state.secondFactors.sort(function (a, b) {
          // Earliest created_at be the first one
          return new Date(a.created_at) - new Date(b.created_at)
        })[0]
      }
      return {
        createdAt: ''
      }
    }
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_USER_SECOND_FACTORS (state, secondFactors) {
      state.secondFactors = secondFactors.map(marshalSecondFactor)
    },
    SET_ERROR (state, error) {
      console.error(error)
      state.error = error
    },

    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async fetchList ({ commit }, id) {
      try {
        commit('SET_LOADING')
        const secondFactors = await client.client.listUserMFA(id)
        commit('SET_USER_SECOND_FACTORS', secondFactors.results || [])
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
