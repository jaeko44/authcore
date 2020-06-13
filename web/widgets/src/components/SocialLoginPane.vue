<template>
  <social-login-pane-list
    v-if="socialLoginPaneOption === 'list'"
    :register="register"
    :openOAuthWindow="openOAuthWindow"
    :social-login-list="socialLoginList"
  />
  <social-login-pane-grid
    v-else-if="socialLoginPaneOption === 'grid'"
    :openOAuthWindow="openOAuthWindow"
    :social-login-list="socialLoginList"
  />
</template>

<script>
import { mapState, mapActions } from 'vuex'

import { openOAuthWindow } from '@/utils/util'

import SocialLoginPaneGrid from '@/components/SocialLoginPaneGrid.vue'
import SocialLoginPaneList from '@/components/SocialLoginPaneList.vue'

export default {
  name: 'SocialLoginPane',
  components: {
    SocialLoginPaneGrid,
    SocialLoginPaneList
  },
  props: {
    socialLoginList: {
      type: Array,
      default: () => []
    },
    socialLoginPaneOption: {
      type: String,
      required: true
    },
    // Props for wording in list button
    register: {
      type: Boolean,
      default: false
    }
  },

  data () {
    return {
      closeOAuthWindowFunc: null,
      marginForSocialLogo: 16,
      socialScrollRow: false
    }
  },

  computed: {
    ...mapState('preferences', [
      'redirectURI',
      'containerId'
    ]),

    ...mapState('authn', [
      'authnState',
      'error'
    ])
  },

  mounted () {
    window.addEventListener('beforeunload', this.closeOAuthWindow)
  },

  beforeDestroy () {
    window.removeEventListener('beforeunload', this.closeOAuthWindow)
  },

  methods: {
    ...mapActions('authn', [
      'startIDP'
    ]),

    async openOAuthWindow (service) {
      this.logAnalytics('Authcore_oauthStarted', { service })
      sessionStorage.setItem('_authcore.scope', this.scope) // this.scope is uninitialized
      sessionStorage.setItem('_authcore.clientId', this.clientId) // this.clientId is uninitialized
      sessionStorage.setItem('_authcore.responseType', this.responseType)
      const query = this.$route.query
      const codeChallenge = query.codeChallenge
      const codeChallengeMethod = query.codeChallengeMethod
      this.closeOAuthWindowFunc = await openOAuthWindow(this.containerId, service, async () => {
        await this.startIDP({ idp: service, redirectURI: this.redirectURI, codeChallenge, codeChallengeMethod })
        if (this.error) {
          throw new Error('error starting IDP authentication')
        }
        if (this.authnState.status === 'IDP') {
          const endpointUri = this.authnState.idp_authorization_url
          const state = this.authnState.state_token
          // Set the state token to allow resumption after redirection
          sessionStorage.setItem('io.authcore.authn_state.resume', state)
          return endpointUri
        }
        throw new Error('illegal state while starting IDP authentication')
      })
    },

    closeOAuthWindow () {
      if (this.closeOAuthWindowFunc) {
        this.closeOAuthWindowFunc()
      }
    }
  }
}
</script>
