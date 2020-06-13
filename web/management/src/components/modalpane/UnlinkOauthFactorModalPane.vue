<template>
  <confirmation-modal-pane
    v-bind:visible="visible"
    @update:visible="$emit('update:visible', $event)"
    :header-title="modalHeaderTitle"
    :button-title="modalButtonTitle"
    button-variant="danger"
    :description="modalDescription"
    @ok="modalSubmit"
  />
</template>

<script>
import { mapActions } from 'vuex'

import ConfirmationModalPane from '@/components/modalpane/ConfirmationModalPane.vue'

export default {
  name: 'UnlinkOauthFactorModalPane',
  components: {
    ConfirmationModalPane
  },
  model: {
    prop: 'visible',
    event: 'change'
  },
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    user: {
      type: Object,
      required: true
    },
    oauthFactor: {
      type: Object,
      required: true
    }
  },

  computed: {
    modalHeaderTitle () {
      if (this.oauthFactor.service) {
        return this.$t('unlink_oauth_factor_modal_pane.title', { service: this.$t(`model.oauth_factor.service.${this.serviceName}`) })
      }
      return ''
    },
    modalButtonTitle () {
      if (this.oauthFactor.service) {
        return this.$t('unlink_oauth_factor_modal_pane.button.unlink_oauth_factor', { service: this.$t(`model.oauth_factor.service.${this.serviceName}`) })
      }
      return ''
    },
    modalDescription () {
      if (this.oauthFactor.service) {
        return this.$t('unlink_oauth_factor_modal_pane.description', { service: this.$t(`model.oauth_factor.service.${this.serviceName}`) })
      }
      return ''
    },
    serviceName () {
      return this.oauthFactor.service.toLowerCase()
    }
  },

  methods: {
    ...mapActions('management/modalPane/unlinkOauthFactor', [
      'unlink'
    ]),
    async modalSubmit () {
      await this.unlink(this.serviceName)
      // Dismiss the modal when the action finished
      this.$emit('update:visible', false)
      this.$emit('updated', true)
    }
  }
}
</script>
