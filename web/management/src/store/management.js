import alert from './management/alert'

import modalPane from './management/modal_pane'

import logsList from './management/logs_list'
import rolesList from './management/roles_list'
import usersList from './management/users_list'
import userDetails from './management/user_details'
import templatesList from './management/templates_list'

export default {
  namespaced: true,
  modules: {
    alert,
    modalPane,
    logsList,
    rolesList,
    usersList,
    userDetails,
    templatesList
  },

  state: {
    ready: false
  },
  getters: {},

  mutations: {
    SET_READY_STATE (state, ready) {
      state.ready = ready
    }
  },
  actions: {}
}
