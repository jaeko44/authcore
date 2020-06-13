<template>
  <!-- loading state in widget-template decides showing the border or not -->
  <!-- For new contact from registration, the border is hidden in loading status or it is finished -->
  <component
    v-bind:is="baseTemplate"
    logo-enabled
    :logo="logo"
    :title="currentTitle"
    :headerError="error"
    :loading="!oldContactFlag && (loading || spinnerLoading || done)"
  >
    <div>
      <b-form @submit="onSubmit">
        <transition
          name="fadeIn"
          mode="out-in"
        >
          <div v-if="step === STEP.CODE_INPUT" key="form">
            <!-- Apart from normal loading case, loading screen will be prompted at the end of registration flow (either finish by entered verification code or skip that). -->
            <loading-spinner
              v-if="loadingStatus"
              @spinning="spinning"
            />
            <div v-else>
              <b-row class="mb-4">
                <b-col class="text-center">
                  {{ contentDescription }}
                </b-col>
              </b-row>
              <b-row class="mb-3">
                <b-col>
                  <b-bsq-input
                    v-focus
                    v-model="verificationCode"
                    :state="verificationCodeError"
                    aria-describedby="verificationCodeError"
                    :label="$t('verification.input.label.verification_code')"
                    autocomplete="off" />
                  <b-form-invalid-feedback
                    id="verificationCodeError"
                    class="d-inline-block w-50"
                  >
                    {{ $t(error) }}
                  </b-form-invalid-feedback>
                  <div class="d-inline-flex w-50 justify-content-end">
                    <b-link
                      v-if="!resendDone"
                      :to="{ name: 'ResendVerification', params: { type: contact.type.toLowerCase() } }"
                    >
                      {{ $t('verification.link.resend_code') }}
                    </b-link>
                    <span
                      v-else
                      class="text-grey-medium"
                    >
                      {{ $t('verification.text.code_sent') }}
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
                    {{ $t('verification.button.verify') }}
                  </b-button>
                </b-col>
              </b-row>
              <b-row v-if="oldContactFlag">
                <b-col class="text-center">
                  <b-link @click="goBack">{{ $t('verification.link.cancel') }}</b-link>
                </b-col>
              </b-row>
              <b-row v-else>
                <b-col class="text-center">
                  <b-link @click="skipVerification">{{ $t('verification.link.verify_later') }}</b-link>
                </b-col>
              </b-row>
            </div>
          </div>
          <div v-else-if="step === STEP.COMPLETED" key="success">
            <div class="mb-3">
              <successful-show-up exclude-wording />
            </div>
            <b-row>
              <b-col class="text-center">
                <b-button
                  block
                  :class="{ 'w-75': buttonSize === 'normal' }"
                  class="d-inline-block"
                  type="submit"
                  variant="primary"
                >
                  {{ $t('verification.button.ok') }}
                </b-button>
              </b-col>
            </b-row>
          </div>
        </transition>
      </b-form>
      <form
        id="oauth-callback-form"
        action="/redirect-oauth"
        method="GET">
        <input
          name="state"
          type="hidden"
          :value="state" />
        <input
          name="code"
          type="hidden"
          :value="authorizationToken" />
        <input
          name="url"
          type="hidden"
          :value="redirectUri" />
      </form>
    </div>
  </component>
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'
import router from '@/router'
import { i18n } from '@/i18n-setup'

import { redirectTo } from '@/utils/util'

import WidgetTemplate from '@/components/WidgetTemplate.vue'
import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import SuccessfulShowUp from '@/components/SuccessfulShowUp.vue'

const STEP = {
  CODE_INPUT: 1,
  COMPLETED: 2
}

