<template>
  <reset-password
    :company="company"
    :logo="logo"
    :contact-token="contactToken"
    :redirect-uri="redirectUri"
    :identifier="decodeURIComponent(identifier)"
    @authenticated="authenticated" />
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'

import ResetPassword from '@/components/ResetPassword.vue'

export default {
  name: 'FullResetPasswordView',
  components: {
    ResetPassword
  },

  props: {
    company: {
      type: String
    },
    logo: {
      type: String
    },
    contactToken: {
      type: String
    },
    redirectUri: {
      type: String
    },
    identifier: {
      type: String
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('client', ['authcoreClient']),
    ...mapState('widgets/account/resetPasswordForm', [
      'loading'
    ])
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {
    authenticated: async function () {
      await store.dispatch('widgets/account/get')
      this.postMessage('AuthCore_onSuccess', {
        current_user: this.user,
        access_token: this.authcoreClient.getAccessToken()
      })
    }
  }
}
</script>
