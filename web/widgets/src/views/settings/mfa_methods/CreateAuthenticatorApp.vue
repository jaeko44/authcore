<template>
  <widget-layout-v2
    :title="$t('mfa_totp_create.title')"
    @back-button="$router.push({ name: 'MFAList' })"
  >
    <template #description>
      <span v-if="inputPageStatus.showQRcode">
        {{ $t('mfa_totp_create.description.scan_qrcode') }}
      </span>
      <span v-else>
        <b-row>
          <b-col>
            <div>
              {{ $t('mfa_totp_create.description.copy_key') }}
            </div>
            <i18n path="mfa_totp_create.description.time_based_key" tag="div" for="time_based">
              <span class="font-weight-bold">{{ $t('mfa_totp_create.description.time_based') }}</span>
            </i18n>
          </b-col>
        </b-row>
      </span>
    </template>
    <component
      :is="currentPane"
      :key="currentPaneKey"
      v-bind.sync="currentSyncProps"
      @create-success="done = true"
    />
  </widget-layout-v2>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import successCallbackMixin from '@/mixins/successCallback'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import InputPane from '@/views/settings/mfa_methods/authenticator_app/InputPane.vue'
import SuccessfulShowUp from '@/components/SuccessfulShowUp.vue'

export default {
  name: 'CreateAuthenticatorApp',
  mixins: [successCallbackMixin],

  components: {
    WidgetLayoutV2,
    InputPane,
    SuccessfulShowUp
  },

  props: {
    callbackAction: {
      default: 'totp_factor_created'
    }
  },

  data () {
    return {
      inputPageStatus: {
        showQRcode: true
      },

      done: false
    }
  },

  computed: {
    currentPane () {
      if (this.done) {
        return SuccessfulShowUp
      }
      return InputPane
    },
    currentPaneKey () {
      if (this.done) {
        return 'success'
      }
      return 'input'
    },
    currentSyncProps () {
      if (!this.done) {
        return this.inputPageStatus
      }
      return {}
    },
    ...mapState('mfa', [
      'loading'
    ])
  },

  watch: {
    done (newVal) {
      if (newVal) {
        setTimeout(() => {
          this.$router.push({ name: 'MFAList' })
        }, 3000)
      }
    }
  },

  async mounted () {
    this.getCurrentUser()
    this.generateTOTPSecret()
  },

  destroyed () {
    this.$store.commit('mfa/RESET')
  },

  methods: {
    ...mapActions('mfa', [
      'generateTOTPSecret'
    ]),
    ...mapActions('users', [
      'getCurrentUser'
    ])
  }
}
</script>
