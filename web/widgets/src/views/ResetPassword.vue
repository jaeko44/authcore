<template>
  <widget-layout-v2
    large-header
    :back-button-enabled="true"
    :title="currentTitle"
    :description="currentDescription"
    @back-button="$router.push({name: 'SignIn'})"
    >
    <component :is="currentPane" />
  </widget-layout-v2>

</template>

<script>
import { mapState, mapActions } from 'vuex'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import ResetPasswordPane from '@/views/reset_password/ResetPasswordPane.vue'
import ResetLinkSentPane from '@/views/reset_password/ResetLinkSentPane.vue'
import NewPasswordPane from '@/views/reset_password/NewPasswordPane.vue'
import ResetPasswordSuccessPane from '@/views/reset_password/ResetPasswordSuccessPane.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'ResetPasswordView',
  components: {
    WidgetLayoutV2
  },

  computed: {
    ...mapState('preferences', [
      'buttonSize'
    ]),
    ...mapState('authn', [
      'authnState',
      'resetToken',
      'error',
      'loading'
    ]),
    currentPane () {
      if (this.loading) {
        return LoadingSpinner
      } else if (!this.authnState) {
        return ResetPasswordPane
      } else if (this.authnState.status === 'PASSWORD_RESET' && !this.resetToken) {
        return ResetLinkSentPane
      } else if (this.authnState.status === 'PASSWORD_RESET') {
        return NewPasswordPane
      } else if (this.authnState.status === 'PASSWORD_RESET_SUCCESS') {
        return ResetPasswordSuccessPane
      }
      console.error('unknown status ' + this.authnState.status)
      return LoadingSpinner
    },

    currentTitle () {
      if (!this.authnState) {
        return this.$t('reset_password.title.reset_password')
      } else if (this.authnState.status === 'PASSWORD_RESET') {
        return this.$t('reset_password.title.set_new_password')
      } else if (this.authnState.status === 'PASSWORD_RESET_SUCCESS') {
        return this.$t('reset_password.title.password_changed')
      }
      return ''
    },

    currentDescription () {
      return ''
    }
  },

  methods: {
    ...mapActions('authn', [
      'verifyPasswordReset'
    ])
  },

  mounted () {
    this.$store.commit('authn/RESET')

    if (this.$route.query.handle) {
      this.$store.commit('authn/SET_HANDLE', this.$route.query.handle)
    }

    if (this.$route.query.stateToken && this.$route.query.resetToken) {
      const stateToken = this.$route.query.stateToken
      const resetToken = this.$route.query.resetToken
      this.verifyPasswordReset({ stateToken, resetToken })
    }
  }
}
</script>
