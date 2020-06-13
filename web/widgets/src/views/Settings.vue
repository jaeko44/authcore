<template>
  <router-view />
</template>

<script>
import { mapState, mapGetters } from 'vuex'

import client from '@/client'

export default {
  name: 'Setting',
  components: {},

  props: {},

  data () {
    return {}
  },

  computed: {
    ...mapState('preferences', [
      'clientId'
    ]),
    ...mapState('client', [
      'authcoreClient'
    ]),
    ...mapGetters('preferences', [
      'isAdminPortal'
    ]),
    ...mapGetters('client', [
      'accessToken'
    ]),
    ...mapGetters('authn', [
      'isAuthenticated'
    ])
  },

  watch: {
    // Preferences in store is changed when the route is changed to initialize the preferences.
    // See App.vue for the initialization
    async clientId (newVal) {
      // User portal
      if (newVal && this.isAdminPortal) {
        const accessToken = client.tokenManager.get('access_token')
        if (accessToken) {
          // For legacy code
          await this.authcoreClient.setAccessToken(accessToken)
        } else {
          const url = new URL('/web/', window.origin)
          window.location.replace(url)
        }
      }
    }
  },

  async mounted () {
    // Hash mode settings
    // As an entry point from app side, check if the hash exists and set it as access token.
    // Assuming hash is access token from client.
    if (location.hash !== '') {
      try {
        const accessTokenFromHash = location.hash.substring(1)
        await this.authcoreClient.setAccessToken(accessTokenFromHash)
        client.tokenManager.add('access_token', accessTokenFromHash)
        // eslint-disable-next-line require-atomic-updates
        location.hash = ''
      } catch (err) {
        console.error(err)
      }
    }

    // Insert global error handler to handle 401 (token expired) response from API server.
    client.client.errorHandler = (e) => {
      if (e.response && e.response.status === 401) {
        console.error('Request failed with status code 401')
        client.tokenManager.clear()
        if (this.isAdminPortal) {
          // Redirect to sign in if current client is the user portal.
          console.warn('Redirecting to sign in')
          const url = new URL('/web/', window.origin)
          window.location.replace(url)
        }
      } else {
        console.error('Authcore client: ', e)
      }
      return Promise.reject(e)
    }
  }
}
</script>
