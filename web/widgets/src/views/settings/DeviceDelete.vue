<template>
  <confirmation-layout
    back-enabled
    :title="$t('device_delete.title')"
    :description="deviceDescription"
    :button-text="$t('device_delete.button.log_out')"
    @click="deleteAction"
    @back-button="$router.push({ name: 'Devices' })">
  </confirmation-layout>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import successCallbackMixin from '@/mixins/successCallback'

import ConfirmationLayout from '@/components/ConfirmationLayout.vue'

export default {
  name: 'DeviceDelete',
  mixins: [successCallbackMixin],
  components: {
    ConfirmationLayout
  },

  props: {
    deviceId: {
      // Allow it to be 0 represents to log out all other devices
      type: Number,
      required: true
    },
    callbackAction: {
      default: function () {
        const action = this.deviceId !== 0 ? 'device_logout' : 'other_devices_logout'
        return action
      }
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('devices', [
      'loading',
      'done',
      'error',

      'sessions'
    ]),
    device () {
      return this.sessions.find(item => item.id === this.deviceId)
    },
    deviceDescription () {
      if (this.deviceId === 0) {
        return this.$t('device_delete.text.log_out_all_other_devices')
      }
      return this.$t('device_delete.text.log_out_the_following_device')
    }
  },

  watch: {},

  destroyed () {
    this.$store.commit('devices/RESET')
  },

  methods: {
    ...mapActions('devices', [
      'delete',
      'deleteAll'
    ]),
    async deleteAction () {
      if (this.deviceId === 0) {
        await this.deleteAll()
      } else {
        await this.delete(this.deviceId)
      }
      this.$router.push({ name: 'Devices' })
    }
  }
}
</script>
