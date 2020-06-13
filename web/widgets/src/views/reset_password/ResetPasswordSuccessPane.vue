<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <p class="text-center lead">{{$t('reset_password.text.reset_password_success')}}</p>
      <b-button block :href="signInURL" variant="primary">{{ $t('reset_password.text.return_home') }}</b-button>
    </b-col>
  </b-row>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'ResetPasswordSuccess',

  computed: {
    ...mapState('client', [
      'widgetsSettings'
    ]),

    signInURL () {
      return this.widgetsSettings.redirectFallbackUrl
    }
  }
}
</script>
