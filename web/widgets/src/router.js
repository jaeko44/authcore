import Vue from 'vue'
import Router from 'vue-router'

import store from '@/store'
import { i18n } from '@/i18n-setup'

// Function for lazy loading view with loading and error component in vue-router
// As routes can only resolve to a single component in Vue Router, a helper
// function is required to support the async component factory function with
// loading, error, delay and timeout key.
// Reference: https://github.com/vuejs/vue-router/pull/2140/files
function lazyLoadView (AsyncView) {
  const AsyncHandler = () => ({
    component: AsyncView,
    // A component to use while the component is loading.
    loading: require('@/components/LoadingSpinner.vue').default
    // A fallback component in case the timeout is exceeded
    // when loading the component.
    // Delay before showing the loading component.
    // Default: 200 (milliseconds).
    // delay: 400,
    // Time before giving up trying to load the component.
    // Default: Infinity (milliseconds).
    // timeout: 10000
  })
  return Promise.resolve({
    functional: true,
    render (h, { data, children }) {
      // Transparently pass any props or children
      // to the view component.
      return h(AsyncHandler, data, children)
    }
  })
}

const SignIn = () => import(/* webpackChunkName: "signin" */ './views/SignIn.vue')
const SignUp = () => lazyLoadView(import(/* webpackChunkName: "password-related" */ './views/SignUp.vue'))
// const ResetPassword = () => import(/* webpackChunkName: "signin" */ './views/ResetPassword.vue')
// const FullResetPassword = () => import(/* webpackChunkName: "signin" */ './views/FullResetPassword.vue')
const ResetPassword = () => lazyLoadView(import(/* webpackChunkName: "password-related" */ './views/ResetPassword.vue'))
const FullResetPassword = () => lazyLoadView(import(/* webpackChunkName: "password-related" */ './views/FullResetPassword.vue'))
const AddRecoveryEmail = () => lazyLoadView(import(/* webpackChunkName: "signin" */ './views/AddRecoveryEmail.vue'))

const ResetPasswordCompleted = () => import(/* webpackChunkName: "signin" */ './views/ResetPasswordCompleted.vue')
const Verification = () => import(/* webpackChunkName: "signin" */ './views/Verification.vue')
const ResendVerification = () => import(/* webpackChunkName: "signin" */ './views/ResendVerification.vue')
const ExternalOauthCallback = () => import(/* webpackChunkName: "signin" */ './views/ExternalOauthCallback.vue')
const OauthArbiter = () => import(/* webpackChunkName: "signin" */ './views/OauthArbiter.vue')
const ErrorPage = () => import('./views/ErrorPage.vue')

const Settings = () => import(/* webpackChunkName: "settings" */ './views/Settings.vue')
const SettingsHome = () => import(/* webpackChunkName: "settings" */ './views/settings/SettingsHome.vue')
const ChangePassword = () => lazyLoadView(import(/* webpackChunkName: "password-related" */ './views/settings/ChangePassword.vue'))

const SocialLogins = () => import(/* webpackChunkName: "settings" */ './views/settings/SocialLogins.vue')
const SocialLoginDelete = () => import(/* webpackChunkName: "settings" */ './views/settings/SocialLoginDelete.vue')
const MFAList = () => import(/* webpackChunkName: "settings" */ './views/settings/MFAList.vue')
const CreateAuthenticatorApp = () => import(/* webpackChunkName: "settings" */ './views/settings/mfa_methods/CreateAuthenticatorApp.vue')
const ManageAuthenticatorApp = () => import(/* webpackChunkName: "settings" */ './views/settings/mfa_methods/ManageAuthenticatorApp.vue')
const RemoveAuthenticatorApp = () => import(/* webpackChunkName: "settings" */ './views/settings/mfa_methods/RemoveAuthenticatorApp.vue')

const Devices = () => import(/* webpackChunkName: "settings" */ './views/settings/Devices.vue')
const DeviceDelete = () => import(/* webpackChunkName: "settings" */ './views/settings/DeviceDelete.vue')

