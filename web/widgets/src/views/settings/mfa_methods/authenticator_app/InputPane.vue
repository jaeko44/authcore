<template>
  <b-form @submit.prevent="createTOTP">
    <b-row class="border-bottom" align-h="center">
      <b-col>
        <b-row>
          <b-col class="text-center">
            <div class="d-flex">
              <b-row
                no-gutters
                align-h="center"
                class="w-100">
                <b-col>
                  <qrcode-view
                    v-if="showQRcode && totpURL !== ''"
                    class="rounded-lg"
                    :content="totpURL"
                  />
                  <div v-else class="">
                    <div>
                      <b-row>
                        <b-col>
                          <textarea
                            id="secret"
                            name="secret"
                            type="text"
                            :value="totpSecretFormatted"
                            class="mt-3 text-center text-break border-0"
                            style="width: 95%; resize: none;"
                            readonly
                          />
                          <b-button
                            variant="link"
                            @click="copyToClipboard"
                          >
                            <span v-if="copiedTextcode">{{ $t('mfa_totp_create.button.copied') }}</span>
                            <span v-else>{{ $t('mfa_totp_create.button.copy') }}</span>
                          </b-button>
                        </b-col>
                      </b-row>
                    </div>
                  </div>
                </b-col>
              </b-row>
            </div>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col class="text-center">
            {{ $t('mfa_totp_create.text.or') }}
            <b-link
              @click="$emit('update:showQRcode', !showQRcode)"
            >
              {{ $t(linkTitle) }}
            </b-link>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
    <b-row align-h="center" class="mt-4">
      <b-col>
        <b-row>
          <b-col class="mb-3">
            <i18n path="mfa_totp_create.text.enter_authenticator_app_code" tag="span">
              <span class="font-weight-bold"><i18n path="description.n_digits" tag="span">6</i18n></span>
            </i18n>
          </b-col>
        </b-row>
        <b-row align-h="center">
          <b-col>
            <b-bsq-input
              class="hide-spin-button"
              v-model="totpCode"
              :state="inputCodeError ? false : null"
              :disabled="loading"
              aria-describedby="pinInvalidFeedback"
              :label="$t('mfa_totp_create.input.label.code')"
              autocomplete="off"
              type="number"
            />
            <b-form-invalid-feedback id="pinInvalidFeedback">
              {{ inputCodeError || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row align-h="center">
          <b-col>
            <b-row>
              <b-col>
                <b-button
                  variant="primary"
                  type="submit"
                  block
                >
                  {{ $t('mfa_totp_create.button.next') }}
                </b-button>
              </b-col>
            </b-row>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
  </b-form>
</template>

<script>
import { mapState } from 'vuex'

import QrcodeView from '@/components/QrcodeView.vue'

export default {
  name: 'InputPane',
  components: {
    QrcodeView
  },
  props: {
    showQRcode: {
      type: Boolean,
      default: false
    }
  },

  data () {
    return {
      copiedTextcode: false
    }
  },

  computed: {
    ...mapState('preferences', [
      'company'
    ]),
    ...mapState('mfa', [
      'totpSecret',
      'loading',
      'error'
    ]),
    ...mapState('users', [
      'displayName'
    ]),
    totpCode: {
      get () {
        return this.$store.state.mfa.totpCode
      },
      set (value) {
        this.$store.commit('mfa/SET_TOTP_CODE', value)
      }
    },
    totpSecretFormatted () {
      return this.totpSecret.replace(/(.{4})/g, '$1 ').trim()
    },
    totpURL () {
      const escapedHandle = encodeURI(this.displayName)
      const { totpSecret, company } = this
      return `otpauth://totp/${escapedHandle}?secret=${totpSecret}&issuer=${company}`
    },
    linkTitle () {
      if (this.showQRcode) {
        return 'mfa_totp_create.link.input_manually'
      }
      return 'mfa_totp_create.link.scan_qrcode'
    },
    inputCodeError () {
      return this.error
    }
  },

  watch: {},

  methods: {
    async createTOTP () {
      const { totpPIN, totpSecret } = this
      await this.$store.dispatch('mfa/createTOTP', {
        identifier: '',
        totpCode: totpPIN,
        totpSecret: totpSecret
      })
      this.$emit('create-success')
    },
    copyToClipboard () {
      const secret = document.getElementById('secret')
      secret.select()
      document.execCommand('copy')
      this.copiedTextcode = true
      setTimeout(() => {
        this.copiedTextcode = false
      }, 3000)
    }
  }
}
</script>
