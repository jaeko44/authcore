<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <!-- Social login section in above the handle/password section case -->
      <div v-if="idpList && socialLoginPaneStyle === 'top'" id="social-login-section">
        <social-login-pane
          :register="true"
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
              :state="handleError ? false : null"
              aria-describedby="handle-error"
              :label="$t('register.input.label.contact')"
              autocorrect="off"
              autocomplete="off"
              autocapitalize="off"
              :spellcheck="false"
              data-cy="handle"
            />
            <b-form-invalid-feedback id="handle-error">
              <i18n v-if="isInputHandleAlreadyExists" :path="handleAlreadyExistsErrorKey" tag="span">
                <template #link>
                  <b-link class="text-danger font-weight-bold" :to="to">{{ $t('register.input.error.sign_in') }}</b-link>
                </template>
              </i18n>
              <span v-else>{{ handleError || $t('general.blank') }}</span>
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <b-bsq-input
              v-if="password !== undefined"
              password
              v-model="password"
              :state="passwordError ? false : null"
              aria-describedby="password-error"
              :label="$t('register.input.label.password')"
              type="password"
            />
            <b-form-invalid-feedback id="password-error">
              {{ passwordError || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row
          v-if="password !== ''"
          class="mb-4"
        >
          <b-col>
            <password-strength-indicator
              :password="password"
              @score="$store.commit('authn/SET_PASSWORD_SCORE', $event)"
            />
          </b-col>
        </b-row>
        <b-row
          v-if="privacyURL"
          class="mb-4"
        >
          <b-col>
            <div v-if="privacyCheckbox">
              <b-form-checkbox
                v-model="privacyChecked"
                :state="privacyError ? false : null"
              >
                <!-- TODO: Fix the i18n string path -->
                <i18n
                  class="align-middle"
                  path="register.text.privacy_policy"
                  tag="div"
                >
                  <template #link>
                    <b-link
                      :href="privacyURL"
                      target="_blank"
                    >
                      {{ privacyLinkText }}
                    </b-link>
                  </template>
                </i18n>
              </b-form-checkbox>
              <b-form-invalid-feedback
                class="ml-4"
                :state="privacyError ? false : null"
              >
                {{ privacyError }}
              </b-form-invalid-feedback>
            </div>
            <b-link
              v-else
              :href="privacyURL"
              target="_blank"
            >
              {{ privacyLinkText }}
            </b-link>
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
              {{ $t('register.button.register') }}
            </with-loading-button>
          </b-col>
        </b-row>
        <b-row v-if="linkEnabled">
          <b-col class="text-center">
            <router-link :to="to"
                         class="font-weight-bold"
                         data-cy="register-link"
            >
              {{ $t('register.link.sign_in') }}
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
          :register="true"
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
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

import { INPUT_HANDLE_ALREADY_EXISTS_ERROR } from '@/store/types'

export default {
  name: 'StartPane',
  components: {
    WithLoadingButton,
    SocialLoginPane,
    PasswordStrengthIndicator
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
      'privacyURL',
      'privacyCheckbox',
      'buttonSize',
      'socialLoginPaneStyle',
      'socialLoginPaneOption'
    ]),
    ...mapState('authn', [
      'signUpErrors',
      'loading',
      'handleType'
    ]),
    handleError () {
      return this.signUpErrors && this.signUpErrors.handle
    },
    passwordError () {
      return this.signUpErrors && this.signUpErrors.password
    },
    privacyError () {
      return this.signUpErrors && this.signUpErrors.privacy
    },
    separatorText () {
      return this.$t('register.text.or')
    },
    privacyLinkText () {
      return this.$t('register.link.privacy_link')
    },
    linkEnabled () {
      return this.widgetsSettings.signUpEnabled
    },
    isInputHandleAlreadyExists () {
      return this.handleError === INPUT_HANDLE_ALREADY_EXISTS_ERROR
    },
    handleAlreadyExistsErrorKey () {
      if (this.handleType === 'email') {
        return 'register.input.error.email_already_exists'
      }
      if (this.handleType === 'phone') {
        return 'register.input.error.phone_already_exists'
      }
      return ''
    },
    to () {
      return {
        name: 'SignIn',
        query: this.mergedQuery
      }
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
    },
    privacyChecked: {
      get () {
        return this.$store.state.authn.privacyChecked
      },
      set (value) {
        this.$store.commit('authn/SET_PRIVACY_CHECKED', value)
      }
    }
  },

  created () {
    this.handle = this.$route.query.handle
    this.$route.query.handle = ''
    this.mergedQuery = this.$route.query
  },

  methods: {
    ...mapActions('authn', [
      'signUp'
    ]),

    onSubmit () {
      this.mergedQuery.handle = this.handle
      this.signUp({
        redirectURI: this.redirectURI,
        privacyCheckbox: this.privacyCheckbox
      })
    }
  }
}
</script>
