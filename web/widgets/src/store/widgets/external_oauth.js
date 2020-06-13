import {
  OPEN_OAUTH_AUTHENTICATOR_STARTED,
  OPEN_OAUTH_AUTHENTICATOR_COMPLETED,
  OPEN_OAUTH_AUTHENTICATOR_FAILED,

  CLEAR_STATES
} from '@/store/types'

import { redirectTo } from '@/utils/util'

function getDefaultState () {
  return {
    loading: false,
    error: undefined,
    done: false,
    oauthState: undefined
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [OPEN_OAUTH_AUTHENTICATOR_STARTED] (state) {
      state.loading = true
      state.error = undefined
    },
    [OPEN_OAUTH_AUTHENTICATOR_COMPLETED] (state, oauthState) {
      state.loading = false
      state.done = true
      state.oauthState = oauthState
    },
    [OPEN_OAUTH_AUTHENTICATOR_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async openOAuthAuthenticator ({ commit, rootState: { client = undefined } }, { purpose, service, inMobile }) {
      let oauthWindow
      try {
        const sizes = {
          google: { height: 600, width: 400 },
          facebook: { height: 500, width: 400 },
          apple: { height: 750, width: 800 },
          matters: { height: 700, width: 500 },
          twitter: { height: 550, width: 720 }
        }
        commit('widgets/account/loginAccountForm/CLEAR_STATES', null, { root: true })
        commit(OPEN_OAUTH_AUTHENTICATOR_STARTED)
        // Open new window for mobile case with `successRedirectUrl` is set
        if (!inMobile || !client.successRedirectUrl) {
          oauthWindow = window.open('about:blank', 'window', `height=${sizes[service].height},width=${sizes[service].width},menubar=no`)
        }
        let endpointUri, state
        if (purpose === 'authenticate') {
          const res = await client.authcoreClient.startAuthenticateOAuth(service, client.successRedirectUrl)
          endpointUri = res.endpointUri
          state = res.state
        } else if (purpose === 'create') {
          const res = await client.authcoreClient.startCreateOAuthFactor(service)
          endpointUri = res.endpointUri
          state = res.state
        } else {
          throw new Error('undefined purpose')
        }
        sessionStorage.setItem('io.authcore.temporary.oauth_state', state)
        // TODO: use postMessage to handle the redirected location for the OAuth window.
        // https://gitlab.com/blocksq/authcore/issues/509
        if (inMobile && client.successRedirectUrl) {
          // Redirect the parent location using the same window
          redirectTo(endpointUri, client.containerId)
        } else {
          oauthWindow.location = endpointUri
        }
        commit(OPEN_OAUTH_AUTHENTICATOR_COMPLETED, state)
      } catch (err) {
        console.error(err)
        if (!inMobile || !client.successRedirectUrl) {
          oauthWindow.close()
        }
        commit(OPEN_OAUTH_AUTHENTICATOR_FAILED, err)
      }
    }
  }
}
