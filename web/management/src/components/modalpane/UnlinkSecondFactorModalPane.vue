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
  name: 'UnlinkSecondFactorModalPane',
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
    secondFactor: {
      type: Object,
      required: true
    }
  },

  computed: {
    modalHeaderTitle () {
      if (this.secondFactor.type) {
        return this.$t('unlink_second_factor_modal_pane.title', { type: this.type })
      }
      return ''
    },
    modalButtonTitle () {
      if (this.secondFactor.type) {
        return this.$t('unlink_second_factor_modal_pane.button.unlink_second_factor', { type: this.type })
      }
      return ''
    },
    modalDescription () {
      if (this.secondFactor.type) {
        return this.$t('unlink_second_factor_modal_pane.description', { type: this.type })
      }
      return ''
    },
    type () {
      return this.$t(`model.second_factor.${this.secondFactor.type}`)
    }
  },

  methods: {
    ...mapActions('management/modalPane/unlinkSecondFactor', [
      'unlink'
    ]),
    async modalSubmit () {
      console.log(this.secondFactor)
      await this.unlink(this.secondFactor.id)
      // Dismiss the modal when the action finished
      this.$emit('update:visible', false)
      this.$emit('updated', true)
    }
  }
}
</script>
