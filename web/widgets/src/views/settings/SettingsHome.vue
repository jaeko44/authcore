<template>
  <widget-layout-v2
    :title="$t('settings_home.title')"
    :description="displayName"
    :back-button-enabled="false"
    @back-button="$router.push({ name: 'SettingsHome' })"
  >
    <div>
      <b-row>
        <b-col cols="12" class="px-0">
          <b-list-group>
            <hyperlink-list-item
              :to="{ name: 'ChangePassword' }"
              :title="$t('settings_home.list_item.title.password')"
            >
              {{ $t('settings_home.list_item.text.manage_password') }}
            </hyperlink-list-item>
            <hyperlink-list-item
              :title="$t('settings_home.list_item.title.two_step_verification')"
              :to="currentUser.is_password_set ? { name: 'MFAList' } : { name: 'ChangePassword' }"
            >
              <span v-if="currentUser.is_password_set">{{ $t('settings_home.list_item.text.two_step_verification.manage') }}</span>
              <span v-else>{{ $t('settings_home.list_item.text.two_step_verification.add') }}</span>
            </hyperlink-list-item>
            <hyperlink-list-item
              v-if="idpList"
              :title="$t('settings_home.list_item.title.social_logins')"
              :to="{ name: 'SocialLogins' }"
            >
              {{ $t('settings_home.list_item.text.social_logins') }}
            </hyperlink-list-item>
            <hyperlink-list-item
              :title="$t('settings_home.list_item.title.devices')"
              :to="{ name: 'Devices' }"
            >
              {{ $t('settings_home.list_item.text.devices') }}
            </hyperlink-list-item>
            <hyperlink-list-item
              v-if="adminLinkAvailable"
              :title="$t('settings_home.list_item.title.admin')"
              :href="adminLink"
            >
              {{ $t('settings_home.list_item.text.switch_to_admin') }}
            </hyperlink-list-item>
            <hyperlink-list-item
              v-if="isAdminPortal"
              :title="$t('settings_home.list_item.title.sign_out')"
              @click="signOut"
            >
            </hyperlink-list-item>
          </b-list-group>
        </b-col>
      </b-row>
    </div>
  </widget-layout-v2>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

import store from '@/store'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import HyperlinkListItem from '@/components/HyperlinkListItem.vue'

export default {
  name: 'SettingsHome',
  components: {
    WidgetLayoutV2,
    HyperlinkListItem
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('preferences', [
      'idpList'
    ]),
    ...mapState('users', [
      'currentUser',
      'displayName'
    ]),
    ...mapGetters('preferences', [
      'isAdminPortal'
    ]),
    adminLinkAvailable () {
      return this.isAdminPortal && this.currentUser.roles && this.currentUser.roles.some(item => (item.name === 'authcore.admin' || item.name === 'authcore.editor'))
    },
    adminLink () {
      const adminLink = new URL('/web/management', document.location)
      return adminLink.toString()
    }
  },

  methods: {
    async signOut () {
      await store.dispatch('authn/signOut')
      const url = new URL('/web/', window.location.origin)
      window.location.replace(url)
    }
  },

  mounted () {
    this.$store.dispatch('users/getCurrentUser')
  }
}
</script>
