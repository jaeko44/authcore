<template>
  <div>
    <unlink-oauth-factor-modal-pane
      :visible.sync="isUnlinkOauthFactorModalVisible"
      :user="user"
      :oauthFactor="oAuthFactorToBeUnlinked"
      @updated="updatedRefresh"
    />
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col>
            <h4 class="font-weight-bold">{{ $t('common.social_login') }}</h4>
            <div class="text-grey">
              <i18n v-if="firstlyCreatedFactor.createdAt" path="model.oauth_factor.created_at_with_date" tag="span">
                <span place="date">{{ firstlyCreatedFactor.createdAt }}</span>
              </i18n>
            </div>
          </b-col>
        </b-row>
        <data-table
          :items="oAuthFactors"
          :fields="oAuthFactorFields"
          :empty-text="$t('user_details_security.text.no_social_account_linked')"
        >
          <template v-slot="slotProps">
            <b-button
              v-if="oAuthFactorsUnlinkAvailable"
              block
              class="btn-general"
              variant="outline-danger"
              @click="unlinkOAuthFactor(slotProps.data)"
            >
              {{ $t('user_details_security.button.unlink') }}
            </b-button>
            <div v-else>{{ $t('user_details_security.text.unlink_unavailable') }}</div>
          </template>
        </data-table>
      </b-col>
    </b-row>
  </div>
</template>

<script>
import { mapState, mapGetters, mapActions } from 'vuex'

import DataTable from '@/components/DataTable.vue'
import UnlinkOauthFactorModalPane from '@/components/modalpane/UnlinkOauthFactorModalPane.vue'

export default {
  name: 'OauthFactorTablePane',
  components: {
    DataTable,
    UnlinkOauthFactorModalPane
  },
  props: {
    user: {
      type: Object,
      required: true
    }
  },

  data () {
    return {
      isUnlinkOauthFactorModalVisible: false
    }
  },

  computed: {
    ...mapState('management/userDetails/oauthFactorsList', [
      'oAuthFactors'
    ]),
    ...mapGetters('management/userDetails/oauthFactorsList', [
      'firstlyCreatedFactor'
    ]),
    oAuthFactorFields () {
      // Constant fields header for OAuth factors table
      return [
        { key: 'service', label: this.$t('model.oauth_factor.social_media'), class: 'pl:2.5rem' },
        { key: 'oauthUserId', label: this.$t('model.oauth_factor.platform_user_id') },
        { key: 'createdAt', label: this.$t('model.oauth_factor.created_at') },
        { key: 'lastUsedAt', label: this.$t('model.oauth_factor.last_used_at') },
        { key: 'actions', label: '' }
      ]
    },
    oAuthFactorsUnlinkAvailable () {
      return this.user.passwordAuthentication || this.oAuthFactors.length > 1
    },
    oAuthFactorToBeUnlinked: {
      get () {
        return this.$store.state.management.modalPane.unlinkOauthFactor.oauthFactor
      },
      set (value) {
        this.$store.commit('management/modalPane/unlinkOauthFactor/SET_OAUTH_FACTOR', value)
      }
    }
  },

  watch: {
    user () {
      this.refresh()
    }
  },

  mounted () {
    if (this.user) {
      this.refresh()
    }
  },

  methods: {
    ...mapActions('management/userDetails/oauthFactorsList', [
      'fetchList'
    ]),
    refresh () {
      this.fetchList(this.user.id)
    },
    unlinkOAuthFactor (data) {
      this.isUnlinkOauthFactorModalVisible = !this.isUnlinkOauthFactorModalVisible
      this.oAuthFactorToBeUnlinked = data
    },

    updatedRefresh () {
      this.isUnlinkOauthFactorModalVisible = false
      this.oAuthFactorToBeUnlinked = {}
      this.fetchList(this.user.id)
    }
  }
}
</script>