const RefreshToken = () => import(/* webpackChunkName: "token" */ './views/RefreshToken.vue')

Vue.use(Router)

const router = new Router({
  mode: 'history',
  base: '/widgets/',
  routes: [{
    path: '/',
    component: SignIn
  }, {
    path: '/signin',
    name: 'SignIn',
    component: SignIn,
    meta: {
      title: 'sign_in.title.sign_in'
    }
  }, {
    path: '/register',
    name: 'SignUp',
    component: SignUp
  }, {
    path: '/recovery-email/add',
    name: 'AddRecoveryEmail',
    component: AddRecoveryEmail
  }, {
    path: '/reset-password',
    name: 'ResetPassword',
    component: ResetPassword,
    props: (route) => ({
      company: route.query.company,
      logo: route.query.logo,
      prefillHandle: route.query.prefill_handle
    })
  }, {
    path: '/reset-password/contact/:contactToken',
    name: 'FullResetPassword',
    component: FullResetPassword,
    props: (route) => ({
      company: route.query.company,
      logo: route.query.logo,
      contactToken: route.params.contactToken,
      redirectUri: route.query.redirect_uri,
      identifier: route.query.identifier
    }),
    meta: {
      title: 'reset_password.title',
      fullScreen: true
    }
  }, {
    path: '/reset-password/completed',
    name: 'ResetPasswordCompleted',
    component: ResetPasswordCompleted,
    props: (route) => ({
      company: route.query.company,
      logo: route.query.logo,
      redirectUri: route.query.redirect_uri
    }),
    meta: {
      fullScreen: true
    }
  }, {
    path: '/verification',
    name: 'Verification',
    component: Verification,
    meta: {
      index: 2
    },
    props (route) {
      return {
        contactId: route.params.contactId,
        contactType: route.params.contactType,
        contactValue: route.params.contactValue,
        newContact: route.params.newContact,
        redirectUri: route.query.redirect_uri,
        state: route.query.state
      }
    }
  }, {
    path: '/verification/resend',
    name: 'ResendVerification',
    component: ResendVerification,
    props (route) {
      return {
        type: route.params.type,
        value: route.params.value
      }
    }
  }, {
    path: '/settings',
    component: Settings,
    children: [{
      path: '',
      name: 'SettingsHome',
      component: SettingsHome,
      meta: {
        title: 'settings_home.title'
      }
    }, {
      path: 'mfa',
      component: Settings,
      children: [{
        path: 'list',
        name: 'MFAList',
        component: MFAList,
        meta: {
          title: 'mfa_list.title'
        }
      }, {
        path: 'authenticator/:factorId/manage',
        name: 'ManageAuthenticatorApp',
        component: ManageAuthenticatorApp,
        props (route) {
          return {
            factorId: parseInt(route.params.factorId, 10)
          }
        },
        meta: {
          title: 'manage_authenticator_app.title'
        }
      }, {
        path: 'authenticator/:factorId/remove',
        name: 'RemoveAuthenticatorApp',
        component: RemoveAuthenticatorApp,
        props (route) {
          return {
            factorId: parseInt(route.params.factorId, 10)
          }
        },
        meta: {
          title: 'remove_authenticator_app.title'
        }
      }, {
        path: 'authenticator/new',
        name: 'CreateAuthenticatorApp',
        component: CreateAuthenticatorApp,
        meta: {
          title: 'mfa_totp_create.title'
        }
      }]
    }, {
      path: 'devices',
      component: Settings,
      children: [{
        path: '',
        name: 'Devices',
        component: Devices,
        meta: {
          title: 'devices.title'
        }
      }, {
        path: ':deviceId/delete',
        name: 'DeviceDelete',
        component: DeviceDelete,
        props (route) {
          return {
            deviceId: parseInt(route.params.deviceId, 10)
          }
        }
      }]
    }, {
      path: 'password',
      component: ChangePassword,
      name: 'ChangePassword',
      meta: {
        title: 'change_password.title'
      }
    }, {
      path: 'oauth',
      component: Settings,
      children: [{
        path: '',
        name: 'SocialLogins',
        component: SocialLogins,
        meta: {
          title: 'manage_social_logins.title'
        }
      }, {
        path: ':id/delete',
        name: 'SocialLoginDelete',
        component: SocialLoginDelete,
        props (route) {
          return {
            id: route.params.id
          }
        }
      }]
    }]
  }, {
    path: '/external-oauth/:service/cb',
    name: 'ExternalOauthCallback',
    component: ExternalOauthCallback,
    props (route) {
      return {
        service: route.params.service,
        // `code` is the authorization code for OAuth2, while `oauth_verifier` is the verifier for OAuth1.
        code: route.query.code || route.query.oauth_verifier,
        state: route.query.state
      }
    },
    meta: {
      // Default to be full screen case to support redirection flow in mobile
      fullScreen: true
    }
  }, {
    path: '/oauth/arbiter',
    name: 'OauthArbiter',
    component: OauthArbiter,
    props (route) {
      return {
        code: route.query.code,
        oauthVerifier: route.query.oauth_verifier,
        state: route.query.state
      }
    }
  }, {
    path: '/error',
    name: 'ErrorPage',
    component: ErrorPage,
    props (route) {}
  }, {
    path: '/refresh-token',
    component: RefreshToken
  }]
})

