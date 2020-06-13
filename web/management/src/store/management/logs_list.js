import { cloneDeep } from 'lodash'

import {
  LIST_LOGS_STARTED,
  LIST_LOGS_COMPLETED,
  LIST_LOGS_FAILED,
  CLEAR_STATES
} from '@/store/types'
import { marshalAuditLog } from '@/utils/marshal'
import client from '@/client'

import { ITEM_PER_PAGE_FOR_PLACEHOLDER } from '@/config'

function getDefaultState () {
  return {
    logs: [],
    prevPageToken: '',
    nextPageToken: '',
    totalItems: -1,

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
    // List logs
    [LIST_LOGS_STARTED] (state) {
      state.loading = true
      // Provide empty object to show loading content
      state.logs = new Array(ITEM_PER_PAGE_FOR_PLACEHOLDER).fill({})
    },
    [LIST_LOGS_COMPLETED] (state, payload) {
      const { logs } = payload
      state.logs = logs.audit_logs.map(marshalAuditLog)
      state.prevPageToken = cloneDeep(logs.previous_page_token)
      state.nextPageToken = cloneDeep(logs.next_page_token)
      state.totalItems = logs.total_size

      state.loading = false
      state.error = undefined
    },
    [LIST_LOGS_FAILED] (state, err) {
      state.loading = false
      state.error = err
    },

    // Clear states
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async list ({ commit }, payload) {
      const { userId, rowsPerPage, pageToken } = payload
      try {
        commit(LIST_LOGS_STARTED)
        const logsObj = await client.client.listUserEvents(userId, pageToken, rowsPerPage)
        commit(LIST_LOGS_COMPLETED, {
          logs: logsObj
        })
      } catch (err) {
        commit(LIST_LOGS_FAILED, err)
      }
    }
  }
}
