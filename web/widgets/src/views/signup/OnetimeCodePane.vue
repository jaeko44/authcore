<template>
  <b-row>
    <b-col cols="12">
      <b-form @submit.prevent="verifyOnetimeCode">
        <b-row class="my-4" align-h="center">
          <b-col class="h5 my-0 text-grey text-center" cols="12">
            {{ handle }}
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <b-bsq-input
              v-focus
              v-model="onetimeCode"
              :state="error ? false : null"
              aria-describedby="code-error"
              :label="$t('register.input.label.one_time_code')"
              type="number"
            />
            <b-form-invalid-feedback id="code-error">
              {{ error || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col class="text-center">
            <with-loading-button
              block
              type="submit"
              :button-size="buttonSize"
              :loading="loading"
            >
              {{ $t('register.button.next') }}
            </with-loading-button>
          </b-col>
        </b-row>
      </b-form>
      <b-row>
        <b-col class="text-center">
          <b-link
            class="font-weight-bold"
            @click="resendCode"
          >
            {{ $t('register.link.resend_code') }}
          </b-link>
        </b-col>
      </b-row>
    </b-col>
  </b-row>
</template>

<script>
import { isPossibleNumber } from 'libphonenumber-js'
import { mapState, mapActions } from 'vuex'

import WithLoadingButton from '@/components/WithLoadingButton.vue'

export default {
  name: 'OnetimeCodePane',
  components: {
    WithLoadingButton
  },
  props: {},

  data () {
    return {
      loading: false
    }
  },

  computed: {
    ...mapState('preferences', [
      'buttonSize'
    ]),
    ...mapState('authn', [
      'error',
      'handle'
    ]),
    onetimeCode: {
      get () {
        return this.$store.state.authn.onetimeCode
      },
      set (value) {
        this.$store.commit('authn/SET_ONETIME_CODE', value)
      }
    }
  },

  methods: {
    ...mapActions('authn', [
      'signUp'
    ]),
    resendCode () {
      console.log('Resend one time code')
    },
    verifyOnetimeCode () {
      console.log('Verify one time code')
      this.loading = !this.loading
      this.$store.dispatch('authn/verifySignUp')
      try {
        let contactType = 'email'
        if (isPossibleNumber(this.handle)) {
          contactType = 'phone'
        }
        this.logAnalytics('Authcore_registerStarted', { contactType })
        this.signUp({
          redirectURI: this.redirectURI,
          privacyCheckbox: this.privacyCheckbox
        })
      } catch (err) {
        console.error(err)
      }
    }
  }
}
</script>
