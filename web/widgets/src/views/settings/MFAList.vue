<template>
  <widget-layout-v2
    :title="$t('mfa_list.title')"
    :description="$t('mfa_list.description')"
    :headerError="error"
    @back-button="$router.push({ name: 'SettingsHome' })"
  >
    <b-row>
      <b-col class="px-0">
        <b-list-group>
          <hyperlink-list-item
            :to="totpFactor ? { name: 'ManageAuthenticatorApp', params: { factorId: totpFactor.id } } : { name: 'CreateAuthenticatorApp' }"
            :title="$t('mfa_list.list_item.title.authenticator_app')"
          >
            <span v-if="totpFactor">{{ $t('mfa_list.list_item.text.authenticator_app.manage') }}</span>
            <span v-else>{{ $t('mfa_list.list_item.text.authenticator_app.create') }}</span>
          </hyperlink-list-item>
        </b-list-group>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import { mapState } from 'vuex'

import store from '@/store'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import HyperlinkListItem from '@/components/HyperlinkListItem.vue'

export default {
  name: 'MFAList',
  components: {
    WidgetLayoutV2,
    HyperlinkListItem
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('mfa', [
      'secondFactors',
      'error'
    ]),
    totpFactor () {
      return this.secondFactors.find(secondFactor => secondFactor.type === 'totp')
    }
  },

  async mounted () {
    await store.dispatch('mfa/list')
  },

  methods: {}
}
</script>
