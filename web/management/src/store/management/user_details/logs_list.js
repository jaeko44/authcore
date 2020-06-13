import { marshalAuditLog } from '@/utils/marshal'

import client from '@/client'

function getDefaultState () {
  return {
    logs: [],
    prevPageToken: '',
    nextPageToken: '',
    totalItems: -1,
    currentPage: 1,

    loading: false,
    error: undefined,
    actionError: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_QUERY_OBJECT (state, queryObject) {
      state.queryObject = queryObject
    },
    // List user logs
    SET_LOADING (state) {
      state.loading = true
    },
    SET_USER_LOG_LIST (state, logs) {
      state.logs = logs.results.map(marshalAuditLog)
      state.prevPageToken = logs.prev_page_token
      state.nextPageToken = logs.next_page_token
      state.totalItems = logs.total_size

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
      const { userId, pageToken } = payload
      try {
        commit('SET_LOADING')
        const logs = await client.client.listUserEvents(userId, pageToken, 10)
        commit('SET_USER_LOG_LIST', logs)
      } catch (err) {
        console.log(err)
        commit('SET_ERROR', err)
      }
    }
  }
}
