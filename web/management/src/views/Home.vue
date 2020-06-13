<template>
  <br/>
</template>

<script>
import { mapGetters } from 'vuex'

import { getRedirectPath, removeRedirectPath, verifyOAuthState } from '@/utils/util'

import queryString from 'query-string'

import router from '@/router'

export default {
  name: 'Home',

  computed: {
    ...mapGetters('authn', [
      'isAuthenticated'
    ])
  },

  async mounted () {
    // Generate access token from authorization code if no access token in local storage
    const qs = queryString.parse(location.search)
    if (qs.code) {
      if (verifyOAuthState(qs.state)) {
        await this.$store.dispatch('authn/createAccessToken', qs.code)
      } else {
        console.error('unknown OAuth state', qs.state)
        // It will fallthrough to redirect to login widget.
      }
    }
    if (this.isAuthenticated) {
      if (getRedirectPath()) {
        const redirectPath = getRedirectPath()
        removeRedirectPath()
        router.replace({ path: redirectPath })
      } else {
        router.replace({ name: 'Settings' })
      }
    } else {
      router.replace({ name: 'SignIn' })
    }
  }
}
</script>
