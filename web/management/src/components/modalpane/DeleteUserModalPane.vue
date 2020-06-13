<template>
  <confirmation-modal-pane
    v-bind:visible="visible"
    @update:visible="$emit('update:visible', $event)"
    :header-title="$t('delete_user_modal_pane.title')"
    :button-title="$t('delete_user_modal_pane.button.delete_user')"
    button-variant="danger"
    :description="$t('delete_user_modal_pane.description')"
    @ok="modalSubmit"
  />
</template>

<script>
import { mapActions } from 'vuex'
import ConfirmationModalPane from '@/components/modalpane/ConfirmationModalPane.vue'

export default {
  name: 'DeleteUserModalPane',
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

  methods: {
    ...mapActions('management/modalPane/deleteUser', {
      deleteUser: 'delete'
    }),
    async modalSubmit () {
      await this.deleteUser(this.user.id)
      // Dismiss the modal when the action finished
      this.$emit('update:visible', false)
      this.$emit('updated', true)
    }
  }
}
</script>
