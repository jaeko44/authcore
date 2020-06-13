<template>
  <confirmation-template
    back-enabled
    :title="$t('resend_verification.title')"
    button-variant="primary"
    :action="resendVerificationCode"
  >
    <b-row>
      <b-col class="text-center">
        <i18n path="resend_verification.text.resend_contact_verification" tag="div">
          <template #contact>
            <div class="h5 mt-4 text-grey-dark">
              {{ contactValue }}
            </div>
          </template>
        </i18n>
      </b-col>
    </b-row>
  </confirmation-template>
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'
import router from '@/router'

import ConfirmationTemplate from '@/components/ConfirmationTemplate.vue'

export default {
  name: 'ResendVerification',
  components: {
    ConfirmationTemplate
  },

  props: {
    type: {
      type: String,
      validator: (value) => {
        return ['email', 'phone', 'factor'].indexOf(value) !== -1
      }
    },
    value: {
      type: String,
      default: ''
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('widgets/account/verifyAccountForm', [
      'contact'
    ]),
    contactValue () {
      switch (this.type) {
        case 'email':
        case 'phone': {
          if (this.contact !== undefined) {
            return this.contact.value
          }
          break
        }
        case 'factor': {
          return this.value
        }
      }
      return ''
    }
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {
    resendVerificationCode () {
      switch (this.type) {
        case 'email':
        case 'phone': {
          store.dispatch('widgets/account/verifyAccountForm/resendPrimaryContactVerification', this.type)
          break
        }
        case 'factor': {
          store.dispatch('widgets/mfaAuthenticatorsList/createSmsAuthenticatorForm/init', this.contactValue)
          break
        }
      }
      router.go(-1)
    }
  }
}
</script>
