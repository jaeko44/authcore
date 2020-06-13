<template>
  <widget-template
    logo-enabled
    :back-enabled="backEnabled"
    :logo="logo"
    :title="currentTitle"
    :headerError="errorContent"
  >
    <b-form class="mb-3" @submit="resetPassword">
      <transition-group tag="div" name="prompt-from-bottom">
        <div v-if="step === STEP.INVALID_PASSWORD_LINK" key="invalid_password_link">
          <b-row class="mb-4">
            <b-col class="h4 text-center text-danger">
              {{ $t(errorContent) }}
            </b-col>
          </b-row>
          <b-row>
            <b-col class="text-center">
              <b-link :href="redirectUri">
                {{ $t('reset_password.text.return_home') }}
              </b-link>
            </b-col>
          </b-row>
        </div>
        <div v-else-if="step === STEP.REACH_LIMIT" key="reach_limit">
          <b-row class="mb-4">
            <b-col class="text-center">
              {{ $t(errorContent) }}
            </b-col>
          </b-row>
          <b-row>
            <b-col class="text-center">
              <b-link
                replace
                :to="{ name: 'SignIn' }"
              >
                {{ $t('reset_password.text.return_home') }}
              </b-link>
            </b-col>
          </b-row>
        </div>
        <div v-else-if="step === STEP.LOADING" key="loading">
          <loading-spinner />
        </div>
        <div v-else-if="step === STEP.INPUT_HANDLE" key="step1">
          <b-row class="mb-4" align-h="center">
            <b-col class="text-center">
              {{ $t('reset_password.text.send_reset_password_instruction') }}
            </b-col>
          </b-row>
          <b-row>
            <b-col>
              <b-form-contact-input
                v-focus
                v-model="handle"
                :state="handleError"
                aria-describedby="handleError"
                :label="$t('reset_password.input.label.contact')"
                autocorrect="off"
                autocomplete="off"
                autocapitalize="off"
                :spellcheck="false"
              />
              <input class="d-none" name="hiddenPassword" type="password" />
              <b-form-invalid-feedback id="handleError">{{ $t(errorContent) }}</b-form-invalid-feedback>
            </b-col>
          </b-row>
          <b-row>
            <b-col>
              <b-button
                block
                type="submit"
                variant="primary"
                data-cy="send"
              >
                {{ $t('reset_password.button.send') }}
              </b-button>
            </b-col>
          </b-row>
        </div>
        <div v-else-if="step === STEP.MESSAGE_SENT" key="step2">
          <!-- TODO: Set for confirmation page before -->
          <b-row class="mb-4">
            <b-col class="text-center">
              {{ $t('reset_password.text.check_contact_instruction', { contact: $t(contactType) }) }}
            </b-col>
          </b-row>
          <b-row class="mb-4">
            <b-col class="text-center">
              <h5 class="text-grey-dark">{{ handle }}</h5>
            </b-col>
          </b-row>
          <b-row>
            <b-col>
              <b-button
                block
                type="submit"
                variant="primary"
                data-cy="ok"
              >
                {{ $t('reset_password.button.return_home') }}
              </b-button>
            </b-col>
          </b-row>
        </div>
        <div v-else-if="step === STEP.INPUT_PASSWORD" key="step3">
          <b-row>
            <b-col>
              <input
                class="d-none"
                name="identifier"
                autocomplete="off"
                tabindex="-1"
                spellcheck="false"
                type="text"
                :value="identifier"
              />
              <b-bsq-input
                password
                v-model="password"
                class="my-3"
                :disabled="loading"
                aria-describedby="passwordError"
                autocomplete="new-password"
                :label="$t('reset_password.input.label.password')"
                type="password"
              />
              <b-form-invalid-feedback
                v-if="password === ''"
                id="passwordError"
              >
                {{ $t('general.blank') }}
              </b-form-invalid-feedback>
              <password-strength-indicator
                v-else
                :password="password"
                @score="passwordScore"
              />
              <b-bsq-input
                password
                v-model="confirmPassword"
                class="my-3"
                :disabled="loading"
                aris-describedby="confirmPasswordError"
                autocomplete="new-password"
                :label="$t('reset_password.input.label.confirm_password')"
                type="password"
              />
              <b-form-invalid-feedback
                id="confirmPasswordError"
              >
                {{ $t('general.blank') }}
              </b-form-invalid-feedback>
            </b-col>
          </b-row>
          <b-row class="mt-1">
            <b-col>
              <b-button
                block
                :disabled="loading || password !== confirmPassword || score < 2"
                type="submit"
                variant="primary"
              >
                {{ $t('reset_password.button.reset_password') }}
              </b-button>
            </b-col>
          </b-row>
        </div>
        <div v-else-if="step === STEP.COMPLETED" key="step4">
          <b-row class="mb-3">
            <b-col class="text-center">
              {{ $t('reset_password.text.reset_password_success') }}
            </b-col>
          </b-row>
          <b-row>
            <b-col>
              <b-button v-if="redirectUri" block :href="redirectUri" variant="primary">{{ $t('reset_password.button.sign_in') }}</b-button>
              <b-button v-else block type="submit" variant="primary">{{ $t('reset_password.button.sign_in') }}</b-button>
            </b-col>
          </b-row>
        </div>
      </transition-group>
    </b-form>
  </widget-template>
