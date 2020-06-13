import {
  UPDATE_PROFILE_SPINNER_STATE,
  UPDATE_PROFILE_STARTED,
  UPDATE_PROFILE_COMPLETED,
  UPDATE_PROFILE_FAILED,

  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined,

    spinnerLoading: false
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [UPDATE_PROFILE_SPINNER_STATE] (state, loading) {
      state.spinnerLoading = loading
    },
    [UPDATE_PROFILE_STARTED] (state) {
      state.loading = true
      state.done = false
      state.error = undefined
    },
    [UPDATE_PROFILE_COMPLETED] (state) {
      state.loading = true
      state.done = true
    },
    [UPDATE_PROFILE_FAILED] (state, err) {
      state.loading = false
      state.error = err
      if (err.response !== undefined) {
        switch (err.response.status) {
          case 400:
            state.error = 'profile_edit.input.error.not_updated_username'
            break
          case 500:
            state.error = 'error.unknown'
            break
        }
      } else {
        // Error returned from authcore client for case with no displayName
        state.usernameError = 'error.blank'
      }
    },
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async set ({ commit, rootState: { client = undefined } }, user) {
      try {
        commit(UPDATE_PROFILE_STARTED)
        const updatedUser = await client.authcoreClient.updateCurrentUser(user)
        commit(UPDATE_PROFILE_COMPLETED)
        commit('widgets/account/SET_PROFILE_COMPLETED', updatedUser, { root: true })
      } catch (err) {
        commit(UPDATE_PROFILE_FAILED, err)
      }
    }
  }
}
