import {
  DELETE_CONTACT_SPINNER_STATE,
  DELETE_CONTACT_STARTED,
  DELETE_CONTACT_COMPLETED,
  DELETE_CONTACT_FAILED,
  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
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
    [DELETE_CONTACT_SPINNER_STATE] (state, loading) {
      state.spinnerLoading = loading
    },
    [DELETE_CONTACT_STARTED] (state) {
      state.loading = true
      state.done = false
    },
    [DELETE_CONTACT_COMPLETED] (state) {
      state.loading = false
      state.done = true
      state.error = undefined
    },
    [DELETE_CONTACT_FAILED] (state, err) {
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
    async delete ({ commit, dispatch, rootState: { client = undefined } }, contactId) {
      try {
        commit(DELETE_CONTACT_STARTED)
        await client.authcoreClient.deleteContact(contactId)
        commit(DELETE_CONTACT_COMPLETED)
      } catch (err) {
        commit(DELETE_CONTACT_FAILED, err)
      }
    }
  }
}
