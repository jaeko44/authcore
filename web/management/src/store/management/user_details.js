import client from '@/client'
import { marshalUser } from '@/utils/marshal'

import oauthFactorsList from './user_details/oauth_factors_list'
import secondFactorsList from './user_details/second_factors_list'
import logsList from './user_details/logs_list'
import updateUserForm from './user_details/update_user_form'
import devicesList from './user_details/devices_list'
import role from './user_details/role'
import metadata from './user_details/metadata'

function getDefaultState () {
  return {
    user: {},

    loading: false,
    error: undefined
  }
}

export default {
  namespaced: true,
  modules: {
    oauthFactorsList,
    secondFactorsList,
    logsList,
    updateUserForm,
    devicesList,
    role,
    metadata
  },

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_USER (state, user) {
      state.loading = false
      state.user = marshalUser(user)
    },
    SET_ERROR (state, error) {
      state.loading = false
      state.error = error
    },
    // Clear states
    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async get ({ commit }, id) {
      try {
        commit('SET_LOADING')
        const user = await client.client.getUser(id)
        commit('SET_USER', user)
      } catch (err) {
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    }
  }
}
