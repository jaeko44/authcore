<template>
  <b-alert
    v-model="showAlert"
    class="text-center rounded-0"
    :variant="alertType"
    dismissible
  >
    {{ $t(alertMessage) }}
  </b-alert>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

export default {
  name: 'AlertPane',

  computed: {
    ...mapState('alert', [
      'alert'
    ]),
    ...mapGetters('alert', [
      'shownAlert'
    ]),
    alertType () {
      return this.alert.type ? this.alert.type : ''
    },
    alertMessage () {
      return this.alert.message ? this.alert.message : ''
    },
    showAlert: {
      get () {
        return this.shownAlert
      },
      set () {
        this.$store.commit('alert/UNSET_MESSAGE')
      }
    }
  }
}
</script>
