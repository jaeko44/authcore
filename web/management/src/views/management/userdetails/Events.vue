<template>
  <b-row
    no-gutters
    class="mt-5"
  >
    <b-col cols="12" class="border rounded">
      <b-row class="p:2.5rem">
        <b-col>
          <h4 class="font-weight-bold">{{ $t('user_details_events.title') }}</h4>
          <div class="text-grey">{{ $t('user_details_events.description') }}</div>
        </b-col>
      </b-row>
      <data-table
        class="mb-5"
        :loading="loading"
        :items="logs"
        :fields="fields"
        :total-size="totalItems"
        :total-items-description="$tc('common.total_items', totalItems)"
        :previous-page-token="prevPageToken"
        :next-page-token="nextPageToken"
        @first-page-click="firstPage"
        @previous-click="previousPage"
        @next-click="nextPage"
      />
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import { isEmpty } from 'lodash'

import router from '@/router'

import DataTable from '@/components/DataTable.vue'

export default {
  name: 'UserDetailsEvents',
  components: {
    DataTable
  },

  props: {
    queryObject: {
      type: Object
    }
  },

  data () {
    return {
      fields: [
        { key: 'action', label: this.$t('user_details_events.table.header.action'), class: 'pl:2.5rem' },
        { key: 'result', label: this.$t('user_details_events.table.header.status') },
        { key: 'ip', label: this.$t('user_details_events.table.header.ip') },
        { key: 'device', label: this.$t('user_details_events.table.header.device') },
        { key: 'createdAt', label: this.$t('user_details_events.table.header.when') }
      ]
    }
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/userDetails/logsList', [
      'loading',
      'error',

      'logs',
      'totalItems',
      'prevPageToken',
      'nextPageToken'
    ]),
    isLoadingStatus () {
      return !this.authenticated || this.loading
    },
    normalizedQueryObject () {
      // Return sort key and query key from query string, fallback to default if not exists
      const result = {
        userId: this.queryObject.id,
        pageToken: this.queryObject.pageToken || ''
      }
      return result
    }
  },

  watch: {
    normalizedQueryObject (newVal) {
      this.fetchList(newVal)
    }
  },

  mounted () {
    this.fetchList(this.normalizedQueryObject)
  },

  methods: {
    ...mapActions('management/userDetails/logsList', [
      'fetchList'
    ]),

    firstPage () {
      if (!isEmpty(router.currentRoute.query)) {
        router.replace({
          name: 'UserDetailsEvents'
        })
      }
    },
    previousPage () {
      router.push({
        name: 'UserDetailsEvents',
        query: {
          pageToken: this.prevPageToken
        }
      })
    },
    nextPage () {
      router.push({
        name: 'UserDetailsEvents',
        query: {
          pageToken: this.nextPageToken
        }
      })
    }
  }
}
</script>
