<template>
  <div ref="widget-container" class="container widget-container">
    <div class="row row-no-gutters justify-content-center min-vh-100">
      <div class="col-md-8 col-lg-6 widget-content px-0">
        <transition
          :name="transitionName"
          mode="out-in">
          <router-view />
        </transition>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

import { AuthCoreAuthClient } from 'authcore-js'

import store from '@/store'
import { colourLuminance } from '@/utils/colour'

export default {
  data () {
    return {
      transitionName: '',
      height: 0
    }
  },

  computed: {
    ...mapState('widgets', [
      'displayMode'
    ]),
    ...mapState('client', [
      'clientId',
      'company',
      'primary',
      'success',
      'danger',
      'buttonSize',
      'socialLoginPaneOption',
      'socialLoginPaneStyle',
      'layoutStateSet',
      'authcoreClient',
      'widgetsSettings'
    ]),
    ...mapGetters('client', [
      'isReady'
    ])
  },

  mounted () {
    try {
      // parse JSON twice as apiserver do JSON serialize twice due to grpc constraint
      // JSON.parse will throw if parse error
      const settingsJSON = JSON.parse(window.__authcore_settings__)
      store.commit('client/SET_WIDGETS_SETTINGS', settingsJSON)
    } catch (err) {
      store.commit('client/GET_WIDGETS_SETTINGS_FAILED', err)
    }

    try {
      this.initAnalytics(this.widgetsSettings.analyticsToken)
    } catch (err) {
      console.error('fail to initialize analytics', err)
    }

    setInterval(() => {
      if (!this.isReady) return
      this.height = this.$refs['widget-container'].clientHeight
    }, 50)

    if (window.performance) {
      var measures = window.performance.getEntriesByType('navigation')
      if (measures.length > 0) {
        this.logAnalytics('Authcore_performance', measures[0].toJSON(), true)
      }
    }

    window.addEventListener('message', this.handleOpenOauthArbiterMessage)
  },
  beforeDestroy () {
    window.removeEventListener('message', this.handleOpenOauthArbiterMessage)
  },
  watch: {
    height () {
      this.postMessage('AuthCore_updateHeight', {
        height: this.height
      })
    },
    async $route (to, from) {
      // $route is watched until $route.query['cid'] exists, so the container id could be correct.
      //
      // Originally the function is called only upon mounted.  However this did not work as the
      // path would be '/' at the time that the widget is mounted.  It would be changed after a
      // while to '/widget-link?cid=THE_CONTAINER_ID'.  This is a temporary fix as the root cause
      // is still unknown.
      //
      // Ref: https://gitlab.com/blocksq/authcore/issues/102
      const toDepth = to.path.split('/').length
      const toIndex = to.meta.index
      const fromDepth = from.path.split('/').length
      const fromIndex = from.meta.index

      // Not apply animation if the widget is first loaded.
      if (from.name !== null) {
        if (toIndex) {
          this.transitionName = toIndex < fromIndex ? 'slide-right' : 'slide-left'
        } else {
          this.transitionName = toDepth < fromDepth ? 'slide-right' : 'slide-left'
        }

        // Custom the transition from verification back to contact list in creating new contact
        if (from.name === 'Verification' && to.name === 'Contacts') {
          this.transitionName = 'slide-right'
        }
      }
      if (from.name === 'FullResetPassword' || from.name === 'ResetPasswordCompleted') {
        this.transitionName = 'slide-left'
      }
      this.getStatesFromQuery()

      // Instantiate the client only if it does not exist
      if (!this.authcoreClient) {
        const authClient = await new AuthCoreAuthClient({
          clientId: this.clientId,
          apiBaseURL: window.origin,
          callbacks: {
            unauthenticated: () => {
              // Notify unauthenticated to the client
              store.commit('client/SET_AUTHENTICATED_STATE', false)
              this.postMessage('AuthCore_unauthenticated', {})
            }
          }
        })
        store.commit('client/GET_CLIENT_COMPLETED', authClient)
      }

      this.$nextTick(function () {
        store.dispatch('preferences/init', this.widgetsSettings.preferences)
        this.postMessage('AuthCore_onLoaded', {})
        this.logAnalytics('Authcore_loginWidgetLoaded', {})
      })
    }
  },

  methods: {
    // called during page route to get states from this.$route.query
    getStatesFromQuery () {
      const logo = this.$route.query.logo
      const company = this.$route.query.company
      const clientId = this.$route.query.clientId

      let primaryColour = this.$route.query.primaryColour
      let successColour = this.$route.query.successColour
      let dangerColour = this.$route.query.dangerColour

      let privacyLink = this.$route.query.privacyLink
      const privacyCheckbox = this.$route.query.privacyCheckbox

      const requireUsername = this.$route.query.requireUsername

      let language = this.$route.query.language

      if (primaryColour === 'undefined') {
        primaryColour = undefined
      }
      if (successColour === 'undefined') {
        successColour = undefined
      }
      if (dangerColour === 'undefined') {
        dangerColour = undefined
      }

      if (privacyLink === 'undefined') {
        privacyLink = undefined
      }

      if (language === 'undefined') {
        language = undefined
      }

      const cid = this.$route.query.cid
      const internal = this.$route.query.internal
      const verification = this.$route.query.verification
      const showAvatar = this.$route.query.showAvatar
      const successRedirectUrl = this.$route.query.successRedirectUrl
      const buttonSize = this.$route.query.buttonSize
      const socialLoginPaneOption = this.$route.query.socialLoginPaneOption
      const socialLoginPaneStyle = this.$route.query.socialLoginPaneStyle
      const fixedContact = this.$route.query.fixedContact
      let prefillContact = this.$route.query.contact
      if (prefillContact === 'undefined') {
        prefillContact = ''
      }

      // UI related layout states
      const layoutState = {
        logo: decodeURIComponent(logo),
        company: company,
        buttonSize: buttonSize,
        socialLoginPaneOption: socialLoginPaneOption,
        socialLoginPaneStyle: socialLoginPaneStyle,
        language: language
      }

      // With cid, the instance comes from AuthCore.{widget_name}
      // As react-native flow do not have cid, this is bypassed and only set the layout state when not set
      // (as normal flow may also not have cid routing to other pages)
      if (cid) {
        store.commit('client/GET_CONTAINER_PARAMETER_COMPLETED', {
          clientId: clientId,
          containerId: cid,
          internal: internal,
          primary: primaryColour,
          success: successColour,
          danger: dangerColour,
          verification: verification,
          privacyLink: privacyLink,
          privacyCheckbox: privacyCheckbox,
          successRedirectUrl: successRedirectUrl,
          requireUsername: requireUsername,
          showAvatar: showAvatar,
          fixedContact: fixedContact,
          prefillContact: prefillContact
        })
        store.commit('client/SET_LAYOUT_STATE', layoutState)
        this.logAnalytics('Authcore_UISettings', layoutState)
      } else if (!this.layoutStateSet) {
        // Set the social login section position and language if it is not set (case in react-native)
        // If it is normal flow, it should be set (as it have cid) and it is set afterwards
        // This pattern shall be expanded in getting widgets settings from settings
        // https://gitlab.com/blocksq/authcore/issues/593
        store.commit('client/SET_LAYOUT_STATE', layoutState)
        this.logAnalytics('Authcore_UISettings', layoutState)
      }

      if (primaryColour !== undefined) {
        const primaryDarken = colourLuminance(primaryColour, -0.2)
        document.documentElement.style.setProperty('--primary', this.primary)
        document.documentElement.style.setProperty('--primary-darken', primaryDarken)
        document.documentElement.style.setProperty('--primary-half-transparent', `${this.primary}80`)
      }
      if (successColour !== undefined) {
        const successDarken = colourLuminance(successColour, -0.2)
        document.documentElement.style.setProperty('--success', this.success)
        document.documentElement.style.setProperty('--success-darken', successDarken)
        document.documentElement.style.setProperty('--success-half-transparent', `${this.success}80`)
      }
      if (dangerColour !== undefined) {
        const dangerDarken = colourLuminance(dangerColour, -0.2)
        document.documentElement.style.setProperty('--danger', this.danger)
        document.documentElement.style.setProperty('--danger-darken', dangerDarken)
        document.documentElement.style.setProperty('--danger-half-transparent', `${this.danger}80`)
      }
    },

    handleOpenOauthArbiterMessage (e) {
      // Handle Authcore_openOauthArbiter message from oauthWindow
      if (typeof e.data !== 'object') return
      if (e.origin !== window.origin) return
      const { type, data } = e.data
      if (type !== 'Authcore_openOauthArbiter') return
      this.$router.push({
        name: 'OauthArbiter',
        query: data
      })
    }
  }
}

</script>

<style lang="scss">
$container-max-height: 600px;
body {
    background-color: transparent;
}

.widget-content {
    @media (min-width: 768px) {
        @media (min-height: $container-max-height) {
            /* Only show the border in case of sign in widget is shown without scrolling */
            border: 1px solid #dfdfdf;
        }
        border-radius: 5px;
        margin-top: auto;
        margin-bottom: auto;
        -ms-flex-item-align: center !important;
        align-self: center !important;

        max-height: $container-max-height;
        overflow-x: hidden;
        overflow-y: auto;
        scrollbar-width: none; /* For Firefox */
        &::-webkit-scrollbar { /* For Chrome and WebKit */
            width: 0px;
        }
    }
    /* Ensure showning the whole widget as well as provide dynamic height to show more information */
    @media (min-height: $container-max-height / 0.65) {
        max-height: 65vh;
    }
}
</style>
