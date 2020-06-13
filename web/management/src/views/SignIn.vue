<template>
  <div class="pt-5">
    <loading-spinner />
  </div>
</template>

<script>
import client from '@/client'
import { createOAuthState } from '@/utils/util'

import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'SignIn',
  components: {
    LoadingSpinner
  },

  mounted () {
    const redirectUrl = new URL('/web/', document.location)
    const state = createOAuthState()
    const url = client.client.authCodeURL(state, redirectUrl.toString())
    window.location.replace(url)
  }
}
</script>
