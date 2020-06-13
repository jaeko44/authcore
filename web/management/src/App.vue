<template>
  <router-view />
</template>

<script>
import client from '@/client'

export default {
  name: 'App',

  async created () {
    const accessToken = client.tokenManager.get('access_token')

    // For legacy code
    if (accessToken) {
      this.$store.dispatch('currentUser/get')
      this.$store.commit('currentUser/SET_AUTHENTICATED', true)
      this.$store.commit('management/SET_READY_STATE', true)
    }

    // Insert global error handler to handle 401 (token expired) response from API server.
    client.client.errorHandler = (e) => {
      if (e.response && e.response.status === 401) {
        console.error('Request failed with status code 401')
        client.tokenManager.clear()
        // Redirect to sign in if current client is the user portal.
        console.warn('Redirecting to sign in')
        const url = new URL('/web/', window.origin)
        window.location.replace(url)
      } else {
        console.error('Authcore client: ', e)
      }
      return Promise.reject(e)
    }
  }
}
</script>
