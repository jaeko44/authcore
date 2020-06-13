<template>
  <b-row>
    <b-col cols="12" v-if="selectedMFA">
      <b-form @submit.prevent="verifyMFA(code)">
        <b-row class="mb-4" align-h="center">
          <b-col class="text-center">
            {{ description }}
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <b-bsq-input
              v-focus
              v-model="code"
              class="hide-spin-button"
              :label="label"
              :state="error ? false : null"
              aria-describedby="mfa-error"
              autocomplete="off"
              type="number"
            />
            <b-form-invalid-feedback
              id="mfa-error"
              class="d-inline-block w-50"
            >
              {{ error || $t('general.blank') }}
            </b-form-invalid-feedback>
            <div v-if="selectedMFA === 'sms_otp'" class="d-inline-flex w-50 justify-content-end">
              <b-link
                v-if="!resendDone"
                @click="resendAuthenticationCode"
                class="text-right"
              >
                {{ $t('sign_in.link.resend_verification_code') }}
              </b-link>
              <span
                v-else
                class="text-grey-medium"
              >
                {{ $t('sign_in.text.code_sent') }}
              </span>
            </div>
          </b-col>
        </b-row>
        <b-row class="mb-3">
          <b-col class="text-center">
            <b-button
              block
              :class="{ 'w-75': buttonSize === 'normal' }"
              class="d-inline-block"
              type="submit"
              variant="primary"
            >
              {{ $t('sign_in.button.next') }}
            </b-button>
          </b-col>
        </b-row>
        <b-row>
          <b-col cols="12" class="text-center">
            <b-link class="font-weight-bold" @click="showTryAnotherWay">{{ $t('sign_in.link.try_another_way') }}</b-link>
          </b-col>
        </b-row>
      </b-form>
    </b-col>
    <b-col
      v-else
      class="px-0"
    >
      <b-list-group>
        <hyperlink-list-item
          v-if="availableFactors.includes('sms_otp')"
          :title="$t('sign_in.list_item.title.sms_code')"
          @click="SET_SELECTED_MFA('sms_otp')"
        >
          <div class="text-grey-dark">
            {{ $t('sign_in.list_item.text.sms_code') }}
          </div>
        </hyperlink-list-item>
        <hyperlink-list-item
          v-if="availableFactors.includes('totp')"
          :title="$t('sign_in.list_item.title.authenticator_app')"
          @click="SET_SELECTED_MFA('totp')"
        >
          <div class="text-grey-dark">
            {{ $t('sign_in.list_item.text.authenticator_app') }}
          </div>
        </hyperlink-list-item>
        <hyperlink-list-item
          v-if="availableFactors.includes('back_upcode')"
          :title="$t('sign_in.list_item.title.backup_code')"
          @click="SET_SELECTED_MFA('backup_code')"
        >
          <div class="text-grey-dark">
            {{ $t('sign_in.list_item.text.backup_code') }}
          </div>
        </hyperlink-list-item>
      </b-list-group>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions, mapMutations } from 'vuex'

import HyperlinkListItem from '@/components/HyperlinkListItem.vue'

const MFA_FACTOR_PRIORITY = ['totp', 'sms_otp', 'backup_code']

export default {
  name: 'PasswordPane',

  components: {
    HyperlinkListItem
  },

  data () {
    return {
      code: ''
    }
  },

  computed: {
    ...mapState('preferences', [
      'buttonSize',
      'socialLoginPaneStyle',
      'socialLoginPaneOption'
    ]),
    ...mapState('authn', [
      'authnState',
      'error',
      'loading',
      'selectedMFA'
    ]),
    description () {
      switch (this.selectedMFA) {
        case 'totp':
          return this.$t('sign_in.list_item.text.authenticator_app')
        case 'sms_otp':
          return this.$t('sign_in.list_item.text.sms_code')
        case 'backup_code':
          return this.$t('sign_in.list_item.text.backup_code')
      }
      return ''
    },
    label () {
      switch (this.selectedMFA) {
        case 'totp':
          return this.$t('sign_in.input.label.authenticator_app')
        case 'sms_otp':
          return this.$t('sign_in.input.label.sms_code')
        case 'backup_code':
          return this.$t('sign_in.input.label.backup_code')
      }
      return ''
    },
    availableFactors () {
      if (this.authnState) {
        return this.authnState.factors
      }
      return []
    }
  },

  mounted () {
    // Choose a MFA method
    for (let i = 0; i < MFA_FACTOR_PRIORITY.length; i++) {
      const factor = MFA_FACTOR_PRIORITY[i]
      if (this.authnState.factors.includes(factor)) {
        this.SET_SELECTED_MFA(factor)
        break
      }
    }
  },

  methods: {
    ...mapActions('authn', [
      'verifyMFA'
    ]),

    ...mapMutations('authn', [
      'SET_SELECTED_MFA',
      'UNSET_SELECTED_MFA'
    ]),

    showTryAnotherWay () {
      this.UNSET_SELECTED_MFA()
    }
  }
}
</script>
