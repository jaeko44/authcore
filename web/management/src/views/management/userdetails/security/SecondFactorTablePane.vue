<template>
  <div>
    <unlink-second-factor-modal-pane
        :visible.sync="isUnlinkSecondFactorModalVisible"
        :user="user"
        :secondFactor="secondFactorToBeUnlinked"
        @updated="updatedRefresh"
    />
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col cols="12">
            <h4 class="font-weight-bold">{{ $t('common.2fa') }}</h4>
            <div class="text-grey">
              <i18n v-if="firstlyCreatedFactor.createdAt" path="model.second_factor.created_at_with_date" tag="span">
                <span place="date">{{ firstlyCreatedFactor.createdAt }}</span>
              </i18n>
              <span class="d-inline-flex" :class="{ 'text-success': isSecondFactorEnabled, 'text-danger': !isSecondFactorEnabled }">
                <i v-if="isSecondFactorEnabled" class="mx-1 text-success ac-icon ac-shield-tick"></i>
                <i v-else class="mx-1 text-danger ac-icon ac-minus-circle"></i>
                {{ $t(secondFactorStatus) }}
              </span>
            </div>
          </b-col>
        </b-row>
        <data-table
          :items="secondFactors"
          :fields="secondFactorFields"
        >
          <template v-slot="slotProps">
            <b-button
              block
              class="btn-general"
              variant="outline-danger"
              @click="deleteMFA(slotProps.data)"
            >
              {{ $t('user_details_security.button.unlink') }}
            </b-button>
          </template>
        </data-table>
      </b-col>
    </b-row>
  </div>
</template>

<script>
import { mapState, mapGetters, mapActions } from 'vuex'

import DataTable from '@/components/DataTable.vue'
import UnlinkSecondFactorModalPane from '@/components/modalpane/UnlinkSecondFactorModalPane.vue'

export default {
  name: 'SecondFactorTablePane',
  components: {
    DataTable,
    UnlinkSecondFactorModalPane
  },
  props: {
    user: {
      type: Object,
      required: true
    }
  },

  data () {
    return {
      isUnlinkSecondFactorModalVisible: false
    }
  },

  computed: {
    ...mapState('management/userDetails/secondFactorsList', [
      'secondFactors'
    ]),
    ...mapGetters('management/userDetails/secondFactorsList', [
      'firstlyCreatedFactor'
    ]),
    secondFactorFields () {
      // Constant fields header for second factors table
      return [
        { key: 'name', label: this.$t('common.2fa_short'), class: 'pl:2.5rem' },
        { key: 'value', label: this.$t('model.second_factor.value') },
        { key: 'lastUsedAt', label: this.$t('model.second_factor.last_used_at') },
        { key: 'actions', label: '' }
      ]
    },
    isSecondFactorEnabled () {
      return this.secondFactors.length > 0
    },
    secondFactorStatus () {
      return this.isSecondFactorEnabled ? 'common.on' : 'common.off'
    },
    secondFactorToBeUnlinked: {
      get () {
        return this.$store.state.management.modalPane.unlinkSecondFactor.secondFactor
      },
      set (value) {
        this.$store.commit('management/modalPane/unlinkSecondFactor/SET_SECOND_FACTOR', value)
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
    ...mapActions('management/userDetails/secondFactorsList', [
      'fetchList'
    ]),
    refresh () {
      this.fetchList(this.user.id)
    },

    deleteMFA (data) {
      this.isUnlinkSecondFactorModalVisible = !this.isUnlinkSecondFactorModalVisible
      this.secondFactorToBeUnlinked = data
    },

    updatedRefresh () {
      this.isUnlinkSecondFactorModalVisible = false
      this.secondFactorToBeUnlinked = {}
      this.fetchList(this.user.id)
    }
  }
}
</script>

<style scoped lang="scss">
.ac-shield-tick,
.ac-minus-circle {
    font-size: 1.25rem;
    vertical-align: top;
    margin-top: -5px;
}
</style>
