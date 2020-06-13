import Vue from 'vue'
import Router from 'vue-router'

import routes from './routes'
import store from '@/store'
import { i18n } from '@/i18n-setup'

import { setRedirectPath } from '@/utils/util'

Vue.use(Router)

const router = new Router({
  routes,
  mode: 'history',
  // Match Boostrap-Vue active case for link
  linkExactActiveClass: 'active',
  base: '/web/'
})

router.beforeEach(async (routeTo, routeFrom, next) => {
  function redirectToSignIn () {
    // Pass the original route to the sign in component
    setRedirectPath(routeTo.fullPath)
    next({ name: 'SignIn' })
  }

  // Rename title in HTML meta tag
  if (routeTo.meta.title !== undefined) {
    document.title = routeTo.meta.title + ' - ' + i18n.t('common.product_name')
  } else {
    document.title = i18n.t('common.product_name')
  }

  const noAuthRequired = routeTo.matched.some((route) => route.meta.noAuthRequired)

  if (noAuthRequired) {
    return next()
  }

  if (store.getters['authn/isAuthenticated']) {
    store.commit('management/alert/POP_MESSAGE')
    return next()
  }

  // If auth is required and the user is NOT currently signed in,
  // redirect to sign in.
  redirectToSignIn()
})

export default router
