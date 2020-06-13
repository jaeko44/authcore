import { cloneDeep } from 'lodash'

import { marshalRole } from '@/utils/marshal'

function getDefaultState () {
  return {
    roles: [
      { id: 1, name: 'authcore.admin' },
      { id: 2, name: 'authcore.editor' }
    ].map(marshalRole),

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
    // List roles
    SET_LOADING (state) {
      state.loading = true
    },
    SET_ROLES (state, payload) {
      const { roles, isMarshalNeeded } = payload
      if (isMarshalNeeded) {
        state.roles = cloneDeep(roles.map(marshalRole))
      } else {
        state.roles = cloneDeep(roles)
      }
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
  actions: {}
}
