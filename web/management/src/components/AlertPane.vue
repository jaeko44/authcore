<template>
  <b-container>
    <b-row>
      <b-col class="mt-1">
        <b-alert
          v-model="showAlert"
          class="text-center rounded-0"
          :variant="alertType"
          dismissible
        >
          {{ $t(alertMessage) }}
        </b-alert>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

export default {
  name: 'AlertPane',

  computed: {
    ...mapState('management/alert', [
      'alert'
    ]),
    ...mapGetters('management/alert', [
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
        this.$store.commit('management/alert/UNSET_MESSAGE')
      }
    }
  }
}
</script>
