<template>
  <widget-layout-v2
    :title="$t('devices.title')"
    :description="user.profileName"
    :headerError="error"
    @back-button="$router.push({ name: 'SettingsHome' })"
  >
    <div>
      <b-row>
        <b-col cols="12" class="px-0">
          <b-list-group class="mb-3">
            <b-list-group-item
              class="generic-list-item"
              v-for="session in sessions"
              :key="session.id"
            >
              <b-row align-v="center">
                <b-col>
                  <div>
                    <div class="font-weight-bold">{{ session.user_agent }}</div>
                    <div>
                      <span
                        v-if="session.id === currentSession.id"
                        class="text-success"
                      >
                        {{ $t('devices.list_item.text.this_device') }}
                      </span>
                      <i18n
                        v-else
                        path="devices.list_item.text.last_active"
                        tag="span"
                      >
                        <template #date>
                          <span>{{ session.last_seen_at | formatDatetime }}</span>
                        </template>
                      </i18n>
                    </div>
                  </div>
                </b-col>
                <b-col
                  class="text-right"
                  cols="auto"
                >
                  <b-button
                    v-if="session.id !== currentSession.id"
                    :to="{ name: 'DeviceDelete', params: { deviceId: session.id } }"
                    class="px-md-4 hover-remove"
                    variant="outline-danger"
                  >
                    {{ $t('devices.button.log_out') }}
                  </b-button>
                </b-col>
              </b-row>
            </b-list-group-item>
          </b-list-group>
        </b-col>
      </b-row>
      <b-button
        block
        :to="{ name: 'DeviceDelete', params: { deviceId: 0 } }"
        variant="outline-danger"
      >
        <!-- TODO: Change into corresponding i18n string when MR !140 is done -->
        {{ $t('devices.button.log_out_all_other_devices') }}
      </b-button>
    </div>
  </widget-layout-v2>
</template>

<script>
import { mapState, mapGetters } from 'vuex'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'

export default {
  name: 'Devices',
  components: {
    WidgetLayoutV2
  },

  computed: {
    ...mapState('client', [
      'authcoreClient'
    ]),
    ...mapState('widgets/account', [
      'user'
    ]),
    ...mapGetters('client', [
      'accessToken'
    ]),
    ...mapState('devices', [
      'sessions',
      'currentSession',

      'error'
    ])
  },

  async mounted () {
    await this.listDevices()
  },

  methods: {
    async listDevices () {
      await this.$store.dispatch('devices/list')
    }
  }
}
</script>
