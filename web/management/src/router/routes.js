import Home from '@/views/Home.vue'

import SignIn from '@/views/SignIn.vue'
import BasicRouterView from '@/views/BasicRouterView.vue'
import ManagementRouterView from '@/views/management/ManagementRouterView.vue'

import Settings from '@/views/settings/Settings.vue'

import UserList from '@/views/management/UserList.vue'
import UserDetails from '@/views/management/UserDetails.vue'
import UserDetailsProfile from '@/views/management/userdetails/Profile.vue'
import UserDetailsEvents from '@/views/management/userdetails/Events.vue'
import UserDetailsSecurity from '@/views/management/userdetails/Security.vue'
import UserDetailsDevices from '@/views/management/userdetails/Devices.vue'
import UserDetailsRoles from '@/views/management/userdetails/Roles.vue'
import UserDetailsMetadata from '@/views/management/userdetails/Metadata.vue'
import UserCreate from '@/views/management/UserCreate.vue'
import EmailTemplateList from '@/views/management/EmailTemplateList.vue'
import EmailTemplateEdit from '@/views/management/EmailTemplateEdit.vue'
import SMSTemplateList from '@/views/management/SMSTemplateList.vue'
import SMSTemplateEdit from '@/views/management/SMSTemplateEdit.vue'

import Unauthorized from '@/views/error/Unauthorized.vue'

import { i18n } from '@/i18n-setup'
import store from '@/store'

export default [{
  path: '/',
  name: 'Home',
  component: Home,
  meta: {
    noAuthRequired: true
  },
  props: true
}, {
  path: '/sign-in',
  name: 'SignIn',
  component: SignIn,
  meta: {
    noAuthRequired: true
  },
  beforeEnter: (routeTo, routeFrom, next) => {
    if (store.getters['authn/isAuthenticated']) {
      next({ name: 'Settings' })
    }
    next()
  }
}, {
  path: '/settings',
  component: BasicRouterView,
  children: [{
    path: '',
    redirect: { name: 'Settings' }
  }, {
    path: 'settings',
    name: 'Settings',
    component: Settings,
    meta: {
      title: i18n.t('settings.meta.title')
    }
  }]
}, {
  path: '/management',
  component: ManagementRouterView,
  children: [{
    path: '',
    name: 'ManagementHome',
    redirect: {
      name: 'UserList'
    }
  }, {
    path: 'users',
    name: 'UserList',
    component: UserList,
    meta: {
      title: i18n.t('user_list.meta.title')
    },
    props (route) {
      const queryObject = {
        pageToken: route.query.pageToken,
        sortKey: route.query.sort,
        ascending: route.query.ascending,
        queryKey: route.query.queryKey,
        queryValue: route.query.queryValue
      }
      return {
        queryObject: queryObject
      }
    }
  }, {
    path: 'users/:id(\\d+)',
    component: UserDetails,
    props (route) {
      return {
        id: parseInt(route.params.id, 10)
      }
    },
    children: [{
      path: '/',
      name: 'UserDetails',
      redirect: {
        name: 'UserDetailsProfile'
      }
    }, {
      path: 'profile',
      name: 'UserDetailsProfile',
      component: UserDetailsProfile,
      meta: {
        title: i18n.t('user_details_profile.meta.title')
      },
      props (route) {
        return {
          id: parseInt(route.params.id, 10)
        }
      }
    }, {
      path: 'events',
      name: 'UserDetailsEvents',
      component: UserDetailsEvents,
      meta: {
        title: i18n.t('user_details_events.meta.title')
      },
      props (route) {
        const queryObject = {
          id: parseInt(route.params.id, 10),
          pageToken: route.query.pageToken
        }
        return {
          queryObject: queryObject
        }
      }
    }, {
      path: 'security',
      name: 'UserDetailsSecurity',
      component: UserDetailsSecurity,
      meta: {
        title: i18n.t('user_details_security.meta.title')
      },
      props (route) {
        return {
          id: parseInt(route.params.id, 10)
        }
      }
    }, {
      path: 'devices',
      name: 'UserDetailsDevices',
      component: UserDetailsDevices,
      meta: {
        title: i18n.t('user_details_devices.meta.title')
      },
      props (route) {
        return {
          id: parseInt(route.params.id, 10)
        }
      }
    }, {
      path: 'roles',
      name: 'UserDetailsRoles',
      component: UserDetailsRoles,
      meta: {
        title: i18n.t('user_details_roles.meta.title')
      }
    }, {
      path: 'metadata',
      name: 'UserDetailsMetadata',
      component: UserDetailsMetadata,
      meta: {
        title: i18n.t('user_details_metadata.meta.title')
      },
      props (route) {
        return {
          id: parseInt(route.params.id, 10)
        }
      }
    }]
  }, {
    path: 'users/new',
    name: 'UserCreate',
    component: UserCreate,
    meta: {
      title: i18n.t('user_create.meta.title')
    }
  }, {
    path: 'settings/email/template',
    name: 'EmailTemplateList',
    component: EmailTemplateList,
    meta: {
      title: i18n.t('email_template_list.meta.title')
    }
  }, {
    path: 'settings/email/template/settings',
    name: 'EmailTemplateEdit',
    component: EmailTemplateEdit,
    props (route) {
      return {
        templateName: route.params.templateName,
        language: route.params.language
      }
    }
  }, {
    path: 'settings/sms/template',
    name: 'SMSTemplateList',
    component: SMSTemplateList,
    meta: {
      title: i18n.t('sms_template_list.meta.title')
    }
  }, {
    path: 'settings/sms/template/settings',
    name: 'SMSTemplateEdit',
    component: SMSTemplateEdit,
    props (route) {
      return {
        templateName: route.params.templateName,
        language: route.params.language
      }
    }
  }]
}, {
  path: '/401',
  name: 'Unauthorized',
  component: Unauthorized
}]
