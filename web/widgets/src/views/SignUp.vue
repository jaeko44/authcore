<template>
  <widget-layout-v2
    large-header
    :back-button-enabled="shouldShowBackButton"
    :title="currentTitle"
    :description="currentDescription"
    @back-button="RESET()"
  >
    <component :is="currentPane" />
  </widget-layout-v2>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

import { redirectTo } from '@/utils/util'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import StartPane from '@/views/signup/StartPane.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'SignUp',
  components: {
    WidgetLayoutV2,
    LoadingSpinner
  },

  computed: {
    ...mapState('preferences', [
      'redirectURI'
    ]),
    ...mapState('authn', [
      'authnState',
      'error',
      'loading',
      'signUpErrors'
    ]),

    currentPane () {
      if (!this.authnState || this.authnState.status === 'IDP') {
        return StartPane
      }
      return LoadingSpinner
    },

    currentTitle () {
      if (this.authnState && this.authnState.status === 'SUCCESS') {
        return ''
      }
      return this.$t('register.title')
    },

    currentDescription () {
      if (this.authnState && this.authnState.status === 'SUCCESS') {
        return ''
      }
      return this.$t('register.description.start')
    },

    shouldShowBackButton () {
      if (!this.authnState) return false
      else if (this.authnState.status === 'IDP') return false
      else if (this.authnState.status === 'SUCCESS') return false
      return true
    },

    privacyChecked: {
      get () {
        return this.$store.state.authn.privacyChecked
      },
      set (value) {
        this.$store.commit('authn/SET_PRIVACY_CHECKED', value)
      }
    },

    shouldShowLoadingSpinner () {
      return this.loading || (this.authnState && this.authnState.status === 'SUCCESS')
    }
  },

  created () {
    this.RESET()
  },

  watch: {
    authnState (authnState) {
      if (!authnState) {
        return
      }
      if (authnState.status === 'SUCCESS') {
        this.logAnalytics('Authcore_registerSuccess', {}, true)
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
  },

  methods: {
    ...mapMutations('authn', [
      'RESET'
    ])
  }
}
</script>
