import client from '@/client'
import { marshalUser } from '@/utils/marshal'

import createUserForm from './users_list/create_user_form'

function getDefaultState () {
  return {
    users: [],
    prevPageToken: '',
    nextPageToken: '',
    queryObject: {},
    totalItems: -1,

    loading: false,
    error: undefined,
    actionError: undefined
  }
}

export default {
  namespaced: true,
  modules: {
    createUserForm
  },

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_QUERY_OBJECT (state, queryObject) {
      state.queryObject = queryObject
    },
    // List users
    SET_LOADING (state) {
      state.loading = true
    },
    SET_USER_LIST (state, payload) {
      const { users } = payload
      state.users = users.results.map(marshalUser)
      state.prevPageToken = users.prev_page_token || ''
      state.nextPageToken = users.next_page_token || ''
      state.totalItems = users.total_size

      state.loading = false
      state.error = undefined
    },
    SET_ERROR (state, err) {
      state.loading = false
      state.error = err
    },

    // Clear states
    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async fetchList ({ commit }, payload) {
      commit('SET_QUERY_OBJECT', payload)
      const {
        limit = 10,
        pageToken,
        sortKey,
        queryKey,
        queryValue,
        ascending
      } = payload

      try {
        commit('SET_LOADING')
        const options = {}
        if (queryValue !== '') {
          options[queryKey] = queryValue
        }
        const asc = ascending ? 'asc' : 'desc'
        const sortBy = `${sortKey} ${asc}`
        const usersObj = await client.client.listUsers(pageToken, limit, sortBy, options)
        commit('SET_USER_LIST', {
          users: usersObj
        })
      } catch (err) {
        console.error(err)
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
