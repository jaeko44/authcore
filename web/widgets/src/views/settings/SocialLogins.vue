<template>
  <widget-layout-v2
    :title="$t('manage_social_logins.title')"
    :description="user.profileName"
    :headerError="error"
    @back-button="$router.push({ name: 'SettingsHome' })"
  >
    <b-row>
      <b-col class="px-0">
        <b-list-group>
          <social-login-list-item
            v-for="(item, index) in socialLoginList"
            :key="index"
            :factor="oauthFactors.find(oauthFactor => oauthFactor.service === item)"
            :service="item"
          />
        </b-list-group>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import SocialLoginListItem from '@/components/SocialLoginListItem.vue'

export default {
  name: 'SocialLogins',
  components: {
    WidgetLayoutV2,
    SocialLoginListItem
  },

  computed: {
    ...mapState('widgets/account', [
      'user'
    ]),
    ...mapState('socialLogin', [
      'loading',
      'done',
      'error',
      'oauthFactors'
    ]),
    ...mapState('widgets/externalOauth', [
      'oauthState'
    ]),
    ...mapState('preferences', {
      socialLoginList: 'idpList'
    })
  },

  mounted () {
    store.dispatch('socialLogin/list')
  }
}
</script>
