import client from '@/client'

import { marshalOAuthFactor } from '@/utils/marshal'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined,

    oAuthFactors: []
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {
    firstlyCreatedFactor (state) {
      if (state.oAuthFactors.length > 0) {
        return state.oAuthFactors.sort(function (a, b) {
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
    SET_USER_OAUTH_FACTORS (state, oAuthFactors) {
      state.oAuthFactors = oAuthFactors.map(marshalOAuthFactor)
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
    async fetchList ({ commit }, id) {
      try {
        commit('SET_LOADING')
        const oAuthFactors = await client.client.listUserIDP(id)
        commit('SET_USER_OAUTH_FACTORS', oAuthFactors.results || [])
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async delete ({ commit }, id) {
      try {
        commit('SET_LOADING')
        await client.client.deleteUserIDP(id)
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
