<template>
  <general-layout>
    <loading-spinner v-if="!authenticated" />
    <b-container v-else>
      <b-row>
        <b-col>
          <h2>{{ $t('title.logs') }}</h2>
        </b-col>
      </b-row>
      <log-table
        include-user
        :loading="loading"
        :log-data="logs"
        :total-items="totalItems"
        :rows-per-page="10"
        :previous-page-token="prevPageToken"
        :next-page-token="nextPageToken"
        @previous-click="previousPage"
        @next-click="nextPage"
      />
    </b-container>
  </general-layout>
</template>

<script>
import { mapState } from 'vuex'

import router from '@/router'
import store from '@/store'

import GeneralLayout from '@/components/GeneralLayout.vue'
import LogTable from '@/components/LogTable.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

export default {
  name: 'LogList',
  components: {
    GeneralLayout,
    LogTable,
    LoadingSpinner
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/logsList', [
      'logs',
      'prevPageToken',
      'nextPageToken',
      'totalItems',

      'loading',
      'error'
    ])
  },

  watch: {
    async authenticated (newVal) {
      if (newVal) {
        // Clear the error message from unauthenticated state
        store.commit('management/logsList/CLEAR_STATES')
        await store.dispatch('management/logsList/list', {
          pageToken: this.pageToken
        })
      }
    }
  },

  mounted () {
    store.dispatch('management/logsList/list', {})
  },
  destroyed () {
    store.commit('management/logsList/CLEAR_STATES')
  },

  methods: {
    previousPage () {
      router.push({ name: 'LogList', query: { pageToken: this.prevPageToken } })
    },
    nextPage () {
      router.push({ name: 'LogList', query: { pageToken: this.nextPageToken } })
    }
  },

  beforeRouteUpdate (to, from, next) {
    store.dispatch('management/logsList/list', {
      pageToken: to.query.pageToken
    })
    next()
  }
}
</script>
