import assign from 'lodash-es/assign'
import merge from 'lodash-es/merge'

import { inMobile } from '@/utils/util.js'
import router from '@/router.js'

export default {
  namespaced: true,

  state: {
    initialized: false,
    signUpEnabled: true,
    idpList: [],
    clientId: '',
    redirectURI: '',
    logo: '',
    company: '',
    containerId: '',
    primaryColor: undefined,
    successColor: undefined,
    dangerColor: undefined,
    privacyURL: '',
    privacyCheckbox: false,
    buttonSize: 'large',
    socialLoginPaneOption: 'grid',
    socialLoginPaneStyle: 'bottom',
    isInternal: false,
    isMobile: false
  },

  mutations: {
    INIT (state, { preferences }) {
      assign(state, preferences)
      // Normalize values
      if (state.idpList && state.idpList.length === 0) {
        state.idpList = null
      }
    }
  },

  getters: {
    isAdminPortal (state) {
      return state.clientId === '_authcore_admin_portal_'
    }
  },

  actions: {
    init ({ commit, state }, presetPreferences) {
      const preferences = Object.assign({}, presetPreferences)
      if (state.initialized) {
        return
      }
      const query = router.currentRoute.query
      const queryPreferences = {
        initialized: true,
        clientId: query.clientId,
        redirectURI: query.redirectURI || query.successRedirectUrl,
        logo: query.logo,
        company: query.company,
        containerId: query.cid,
        primaryColor: query.primaryColor,
        successColor: query.successColor,
        dangerColor: query.dangerColor,
        privacyURL: query.privacy,
        privacyCheckbox: query.privacyCheckbox === 'true',
        buttonSize: query.buttonSize,
        socialLoginPaneOption: query.socialLoginPaneOption || 'grid',
        socialLoginPaneStyle: query.socialLoginPaneStyle || 'bottom',
        isInternal: query.internal,
        isMobile: inMobile()
      }
      merge(preferences, queryPreferences)

      commit('INIT', { preferences })
    }
  }
}
