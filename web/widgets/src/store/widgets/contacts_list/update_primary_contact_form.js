import {
  UPDATE_PRIMARY_CONTACT_SPINNER_STATE,
  UPDATE_PRIMARY_CONTACT_STARTED,
  UPDATE_PRIMARY_CONTACT_COMPLETED,
  UPDATE_PRIMARY_CONTACT_FAILED,
  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
    loading: false,
    error: undefined,
    done: false
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [UPDATE_PRIMARY_CONTACT_SPINNER_STATE] (state, loading) {
      state.spinnerLoading = loading
    },
    [UPDATE_PRIMARY_CONTACT_STARTED] (state) {
      state.loading = true
      state.done = false
    },
    [UPDATE_PRIMARY_CONTACT_COMPLETED] (state) {
      state.loading = false
      state.done = true
      state.error = undefined
    },
    [UPDATE_PRIMARY_CONTACT_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    // Clear states
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async updatePrimaryContact ({ commit, rootState: { client = undefined } }, contactId) {
      try {
        commit(UPDATE_PRIMARY_CONTACT_STARTED)
        await client.authcoreClient.updatePrimaryContact(contactId)
        commit(UPDATE_PRIMARY_CONTACT_COMPLETED)
      } catch (err) {
        commit(UPDATE_PRIMARY_CONTACT_FAILED, err)
      }
    }
  }
}
