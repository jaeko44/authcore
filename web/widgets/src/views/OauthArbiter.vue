<template>
  <div class="pt-5">
    <loading-spinner />
    <form
      ref="form"
      action="/arbiter-redirect"
      method="GET">
      <input
        name="redirect_uri"
        type="hidden"
        :value="redirectURI" />
      <input
        name="code"
        type="hidden"
        :value="authorizationCode" />
      <input
        name="client_id"
        type="hidden"
        :value="clientId" />
    </form>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import store from '@/store'
import router from '@/router'
import { redirectTo } from '@/utils/util'

import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'OauthArbiter',
  components: {
    LoadingSpinner
  },

  props: {
    state: {
      required: true,
      type: String
    },
    code: {
      type: String
    },
    oauthVerifier: {
      type: String
    }
  },

  computed: {
    ...mapState('preferences', [
      'inMobile'
    ]),

    ...mapState('authn', [
      'authnState',
      'error',
      'loading'
    ]),

    redirectURI () {
      if (this.authnState) {
        if (this.authnState.status === 'SUCCESS') {
          var url = new URL(this.authnState.redirect_uri)
          url.searchParams.set('code', this.authorizationCode)
          url.searchParams.set('state', this.clientState)
          return url.toString()
        } else if (this.authnState.status === 'IDP_BINDING_SUCCESS') {
          return this.authnState.redirect_uri
        }
      }
      return ''
    },

    authorizationCode () {
      return this.authnState ? this.authnState.authorization_code : ''
    },

    clientId () {
      return this.authnState ? this.authnState.client_id : ''
    },

    clientState () {
      return this.authnState ? this.authnState.client_state : ''
    }
  },

  async mounted () {
    // This page is called when an external IdP approved a login and redirect back to Authcore.
    // There are several cases to continue the login flow:
    //
    // 1. Current frame is a new window opened by widget frame for social login screens. Use
    //    postMessage to ask to open this view in widget frame. The new window will be closed
    //    immediately.
    //
    // 2. Login success:
    //    a. In mobile (i.e. current frame is the main frame), use HTML form to redirect. The form is
    //       to work around an issue in Android Chrome that blocks redirection to apps.
    //    b. In desktop where the widget may be contained in an iframe, use redirectTo utils method to
    //       redirect.
    //
    // 3. IdP login success, no Authcore account is found, and email is not taken. Open Register view.
    //
    // 4. IdP login success, no Authcore account is found, and email is already taken. Open ErrorPage.

    if (window.opener) {
      // Desktop case, using postMessage to pass information back to SignIn page
      window.opener.postMessage({
        type: 'Authcore_openOauthArbiter',
        data: { state: this.state, code: this.code, oauthVerifier: this.oauthVerifier }
      }, window.location.origin)
      window.close()
      return
    }

    const stateResumption = sessionStorage.getItem('io.authcore.authn_state.resume')
    if (this.state !== stateResumption) {
      throw new Error('cannot resume state from redirection')
    }

    const code = this.oauthVerifier || this.code

    await this.getState(this.state)
    if (this.error) {
      return
    }

    if (this.authnState.status === 'IDP') {
      await this.verifyIDP({ stateToken: this.state, code })
    } else if (this.authnState.status === 'IDP_BINDING') {
      await this.verifyIDPBinding({ stateToken: this.state, code })
    }
    if (this.error) {
      return
    }

    if (this.authnState.status === 'SUCCESS') {
      this.logAnalytics('Authcore_loginSuccess', {}, true)
      this.redirectToDestination()
    } else if (this.authnState.status === 'IDP_BINDING_SUCCESS') {
      this.redirectToDestination()
    } else if (this.authnState.status === 'IDP_ALREADY_EXISTS') {
      store.commit('widgets/errorPage/SET_ERROR', {
        key: 'sign_in.description.error.used_contact_in_system',
        message: ''
      })
      router.push({
        name: 'ErrorPage'
      })
    } else {
      throw new Error('illegal authentication state')
    }
  },

  methods: {
    ...mapActions('authn', [
      'verifyIDP',
      'verifyIDPBinding',
      'getState'
    ]),

    redirectToDestination () {
      if (this.inMobile) {
        setTimeout(() => {
          this.$refs.form.submit()
        }, 10)
      } else {
        // Redirection flow for desktop case
        redirectTo(this.redirectURI, this.containerId)
      }
    }
  }
}
</script>
