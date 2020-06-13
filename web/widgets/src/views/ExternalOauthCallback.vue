<template>
  <!-- This is assumed to be opened by a window. -->
  <div>
    <form
      id="auth-callback-form"
      action="/redirect-oauth"
      method="GET">
      <input
        name="state"
        type="hidden"
        :value="pkceState" />
      <input
        name="code"
        type="hidden"
        :value="authorizationToken" />
      <input
        name="url"
        type="hidden"
        :value="redirectUri" />
    </form>
  </div>
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'
import router from '@/router'

export default {
  name: 'ExternalOauthCallback',
  components: {},

  props: {
    service: {
      type: String,
      required: true
    },
    code: {
      type: String,
      required: true
    },
    state: {
      type: String
    }
  },

  data () {
    return {
      redirectUri: undefined,
      pkceState: undefined,
      internalState: undefined
    }
  },

  computed: {
    ...mapState('client', [
      'authcoreClient',
      'inMobile'
    ]),
    ...mapState('widgets/errorPage', [
      'errorKey'
    ]),
    ...mapState('widgets/account/loginAccountForm', [
      'isCreateAccount',
      'createAccountInfo',
      'authorizationToken',
      'headerError',
      'redirectUrl'
    ])
  },

  watch: {},

  created () {},
  async mounted () {
    await this.init()
  },
  updated () {},
  destroyed () {},

  methods: {
    async init () {
      // `state` and `code` are OAuth2-specific parameters, while `token` is OAuth1-specific.
      const { service, state, code } = this
      const pkceState = sessionStorage.getItem('pkce_state')
      // redirectUrl is provided in OAuth case, which should be used in mobile case mainly
      const redirectUri = sessionStorage.getItem('redirect_uri')
      let successRedirectUrl = encodeURIComponent(sessionStorage.getItem('io.authcore.successRedirectUrl'))
      const fromSettings = sessionStorage.getItem('io.authcore.from_settings')
      sessionStorage.removeItem('io.authcore.from_settings')
      // Open new window flow only for mobile case with `successRedirectUrl` is set
      // Also from settings widgets as the flow should be run in normal browser
      // which allows multiple windows.
      if (this.inMobile && redirectUri === null && !fromSettings) {
        // Get back successRedirectUrl from sessionStorage
        // Using PKCE state as it is required in mobile apps case
        this.pkceState = pkceState
        const scope = sessionStorage.getItem('_authcore.scope')
        const clientId = sessionStorage.getItem('_authcore.clientId')
        const responseType = sessionStorage.getItem('_authcore.responseType')
        const codeChallenge = ''
        const codeChallengeMethod = ''
        const tempState = sessionStorage.getItem('io.authcore.temporary.oauth_state')
        store.commit('client/SET_OAUTH_WIDGET_STATE', {
          responseType, clientId, successRedirectUrl, scope, state, codeChallenge, codeChallengeMethod
        })
        if (state !== undefined && state !== tempState) {
          throw new Error('invalid oauth state')
        }
        await store.dispatch(
          'widgets/account/loginAccountForm/authenticateWithOAuth',
          { service, oauthState: tempState, code }
        )
        // redirectUrl should be provided after authenticateWithOAuth from server
        if (this.redirectUrl) {
          successRedirectUrl = this.redirectUrl
        }
        // Redirect Uri used for redirection flow
        // redirectUri is encoded to prevent injection in server.
        this.redirectUri = successRedirectUrl
        if (this.errorKey === 'sign_in.description.error.used_contact_in_system') {
          router.push({
            name: 'ErrorPage'
          })
        } else if (this.isCreateAccount) {
          // Decide which state is used. With PKCE, using that state to return
          // back to native app for the flow.
          const finalState = pkceState !== null ? pkceState : state
          router.replace({
            name: 'Register',
            query: {
              scope: scope,
              client_id: clientId,
              state: finalState,
              redirect_uri: successRedirectUrl,
              response_type: responseType,
              code_challenge: codeChallenge,
              code_challenge_method: codeChallengeMethod
            }
          })
        } else {
          // Using form submission for redirection
          this.logAnalytics('Authcore_loginSuccess', {}, true)
          setTimeout(() => {
            document.getElementById('auth-callback-form').submit()
          }, 10)
        }
      } else {
        // Desktop case, using postMessage to pass information back to SignIn page
        window.opener.postMessage({
          type: 'AuthCore_externalOAuthCallback',
          data: { service, state, code, pkceState }
        }, window.location.origin)
        window.close()
      }
    }
  }
}
</script>
