import cloneDeep from 'lodash/cloneDeep'

import client from '@/client'
import { marshalSession } from '@/utils/marshal'

function getDefaultState () {
  return {
    sessions: [],
    prevPageToken: '',
    nextPageToken: '',
    totalItems: -1,

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
      state.error = undefined
    },
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_DONE (state) {
      state.loading = false
      state.done = true
    },
    SET_ERROR (state, err) {
      console.log(err)
      state.loading = false
      state.error = err
    },

    SET_SESSIONS (state, payload) {
      const {
        results,
        previous_page_token: prevPageToken,
        next_page_token: nextPageToken,
        total_size: totalSize
      } = payload
      state.sessions = results.map(marshalSession)
      state.prevPageToken = cloneDeep(prevPageToken)
      state.nextPageToken = cloneDeep(nextPageToken)
      state.totalItems = totalSize
    },
    DELETE_SESSION (state, payload) {
      const { sessionId } = payload
      state.sessions = state.sessions.filter(item => item.id !== sessionId)
      state.totalItems = state.totalItems - 1
    },

    CLEAR_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async list ({ commit, rootState: { management } }, payload) {
      const { rowsPerPage, pageToken } = payload
      try {
        commit('SET_LOADING')
        const { id: userId } = management.userDetails.user
        const response = await client.client.listUserSessions(userId, pageToken, rowsPerPage)
        commit('SET_SESSIONS', response)
        commit('SET_DONE')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async delete ({ commit }, sessionId) {
      try {
        commit('SET_LOADING')
        await client.client.deleteSession(sessionId)
        commit('DELETE_SESSION', {
          sessionId: sessionId
        })
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
