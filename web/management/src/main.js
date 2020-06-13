import Vue from 'vue'
import App from './App.vue'
import router from './router/index'
import store from './store'
import { i18n } from '@/i18n-setup'
import VueScrollTo from 'vue-scrollto'

// bootstrap-vue is not built as it does not publish
import BootstrapVue from 'bootstrap-vue/src/index/'
import '@/style.scss'
import '@/css/ac.css'

Vue.config.productionTip = false
Vue.use(BootstrapVue)
Vue.use(VueScrollTo)

Vue.directive('focus', {
  inserted (el) {
    if (el.classList.contains('bsq-form-field')) {
      setTimeout(() => {
        el.getElementsByTagName('input')[0].focus()
      }, 50)
    }
  }
})

new Vue({
  router,
  store,
  render: h => h(App),
  i18n
}).$mount('#app')
