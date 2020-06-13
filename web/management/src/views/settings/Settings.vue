<template>
  <div class="pt-5">
    <loading-spinner />
  </div>
</template>

<script>
import LoadingSpinner from '@/components/LoadingSpinner.vue'

import client from '@/client'

export default {
  name: 'Settings',
  components: {
    LoadingSpinner
  },

  mounted () {
    const params = new URLSearchParams()
    params.append('clientId', client.clientId)
    const url = new URL('../widgets/settings?' + params.toString(), client.baseURL).toString()
    window.location.replace(url)
  },

  methods: {}
}
</script>