</template>

<script>
import { mapState } from 'vuex'

import router from '@/router'
import store from '@/store'
import { i18n } from '@/i18n-setup'

import WidgetTemplate from '@/components/WidgetTemplate.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

const STEP = {
  REACH_LIMIT: -2,
  INVALID_PASSWORD_LINK: -1,
  LOADING: 0,
  INPUT_HANDLE: 1,
  MESSAGE_SENT: 2,
  INPUT_PASSWORD: 3,
  COMPLETED: 4
}

export default {
  name: 'ResetPassword',
  components: {
    WidgetTemplate,
    LoadingSpinner,
    PasswordStrengthIndicator
  },

  props: {
    company: {
      type: String
    },
    logo: {
      type: String
    },
    contactToken: {
      type: String
    },
    redirectUri: {
      type: String
    },
    identifier: {
      type: String
    },
    prefillHandle: {
      type: String,
      default: ''
    }
  },

  data () {
    return {
      // User inputs
      handle: '',
      password: '',
      confirmPassword: '',
      score: -1,
      handleFieldError: undefined,

      STEP
    }
  },

  computed: {
    ...mapState('widgets/account/resetPasswordForm', [
      'authenticateHandleDone',
      'resetPasswordDone',
      'selectedChallengeMethod',
      'authorizationToken',
      'loading',
      'error'
    ]),
    step () {
      if (this.loading) {
        return STEP.LOADING
      } else if (this.resetPasswordDone) {
        // Even though reset password process is done remains in loading status
        // Wait the page to transit to ResetPasswordCompleted page to allow
        // save password prompt to show, allowing the user to save new password.
        return STEP.LOADING
      } else if (this.authorizationToken) {
        return STEP.INPUT_PASSWORD
      } else if (this.authenticateHandleDone) {
        return STEP.MESSAGE_SENT
      } else if (!this.authenticateHandleDone && this.error && this.contactToken) {
        // Case for invalid password link which includes contact token
        return STEP.INVALID_PASSWORD_LINK
      } else if (this.error === 'reset_password.text.error.reach_limit') {
        return STEP.REACH_LIMIT
      } else {
        return STEP.INPUT_HANDLE
      }
    },
    backEnabled () {
      switch (this.step) {
        case 1:
          return true
        default:
          return false
      }
    },
    currentTitle () {
      switch (this.step) {
        case STEP.LOADING:
        case STEP.INPUT_HANDLE:
        case STEP.MESSAGE_SENT:
        case STEP.INPUT_PASSWORD:
        case STEP.COMPLETED:
          return i18n.t('reset_password.title')
        case STEP.INVALID_PASSWORD_LINK:
          return i18n.t('')
        default:
          return ''
      }
    },
    errorContent () {
      if (this.handleFieldError) {
        return this.handleFieldError
      }
      return this.error === undefined ? 'general.blank' : this.error
    },
    handleError () {
      if (this.handleFieldError) {
        return false
      }
      return this.error === undefined ? null : false
    },
    factorError () {
      return this.error === undefined ? null : false
    },
    contactType () {
      // Return type of contact to decide wordings, phone considered to be +{number}
      return /^\+\d+$/.test(this.handle) ? 'reset_password.text.phone' : 'reset_password.text.email'
    }
  },

  watch: {},

  async mounted () {
    if (this.prefillHandle) {
      this.handle = this.prefillHandle
    }
    if (this.contactToken) {
      await store.dispatch(
        'widgets/account/resetPasswordForm/authenticateWithContact',
        this.contactToken
      )
    }
  },

  destroyed () {},

  methods: {
    passwordScore (result) {
      this.score = result
    },
    async clickNext () {
      switch (this.step) {
        case STEP.INPUT_HANDLE:
          if (this.handle === '') {
            this.handleFieldError = 'reset_password.input.error.blank_contact'
          } else {
            this.handleFieldError = undefined
            await store.dispatch(
              'widgets/account/resetPasswordForm/authenticateHandle',
              this.handle
            )
          }
          break
        case STEP.MESSAGE_SENT:
          store.commit('widgets/account/resetPasswordForm/CLEAR_STATES')
          router.replace({ name: 'SignIn' })
          break
        case STEP.INPUT_PASSWORD:
          await store.dispatch(
            'widgets/account/resetPasswordForm/resetPassword', {
              password: this.password,
              confirmPassword: this.confirmPassword
            }
          )
          router.push({
            name: 'ResetPasswordCompleted',
            query: {
              company: this.company,
              logo: this.logo,
              redirect_uri: this.redirectUri
            }
          })
          break
        case STEP.COMPLETED:
          router.push({ name: 'SignIn' })
          break
      }
    },

    async resetPassword (e) {
      e.preventDefault()
      await this.clickNext()
    }
  }
}
</script>
