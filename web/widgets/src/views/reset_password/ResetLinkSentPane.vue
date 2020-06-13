<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <p class="text-center lead">{{$t('reset_password.text.reset_link_sent', { handle })}}</p>
    </b-col>
  </b-row>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'ResetLinkSent',

  computed: {
    ...mapState('authn', [
      'handle'
    ])
  }
}
</script>
