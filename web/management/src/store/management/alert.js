export default {
  namespaced: true,

  state: {
    alert: {
      type: '',
      message: '',
      pending: false
    }
  },
  getters: {
    shownAlert: (state) => {
      return !!(state.alert.message && !state.alert.pending)
    }
  },

  mutations: {
    SET_MESSAGE (state, alert) {
      const alertKeys = Object.keys(alert)
      if (!alertKeys.includes('pending')) {
        throw new Error('missing \'pending\' key for the alert message')
      }
      state.alert = alert
    },
    UNSET_MESSAGE (state) {
      state.alert = {
        type: '',
        message: '',
        pending: false
      }
    },
    POP_MESSAGE (state, commit) {
      if (state.alert.message) {
        if (state.alert.pending) {
          // Show pending message in alert pane
          state.alert.pending = false
        } else {
          // Remove message in alert pane
          state.alert = {
            type: '',
            message: '',
            pending: false
          }
        }
      }
    }
  }
}
