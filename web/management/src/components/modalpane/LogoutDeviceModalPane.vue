<template>
  <b-bsq-modal
    v-bind:visible="visible"
    v-on:change="$emit('update:visible', $event)"
    :header-title="$t('logout_device_modal_pane.title')"
    :button-title="$t('logout_device_modal_pane.button.log_out')"
    button-variant="danger"
    @ok="modalSubmit"
  >
    <div class="mt-3 mb-5">
      <div class="mb-3">
        {{ $t('logout_device_modal_pane.description') }}
      </div>
      <i18n path="model.device.last_seen_at_with_date" tag="div" class="mb-3">
        <span place="date">{{ lastSeenAt }}</span>
      </i18n>
      <div>
        {{ userAgent }}
      </div>
    </div>
  </b-bsq-modal>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'LogoutDeviceModalPane',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    logoutSession: {
      type: Object,
      default: () => {}
    }
  },

  computed: {
    ...mapState('management/modalPane/lockUser', [
      'loading'
    ]),
    lastSeenAt () {
      if (this.logoutSession) {
        return this.logoutSession.lastSeenAt
      }
      return ''
    },
    userAgent () {
      if (this.logoutSession) {
        return this.logoutSession.userAgent
      }
      return ''
    }
  },

  methods: {
    ...mapActions('management/userDetails/devicesList', {
      logoutDevice: 'delete'
    }),
    async modalSubmit () {
      await this.logoutDevice(this.logoutSession.id)
      this.$emit('update:visible', false)
    }
  }
}
</script>
