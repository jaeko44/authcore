<template>
  <widget-layout-v2
    large-header
    :back-button-enabled="shouldShowBackButton"
    :title="currentTitle"
    :description="currentDescription"
    @back-button="RESET()"
    >
    <component :is="currentPane"/>
  </widget-layout-v2>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

import { redirectTo } from '@/utils/util'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import StartPane from '@/views/signin/StartPane.vue'
import PasswordPane from '@/views/signin/PasswordPane.vue'
import MFAPane from '@/views/signin/MFAPane.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'SignInView',
  components: {
    WidgetLayoutV2
  },
  computed: {
    ...mapState('authn', [
      'authnState',
      'selectedMFA',
      'error',
      'loading'
    ]),

    currentPane () {
      if (!this.authnState || this.authnState.status === 'IDP') {
        return StartPane
      } else if (this.authnState.status === 'PRIMARY') {
        return PasswordPane
      } else if (this.authnState.status === 'MFA_REQUIRED') {
        return MFAPane
      } else if (this.authnState.status === 'SUCCESS') {
        return LoadingSpinner
      }
      console.error('unknown status ' + this.authnState.status)
      return LoadingSpinner
    },

    currentTitle () {
      if (!this.authnState || this.authnState.status === 'IDP') {
        return this.$t('sign_in.title.continue')
      } else if (this.authnState.status === 'PRIMARY') {
        return this.$t('sign_in.title.continue')
      } else if (this.authnState.status === 'MFA_REQUIRED') {
        return this.$t('sign_in.title.two_step_verification')
      }
      return ''
    },

    currentDescription () {
      if (!this.authnState || this.authnState.status === 'IDP') {
        return this.$t('sign_in.description.continue')
      } else if (this.authnState.status === 'PRIMARY') {
        return this.$t('sign_in.description.enter_password')
      }
      return ''
    },

    shouldShowBackButton () {
      if (!this.authnState) return false
      if (this.authnState.status === 'IDP') return false
      if (this.authnState.status === 'SUCCESS') return false
      if (this.authnState.status === 'MFA_REQUIRED' && !this.selectedMFA) return false
      return true
    }
  },

  created () {
    this.RESET()
  },

  methods: {
    ...mapMutations('authn', [
      'RESET'
    ])
  },

  watch: {
    authnState (authnState) {
      if (!authnState) {
        return
      }
      if (authnState.status === 'SUCCESS') {
        this.logAnalytics('Authcore_loginSuccess', {}, true)
        const redirectURI = authnState.redirect_uri
        if (redirectURI) {
          // Redirection flow for desktop case
          const url = new URL(redirectURI)
          url.searchParams.set('code', authnState.authorization_code)
          url.searchParams.set('state', authnState.client_state)
          // Set redirectTo as timeout function to ensure showing loading spinner after page transition
          setTimeout(() => {
            redirectTo(url.toString(), this.containerId)
          }, 300)
        } else {
          // FIXME: legacy PostMessage flow for desktop/mobile case
        }
      }
    }
  }
}
</script>
