<template>
  <div>
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col class="mr-auto">
            <h4 class="font-weight-bold">{{ $t('user_details_devices.title') }}</h4>
            <div class="text-grey">{{ $t('user_details_devices.description') }}</div>
          </b-col>
          <b-col class="text-right">
            <b-button
              v-b-tooltip.hover
              :title="$t('common.work_in_progress')"
              class="btn-general"
              variant="outline-danger"
            >
              {{ $t('user_details_devices.button.log_out_all') }}
            </b-button>
          </b-col>
        </b-row>

        <data-table
          class="mb-5"
          :loading="isLoadingStatus"
          :items="sessions"
          :fields="fields"
          :total-size="totalItems"
          :total-items-description="$t('common.total_items', { number: totalItems })"
          :previous-page-token="prevPageToken"
          :next-page-token="nextPageToken"
          @previous-click="previousPage"
          @next-click="nextPage"
        >
          <template v-slot="slotProps">
            <b-button
              variant="danger"
              class="btn-general"
              @click="setLogoutItem(slotProps.data.id)"
            >
              {{ $t('user_details_devices.button.log_out') }}
            </b-button>
          </template>
        </data-table>
      </b-col>
    </b-row>
    <logout-device-modal-pane
      :visible.sync="isModalEnabled"
      :logoutSession="logoutSession"
    />
  </div>
</template>

<script>
import { mapState } from 'vuex'

import DataTable from '@/components/DataTable.vue'
import LogoutDeviceModalPane from '@/components/modalpane/LogoutDeviceModalPane.vue'

import router from '@/router'

export default {
  name: 'UserDetailsDevices',
  components: {
    DataTable,
    LogoutDeviceModalPane
  },

  props: {
    id: {
      type: Number,
      required: true
    },
    pageToken: {
      type: String,
      default: ''
    }
  },

  data () {
    return {
      isLogoutDeviceModalEnabled: false,
      logoutSession: undefined
    }
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/userDetails', [
      'user'
    ]),
    ...mapState('management/userDetails/devicesList', [
      'sessions',
      'prevPageToken',
      'nextPageToken',
      'totalItems',

      'loading'
    ]),
    isLoadingStatus () {
      return !this.authenticated || this.loading
    },
    fields () {
      return [
        { key: 'userAgent', label: this.$t('model.device.user_agent'), class: 'pl:2.5rem' },
        { key: 'lastSeenAt', label: this.$t('model.device.last_seen_at') },
        { key: 'actions', label: this.$t('user_details_devices.text.log_out_device'), class: 'pr:2.5rem text-right' }
      ]
    },
    isModalEnabled: {
      get () {
        return this.isLogoutDeviceModalEnabled
      },
      set () {
        setTimeout(() => {
          this.isLogoutDeviceModalEnabled = false
        }, 500)
      }
    }
  },

  watch: {
    async user (newVal) {
      if (newVal !== undefined) {
        await this.list()
      }
    }
  },

  async created () {
    await this.list()
  },

  methods: {
    async list () {
      await this.$store.dispatch('management/userDetails/devicesList/list', {
        pageToken: this.pageToken
      })
    },
    async modalSubmit () {
      if (this.isLogoutDeviceModalEnabled) {
        await this.logoutDevice()
      }
      this.$nextTick(() => {
        this.isLogoutDeviceModalEnabled = false
      })
    },
    setLogoutItem (id) {
      this.isLogoutDeviceModalEnabled = !this.isLogoutDeviceModalEnabled
      this.logoutSession = this.sessions.find(item => item.id === id)
    },
    previousPage () {
      router.replace({ name: 'UserDetailsDevices', query: { pageToken: this.prevPageToken } })
    },
    nextPage () {
      router.replace({ name: 'UserDetailsDevices', query: { pageToken: this.nextPageToken } })
    }
  }
}
</script>
