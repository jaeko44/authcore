<template>
  <confirmation-layout
    :title="$t('remove_authenticator_app.title')"
    :description="$t('remove_authenticator_app.description')"
    :button-text="$t('remove_authenticator_app.button.remove')"
    @click="submit"
    @back-button="$router.push({ name: 'MFAList' })"
  />
</template>

<script>
import { mapActions } from 'vuex'

import ConfirmationLayout from '@/components/ConfirmationLayout.vue'

export default {
  name: 'RemoveAuthenticatorApp',
  components: {
    ConfirmationLayout
  },
  props: {
    factorId: {
      type: Number,
      required: true
    }
  },

  methods: {
    ...mapActions('mfa', {
      removeAuthenticatorApp: 'delete'
    }),
    async submit () {
      await this.removeAuthenticatorApp(this.factorId)
      this.$router.push({ name: 'MFAList' })
    }
  }
}
</script>
