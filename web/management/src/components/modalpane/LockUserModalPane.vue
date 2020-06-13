<template>
  <confirmation-modal-pane
    v-bind:visible="visible"
    @update:visible="$emit('update:visible', $event)"
    :header-title="modalHeaderTitle"
    :button-title="modalButtonTitle"
    button-variant="primary"
    :description="modalDescription"
    @ok="modalSubmit"
  />
</template>

<script>
import { mapActions } from 'vuex'
import ConfirmationModalPane from '@/components/modalpane/ConfirmationModalPane.vue'

export default {
  name: 'LockUserModalPane',
  components: {
    ConfirmationModalPane
  },
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    user: {
      type: Object,
      default: () => {}
    }
  },

  computed: {
    modalHeaderTitle () {
      if (this.user.locked) {
        return this.$t('lock_user_modal_pane.title.unlock_user')
      }
      return this.$t('lock_user_modal_pane.title.lock_user')
    },
    modalButtonTitle () {
      if (this.user.locked) {
        return this.$t('lock_user_modal_pane.button.unlock_user')
      }
      return this.$t('lock_user_modal_pane.button.lock_user')
    },
    modalDescription () {
      if (this.user.locked) {
        return this.$t('lock_user_modal_pane.description.unlock_user')
      }
      return this.$t('lock_user_modal_pane.description.lock_user')
    }
  },

  methods: {
    ...mapActions('management/modalPane/lockUser', [
      'lock',
      'unlock'
    ]),
    modalSubmit () {
      if (this.user.locked) {
        this.unlock()
      } else {
        this.lock()
      }
      // Dismiss the modal when the action finished
      this.$emit('update:visible', false)
    }
  }
}
</script>