export default {
  name: 'Verification',
  components: {
    WidgetTemplate,
    WidgetLayoutV2,
    LoadingSpinner,
    SuccessfulShowUp
  },

  props: {
    // Default to be -1 for contact from new account
    contactId: {
      type: Number,
      default: -1
    },
    contactType: {
      type: String
    },
    contactValue: {
      type: String
    },
    newContact: {
      type: Boolean,
      default: false
    },
    callbackAction: {
      default: function () {
        const action = this.contactId !== -1 ? 'verification' : 'register'
        return action
      }
    },
    state: {
      type: String
    },
    redirectUri: {
      type: String
    }
  },

  data () {
    return {
      logo: '',
      verificationCode: '',
      token: undefined,

      STEP
    }
  },

  computed: {
    ...mapState('client', [
      'containerId',
      'verification',
      'authcoreClient',
      'successRedirectUrl',

      'oAuthWidget',
      'buttonSize'
    ]),
    ...mapState('widgets/account', [
      'user'
    ]),
    ...mapState('widgets/account/verifyAccountForm', [
      'contact',
      'authorizationToken',
      'oldContactFlag',
      'loading',
      'spinnerLoading',
      'done',
      'error',
      'resendDone',
      'createAuthorizationTokenDone'
    ]),
    baseTemplate () {
      // For register flow, the template used is refactored version which does not updated in verification page in profile contact, using this to differentiate the template for Verification view
      if (this.oldContactFlag) {
        return 'widget-template'
      }
      return 'widget-layout-v2'
    },
    step () {
      // For new account verification (this.oldContactFlag to be false), not using completed screen
      // Instead using loading spinner and redirect to url set by client
      return this.done && !this.spinnerLoading && this.oldContactFlag ? STEP.COMPLETED : STEP.CODE_INPUT
    },
    verificationCodeError () {
      return this.error === undefined ? null : false
    },
    loadingStatus () {
      return this.loading || this.spinnerLoading || (this.done && !this.oldContactFlag)
    },
    currentTitle () {
      if (this.loadingStatus) {
        return ''
      }
      switch (this.step) {
        case 1: {
          if (this.oldAccountContact) {
            return i18n.t('verification.title.contact_verification')
          } else {
            return i18n.t('verification.title.account_verification')
          }
        }
        case 2: {
          if (this.oldAccountContact) {
            return i18n.t('verification.title.contact_verification')
          } else {
            return i18n.t('verification.title.account_created')
          }
        }
        default: return ''
      }
    },
    contentDescription () {
      return i18n.t('verification.text.contact_verification', {
        contact: this.contactValue !== undefined ? this.contactValue : this.contact.value
      })
    },
    currentDescription () {
      switch (this.step) {
        case 1: {
          return i18n.t('verification.text.contact_verification', {
            contact: this.contactValue !== undefined ? this.contactValue : this.contact.value
          })
        }
        case 2: return i18n.t('general.blank')
        default: return ''
      }
    },
    // Denote whether it is old contact verification or not
    oldAccountContact () {
      // contactId is the prop passes into this view, denotes whether old account contact when it first directed into this view. It will be reseted to -1 if it browses to other page in verification step(e.g. ResendVerification)
      // oldContactFlag is set when it first directed into this view. It denotes whether old account contact within verification process
      return this.oldContactFlag || this.contactId !== -1
    }
  },

  watch: {
    done () {
      if (this.successRedirectUrl || this.oAuthWidget) {
        store.dispatch('widgets/account/verifyAccountForm/createAuthorizationToken', {
          codeChallenge: this.codeChallenge,
          codeChallengeMethod: this.codeChallengeMethod
        })
      } else if (this.oldAccountContact && this.callbackAction !== undefined) {
        const callbackMessage = {
          action: this.callbackAction
        }
        this.postMessage('AuthCore_onSuccess', callbackMessage)
      }
    },
    resendDone (newVal) {
      if (newVal) {
        // Set 10 seconds to disable the reset link
        setTimeout(() => {
          store.commit('widgets/account/verifyAccountForm/RESEND_VERIFY_ACCOUNT_CONTACT_INIT')
        }, 10000)
      }
    },
    // Authorization token created implies registration flow is done
    createAuthorizationTokenDone (newVal) {
      if (newVal) {
        setTimeout(() => {
          this.logAnalytics('Authcore_verifySuccess', {}, true)
          this.dismiss()
        }, 500)
      }
    }
  },

  async mounted () {
    if (this.contact.value === undefined) {
      store.commit('widgets/account/verifyAccountForm/VERIFY_ACCOUNT_CONTACT_SPINNER_STATE', true)
      // Check if new account contact to process with different flow
      if (this.oldAccountContact) {
        // Flow for oldAccountContact is DEPRECATED, keep it for later refactoring.
        // For existing contact, using the following flow.
        await store.dispatch('widgets/account/verifyAccountForm/contactVerificationInitFromId', this.contactId)
        if (!this.newContact) {
          await store.dispatch('widgets/account/verifyAccountForm/resendContactVerification')
        }
        // End of DEPRECATED flow
      } else {
        // For new contact, using the following flow.
        this.token = sessionStorage.getItem('io.authcore.temporaryToken')
        sessionStorage.removeItem('io.authcore.temporaryToken')
        await this.authcoreClient.setAccessToken(this.token)
        if (this.verification) {
          // Contact verification for next account which has single contact
          store.dispatch('widgets/account/verifyAccountForm/contactVerificationInit')
        } else {
          // Complete the verification process if no token is provided
          store.commit('widgets/account/verifyAccountForm/VERIFY_ACCOUNT_CONTACT_COMPLETED')
        }
      }
    }
  },
  updated () {},
  destroyed () {},

  methods: {
    spinning (status) {
      store.commit('widgets/account/verifyAccountForm/VERIFY_ACCOUNT_CONTACT_SPINNER_STATE', status)
    },
    onSubmit (e) {
      e.preventDefault()
      switch (this.step) {
        case STEP.CODE_INPUT:
          this.sendVerification()
          break
        case STEP.COMPLETED:
          this.logAnalytics('Authcore_verifySuccess', {}, true)
          this.dismiss()
          break
        default:
          break
      }
    },
    sendVerification () {
      store.dispatch('widgets/account/verifyAccountForm/primaryContactVerification', {
        type: this.contactType || this.contact.type,
        verifyCode: this.verificationCode
      })
    },
    skipVerification () {
      store.commit('widgets/account/verifyAccountForm/VERIFY_ACCOUNT_CONTACT_COMPLETED')
    },
    goBack () {
      router.go(-1)
    },
    dismiss () {
      if (this.oldAccountContact) {
        // For existing contact, go back to the profile page
        router.go(-1)
        store.commit('widgets/account/verifyAccountForm/CLEAR_STATES')
        this.postMessage('AuthCore_onSuccess', {

        })
      } else {
        // For contact from new account, trigger the callback for the client to implement corresponding function.
        store.dispatch('widgets/account/get')
        const idToken = sessionStorage.getItem('io.authcore.idToken')
        sessionStorage.removeItem('io.authcore.idToken')
        // For register case in mobile for redirection
        if (this.oAuthWidget) {
          if (this.createAuthorizationTokenDone) {
            // On Android Chrome, the address cannot be updated by `window.location` as it requires
            // user interaction. However form submission not, so we are using form submission.
            setTimeout(() => {
              document.getElementById('oauth-callback-form').submit()
            }, 10)
          }
        } else {
          // For register case in desktop or OAuth for redirection
          if (this.successRedirectUrl || this.redirectUri) {
            // Follow the redirection flow using `successRedirectUrl`, the corresponding item is not used anymore.
            let redirectUrl
            if (this.successRedirectUrl) {
              redirectUrl = new URL(this.successRedirectUrl)
            } else {
              redirectUrl = new URL(this.redirectUri)
            }
            redirectUrl.searchParams.set('code', this.authorizationToken)
            redirectUrl.searchParams.set('state', this.$store.query.clientState)
            redirectTo(redirectUrl.toString(), this.containerId)
          } else {
            this.postMessage('AuthCore_onSuccess', {
              action: this.callbackAction,
              current_user: this.user,
              access_token: this.authcoreClient.getAccessToken(),
              id_token: idToken
            })
          }
        }
      }
    }
  },

  beforeRouteEnter (to, from, next) {
    // Get the logo from previous query (Should be Register component). May change the hack to save the query in store
    next(vm => {
      vm.logo = from.query.logo
    })
  }
}
</script>