// Route setting, check meta field to run corresponding behaviour
router.beforeEach((to, from, next) => {
  if (to.meta.fullScreen) {
    store.commit('widgets/SET_DISPLAY_MODE_STATE', 'full-screen')
  }
  if (to.meta.title !== undefined) {
    document.title = i18n.t(to.meta.title) + ' - Authcore'
  }
  // Update logo and company when accessing to FullResetPassword view
  if (to.name === 'FullResetPassword') {
    const logo = decodeURIComponent(to.query.logo)
    const company = to.query.company || 'Authcore'
    store.commit('client/GET_CONTAINER_PARAMETER_COMPLETED', {
      logo: logo,
      company: company
    })
    // Change the document title to use app name
    document.title = `${i18n.t(to.meta.title)} - ${company}`
    // Change favicon if icon is set
    if (logo !== '') {
      document.getElementById('favicon').href = logo
    }
  }
  if (!to.query.clientId && store.state.preferences.clientId) {
    to.query.clientId = store.state.preferences.clientId
    // Pop alert message out only for the case to inject query string into next route.
    // For most cases in widgets require injecting query string(clientId), if poping pending message for every cases the message will be missed as for the first time the pending message is shown. To pop the message again will remove the message.
    store.commit('alert/POP_MESSAGE')
    next(to)
  } else {
    next()
  }
})

// Override the original push/replace function to catch and show the error only if it is not undefined to ensure meaningful error
// Every routes are injected clientId in query string to keep consistent behaviour for refresh case.
// For every redirect case using `next({})` every routing is aborted and redirect into the route with query.
// As the push/replace retuns a promise, for every aborted case it counts as navigation failure which result in undefined error in the console because the promise rejection is not caught.
// See more: https://github.com/vuejs/vue-router/issues/2881#issuecomment-520554378
// Reference for the workaround: https://github.com/vuejs/vue-router/issues/2932#issuecomment-533453711
const originalPush = Router.prototype.push
Router.prototype.push = function push (location, onComplete, onAbort) {
  if (onComplete || onAbort) {
    return originalPush.call(this, location, onComplete, onAbort)
  }
  return originalPush.call(this, location)
    .catch(err => {
      if (err) {
        return Promise.reject(err)
      }
      // Keep to return the location which is same as cases without injecting query
      return Promise.resolve(location)
    })
}

const originalReplace = Router.prototype.replace
Router.prototype.replace = function replace (location, onComplete, onAbort) {
  if (onComplete || onAbort) {
    return originalReplace.call(this, location, onComplete, onAbort)
  }
  return originalReplace.call(this, location)
    .catch(err => {
      if (err) {
        return Promise.reject(err)
      }
      // Keep to return the location which is same as cases without injecting query
      return Promise.resolve(location)
    })
}

export default router
