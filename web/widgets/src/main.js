import Vue from 'vue'
import amplitude from 'amplitude-js'

import BootstrapVue from 'bootstrap-vue/src/index'
import '@/style.scss'
import '@/css/ac.css'

import '@/dynamic-style.scss'

import { i18n } from '@/i18n-setup'

import App from './App.vue'
import router from './router'
import store from './store'

import { inMobile } from './utils/util.js'
import { formatDatetime } from './utils/format'

Vue.config.productionTip = false
Vue.use(BootstrapVue)

const fullPageWidgetsPostMessageException = [
  'AuthCore_updateHeight',
  'AuthCore_analytics',
  'AuthCore_unauthenticated',
  'AuthCore_unauthorized'
]

const analytics = amplitude.getInstance()

Vue.mixin({
  methods: {
    postMessage (type, data) {
      const parentOrigin = document.referrer
      const mergedData = data
      mergedData.containerId = store.state.preferences.containerId
      mergedData.clientId = store.state.preferences.clientId

      if (mergedData.clientId === '_authcore_admin_portal_') {
        return
      }

      if (!mergedData.containerId) {
        // For full page widgets (like callbacks from external OAuth, or a full-page OAuth signin
        // widget), we do not interact with the parent frame as it does not even exist. Hence, all
        // postMessage invoked (should be `updateHeight`) should not be called. As we expect only
        // `AuthCore_updateHeight` will be triggered, we are throwing all the other calls.
        // Excepted message types are in fullPageWidgetsPostMessageException array, they should be
        // return sliently instead of show error.

        // TODO: Filter out postMessage execution for AuthCore_analytics when containerId does not
        // exist. Refer to logAnalytics
        if (!fullPageWidgetsPostMessageException.includes(type)) {
          console.error(`type ${type} cannot postMessage via full page widgets`)
        }
        return
      }
      if (parentOrigin === '') {
        console.error('cannot find parent origin')
        return
      }
      const allowedHosts = store.state.client.widgetsSettings.appHosts
      const url = new URL(parentOrigin)
      if (!allowedHosts.includes(url.host)) {
        throw new Error('parent origin is not whitelisted')
      }
      window.parent.postMessage({
        type,
        data: mergedData
      }, parentOrigin)
    },
    logAnalytics (type, data, internal) {
      const widgetsSettings = store.state.client.widgetsSettings
      if (widgetsSettings) {
        const analyticsToken = widgetsSettings.analyticsToken
        // check analyticsToken set
        if (typeof analyticsToken === 'string' && analyticsToken !== '') {
          analytics.logEvent(type, data)
        }
      }
      if (!internal) {
        this.postMessage('AuthCore_analytics', { type, data })
      }
    },
    initAnalytics (analyticsToken) {
      if (typeof analyticsToken === 'string' && analyticsToken !== '') {
        analytics.init(analyticsToken)
      }
    }
  }
})

Vue.directive('focus', {
  inserted (el) {
    // Using the recommendation from https://developer.mozilla.org/en-US/docs/Web/HTTP/Browser_detection_using_the_user_agent for mobile checking
    if (el.classList.contains('bsq-form-field') && !inMobile()) {
      setTimeout(() => {
        el.getElementsByTagName('input')[0].focus()
      }, 50)
    }
  }
})

Vue.filter('formatDatetime', function (value) {
  return formatDatetime(value)
})

new Vue({
  router,
  store,
  render: h => h(App),
  i18n
}).$mount('#app')
