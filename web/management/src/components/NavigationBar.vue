<template>
  <b-navbar fixed="top" class="header header-root background-white">
    <b-container class="navbar-container">
      <b-row align-v="center">
        <b-col cols="auto">
          <header-brand :to="{ name: 'ManagementHome' }">
            <header-brand-logo :src="require('@/assets/logo.svg')"></header-brand-logo>
          </header-brand>
        </b-col>
        <b-col v-if="managementEnabled" cols="auto" class="d-inline-flex">
          <b-button variant="link" class="text-decoration-none navigation-bar-link-color" :to="{ name: 'UserList' }">
            {{ $t('navigation.title.users') }}
          </b-button>
          <b-dropdown right :text="$t('navigation.title.settings')" variant="link" toggle-class="text-decoration-none navigation-bar-link-color">
            <b-dropdown-item :to="{ name: 'EmailTemplateList' }" link-class="navigation-bar-link-color">{{ $t('navigation.title.email_settings') }}</b-dropdown-item>
            <b-dropdown-item :to="{ name: 'SMSTemplateList' }">{{ $t('navigation.title.sms_settings') }}</b-dropdown-item>
          </b-dropdown>
        </b-col>
        <b-col v-if="isLoggedIn" cols="auto" class="ml-auto">
          <b-dropdown right variant="link" toggle-class="text-decoration-none py-0 d-flex align-items-center" no-caret>
            <template v-slot:button-content>
              <div class="navigation-monogram" :data-letter="firstCharProfileName"></div>
              <i class="ml-2 ac ac-down-arrow"></i>
            </template>
            <b-dropdown-item :to="{ name: 'Settings' }">{{ $t('navigation.title.user_portal') }}</b-dropdown-item>
          </b-dropdown>
        </b-col>
      </b-row>
    </b-container>
  </b-navbar>
</template>

<script>
import { mapState, mapGetters } from 'vuex'
import { isAdmin } from '@/utils/permission'

export default {
  name: 'NavigationBar',
  components: {},

  props: {},

  data () {
    return {}
  },

  computed: {
    ...mapState('management', [
      'ready'
    ]),
    ...mapState('currentUser', [
      'user'
    ]),
    ...mapGetters('currentUser', [
      'isLoggedIn',
      'firstCharProfileName'
    ]),
    managementEnabled () {
      return isAdmin(this.user)
    }
  },

  methods: {
    // For parent `sidebar-item` require this function to check the path to activate the visible and active states
    // Check if the sub item href includes the input result
    subIsActive (input) {
      const paths = Array.isArray(input) ? input : [input]
      return paths.some(path => {
        return this.$route.path.indexOf(path) === 0 // current path starts with this path string
      })
    }
  }
}
</script>

<style lang="scss">
.navbar-container {
    display: block !important;
    padding-left: 15px !important;
    padding-right: 15px !important;
}

.navigation-bar-link-color {
    color: $navigation-bar-link-color !important;

    &.active,
    &.router-link-active{
        color: $primary !important;
    }
}
</style>
