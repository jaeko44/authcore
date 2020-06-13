import client from '@/client'
import { marshalUser } from '@/utils/marshal'

function getDefaultState () {
  return {
    user: {
      id: -1
    },

    authenticated: false,

    loading: false,
    error: undefined,
    actionError: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {
    isLoggedIn: function (state) {
      return state.user.profileName !== undefined
    },
    firstCharProfileName (state) {
      if (state.user.profileName) {
        return state.user.profileName[0].toUpperCase()
      }
      return ''
    }
  },

  mutations: {
    // Get current user
    SET_LOADING (state) {
      state.loading = true
    },
    SET_USER (state, currentUser) {
      state.user = marshalUser(currentUser)
      state.loading = false
      state.error = undefined
    },
    SET_ERROR (state, err) {
      state.loading = false
      state.error = err
    },
    SET_AUTHENTICATED (state, authenticated) {
      state.authenticated = authenticated
    },

    // Clear states
    CLEAR_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async get ({ commit }) {
      try {
        commit('SET_LOADING')
        const currentUser = await client.client.getCurrentUser()
        commit('SET_USER', currentUser)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
