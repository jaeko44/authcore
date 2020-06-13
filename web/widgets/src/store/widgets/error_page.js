import {
  SET_ERROR,

  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
    loading: false,

    errorKey: '',
    errorMessage: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),

  mutations: {
    [SET_ERROR] (state, { key, message }) {
      state.errorKey = key
      state.errorMessage = message
    },
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  }
}
