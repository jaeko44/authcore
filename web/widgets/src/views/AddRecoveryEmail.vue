<template>
  <widget-layout-v2
    large-header
    :back-button-enabled="false"
    :title="$t('add_recovery_email.title')"
    :description="$t('add_recovery_email.description')"
  >
    <b-row class="mt-4">
      <b-col>
        <b-form @submit.prevent="submitRecoveryEmail">
          <b-row>
            <b-col>
              <b-bsq-input
                v-focus
                v-model="recoveryEmail"
                :state="recoveryEmailError ? false : null"
                aria-describedby="recovery-email-error"
                :label="$t('add_recovery_email.input.label.recovery_email')"
                type="text"
              />
              <b-form-invalid-feedback id="recovery-email-error">
                {{ recoveryEmailError || $t('general.blank') }}
              </b-form-invalid-feedback>
            </b-col>
          </b-row>
          <b-row class="mb-4">
            <b-col class="text-center">
              <with-loading-button
                block
                type="submit"
              >
                {{ $t('add_recovery_email.button.next') }}
              </with-loading-button>
            </b-col>
          </b-row>
        </b-form>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import { mapState } from 'vuex'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import WithLoadingButton from '@/components/WithLoadingButton.vue'

export default {
  name: 'AddRecoveryEmail',
  components: {
    WidgetLayoutV2,
    WithLoadingButton
  },

  computed: {
    ...mapState('authn', [
      'recoveryEmailError'
    ]),
    recoveryEmail: {
      get () {
        return this.$store.state.authn.recoveryEmail
      },
      set (value) {
        this.$store.commit('authn/SET_RECOVERY_EMAIL', value)
      }
    }
  }
}
</script>
