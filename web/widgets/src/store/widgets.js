import { SET_DISPLAY_MODE_STATE } from '@/store/types'

import contactsList from './widgets/contacts_list'
import account from './widgets/account'
import externalOauth from './widgets/external_oauth'
import errorPage from './widgets/error_page'

export default {
  namespaced: true,
  modules: {
    contactsList,
    account,
    externalOauth,
    errorPage
  },

  state: {
    displayMode: undefined
  },
  getters: {},

  mutations: {
    [SET_DISPLAY_MODE_STATE] (state, option) {
      state.displayMode = option
    }
  },
  actions: {}
}
