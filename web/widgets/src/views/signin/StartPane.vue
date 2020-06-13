<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <!-- Social login section in above the handle/password section case -->
      <div v-if="idpList && socialLoginPaneStyle === 'top'" id="social-login-section">
        <social-login-pane
          :register="false"
          :social-login-list="idpList"
          :social-login-pane-option="socialLoginPaneOption"
        />
        <b-row>
          <b-col class="my-4">
            <div class="center-separator">{{ separatorText }}</div>
          </b-col>
        </b-row>
      </div>
      <b-form @submit.prevent="onSubmit">
        <b-row>
          <b-col>
            <b-form-contact-input
              v-focus
              v-model="handle"
              :state="error ? false : null"
              aria-describedby="handle-error"
              :label="$t('sign_in.input.label.contact')"
              autocorrect="off"
              autocomplete="off"
              autocapitalize="off"
              :spellcheck="false"
              data-cy="handle"
            />
            <!-- Hidden password field to allow password manager to fill password -->
            <input
              v-if="password !== undefined"
              v-model="password"
              class="position-absolute layout-hidden"
              name="hiddenPassword"
              type="password"
              autocomplete="current-password"
              tabindex="-1"
            />
            <b-form-invalid-feedback id="handle-error">
              <i18n v-if="isInputHandleNotFound" :path="handleErrorNotFoundKey" tag="span">
                <template #link>
                  <b-link class="text-danger font-weight-bold" :to="to">{{ $t('sign_in.input.error.create_account') }}</b-link>
                </template>
              </i18n>
              <span v-else>{{ error || $t('general.blank') }}</span>
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col class="text-center">
            <with-loading-button
              block
              :class="{ 'w-75': buttonSize === 'normal' }"
              class="d-inline-block"
              type="submit"
              variant="primary"
              :loading="loading"
            >
              {{ $t('sign_in.button.sign_in') }}
            </with-loading-button>
          </b-col>
        </b-row>
        <b-row v-if="linkEnabled">
          <b-col class="text-center">
            <router-link :to="to"
                         class="font-weight-bold"
                         data-cy="register-link"
            >
              {{ $t('sign_in.link.register') }}
            </router-link>
          </b-col>
        </b-row>
      </b-form>
      <!-- Social login section in bottom the handle/password section case -->
      <div v-if="idpList && socialLoginPaneStyle === 'bottom'" id="social-login-section">
        <b-row>
          <b-col class="my-4">
            <div class="center-separator">{{ separatorText }}</div>
          </b-col>
        </b-row>
        <social-login-pane
          :register="false"
          :social-login-list="idpList"
          :social-login-pane-option="socialLoginPaneOption"
        />
      </div>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import WithLoadingButton from '@/components/WithLoadingButton.vue'
import SocialLoginPane from '@/components/SocialLoginPane.vue'

import { INPUT_HANDLE_NOT_FOUND_ERROR } from '@/store/types'

export default {
  name: 'StartPane',
  components: {
    WithLoadingButton,
    SocialLoginPane
  },

  data () {
    return {
      mergedQuery: {}
    }
  },

  computed: {
    ...mapState('client', [
      'widgetsSettings'
    ]),
    ...mapState('preferences', [
      'idpList',
      'redirectURI',
      'buttonSize',
      'socialLoginPaneStyle',
      'socialLoginPaneOption'
    ]),
    ...mapState('authn', [
      'authnState',
      'error',
      'loading'
    ]),
    separatorText () {
      return this.$t('register.text.or')
    },
    privacyLinkText () {
      return this.$t('register.link.privacy_link')
    },
    linkEnabled () {
      return this.widgetsSettings.signUpEnabled
    },
    isInputHandleNotFound () {
      return this.error === INPUT_HANDLE_NOT_FOUND_ERROR
    },
    to () {
      return {
        name: 'SignUp',
        query: this.mergedQuery
      }
    },
    handleErrorNotFoundKey () {
      if (this.widgetsSettings.signUpEnabled) {
        return 'sign_in.input.error.user_not_found_allow_create_account'
      }
      return 'sign_in.input.error.user_not_found_not_allow_create_account'
    },
    handle: {
      get () {
        return this.$store.state.authn.handle
      },
      set (value) {
        this.$store.commit('authn/SET_HANDLE', value)
      }
    },
    password: {
      get () {
        return this.$store.state.authn.password
      },
      set (value) {
        this.$store.commit('authn/SET_PASSWORD', value)
      }
    }
  },

  created () {
    this.handle = this.$route.query.handle
    this.$route.query.handle = ''
    this.mergedQuery = this.$route.query
  },

  methods: {
    ...mapActions('authn', {
      startAuthn: 'start'
    }),

    onSubmit () {
      this.logAnalytics('Authcore_loginStarted', { method: 'password' })
      const redirectURI = this.redirectURI
      const query = this.$route.query
      const codeChallenge = query.codeChallenge
      const codeChallengeMethod = query.codeChallengeMethod
      const clientState = query.clientState
      this.mergedQuery.handle = this.handle
      this.startAuthn({
        redirectURI,
        codeChallenge,
        codeChallengeMethod,
        clientState
      })
    }
  }
}
</script>

<style scoped lang="scss">
.layout-hidden {
    width: 0;
    height: 0;
    border: 0;
    padding: 1px;
    left: 50px;
    bottom: 50px;
    z-index: -1;
}
</style>
