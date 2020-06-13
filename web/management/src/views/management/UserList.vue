<template>
  <general-layout>
    <loading-spinner v-if="!authenticated" />
    <div v-else>
      <change-password-modal-pane
        :visible.sync="isChangePasswordModalVisible"
        :user="user"
      />
      <lock-user-modal-pane
        :visible.sync="isLockUserModalVisible"
        :user="user"
      />
      <b-container class="mt-5">
        <b-row align-v="center" class="mb-5">
          <b-col class="mr-auto">
            <h1 class="font-weight-bold">{{ $t('user_list.title') }}</h1>
          </b-col>
          <b-col class="text-right" cols="3">
            <b-button
              block
              class="btn-general"
              variant="primary"
              @click="showCreateUserView"
            >
              {{ $t('user_list.button.create_user') }}
            </b-button>
          </b-col>
        </b-row>
        <b-row>
          <b-col cols="12">
            <filter-bar
              class="mb-5"
              :query-object="normalizedQueryObject"
              :query-placeholder="queryPlaceholder"
              :query-keys="queryKeys"
              :sort-keys="sortKeys"
              @submit="submitQuery"
            />
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <data-table
              theme-full
              :loading="loading"
              :items="formattedUsers"
              :fields="fields"
              :total-size="totalItems"
              :total-items-description="$tc('common.total_users', totalItems)"
              :previous-page-token="prevPageToken"
              :next-page-token="nextPageToken"
              @row-clicked="viewItem"
              @first-page-click="firstPage"
              @previous-click="previousPage"
              @next-click="nextPage"
            />
          </b-col>
        </b-row>
      </b-container>
    </div>
  </general-layout>
</template>

<script>
import { mapState } from 'vuex'
import { isEmpty } from 'lodash'

import router from '@/router'

import { formatToPhoneString, formatDatetime } from '@/utils/format'

import GeneralLayout from '@/components/GeneralLayout.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import DataTable from '@/components/DataTable.vue'
import FilterBar from '@/components/FilterBar.vue'

import ChangePasswordModalPane from '@/components/modalpane/ChangePasswordModalPane.vue'
import LockUserModalPane from '@/components/modalpane/LockUserModalPane.vue'

export default {
  name: 'UserList',
  components: {
    GeneralLayout,
    LoadingSpinner,
    DataTable,
    FilterBar,

    ChangePasswordModalPane,
    LockUserModalPane
  },

  props: {
    queryObject: {
      type: Object
    }
  },

  data () {
    return {
      isChangePasswordModalVisible: false,
      isLockUserModalVisible: false
    }
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/usersList', [
      'users',
      'prevPageToken',
      'nextPageToken',
      'totalItems',

      'loading',
      'error'
    ]),
    ...mapState('management/userDetails', [
      'user'
    ]),
    // Constant state for the page, should not be changed except codebase changes.
    fields () {
      return [
        { key: 'name', label: this.$t('model.user.name'), class: 'pl-3' },
        { key: 'email', label: this.$t('model.user.email') },
        { key: 'phoneNumber', label: this.$t('model.user.phone') },
        { key: 'lastSeenAt', label: this.$t('model.user.last_seen_at') }
      ]
    },
    queryPlaceholder () {
      return this.$t('user_list.text.search_for_user')
    },
    queryKeys () {
      return [
        { key: 'search', name: this.$t('user_list.text.all'), default: true },
        { key: 'email', name: this.$t('model.user.email') },
        { key: 'phoneNumber', name: this.$t('model.user.phone') }
      ]
    },
    sortKeys () {
      return [
        { key: 'is_locked', name: this.$t('model.user.locked_user') },
        { key: 'created_at', name: this.$t('model.user.created_at') },
        { key: 'last_seen_at', name: this.$t('model.user.last_seen_at'), default: true },
        { key: 'email', name: this.$t('model.user.email') }
      ]
    },
    userLocked () {
      return (this.user && this.user.locked)
    },
    formattedUsers () {
      return this.users.map(item => {
        item.phoneNumber = formatToPhoneString(item.phoneNumber)
        item.lastSeenAt = item.lastSeenAt !== '1970-01-01T00:00:01Z' ? formatDatetime(item.lastSeenAt) : this.$t('model.user.no_last_seen_at')
        return item
      })
    },
    normalizedQueryObject () {
      // Check value is "true" in query string or return true from component
      const ascending = this.queryObject.ascending === 'true' || this.queryObject.ascending === true || false
      // Return sort key and query key from query string, fallback to default if not exists
      let sortKey
      this.sortKeys.forEach(item => {
        if (item.default) {
          sortKey = item.key
        }
        // Selected value has higher priority over default value
        if (item.key === this.queryObject.sortKey) {
          sortKey = item.key
        }
      })
      let queryKey
      this.queryKeys.forEach(item => {
        if (item.default) {
          queryKey = item.key
        }
        // Selected value has higher priority over default value
        if (item.key === this.queryObject.queryKey) {
          queryKey = item.key
        }
      })
      const result = {
        pageToken: this.queryObject.pageToken || '',
        sortKey: sortKey,
        ascending: ascending,
        queryKey: queryKey,
        queryValue: this.queryObject.queryValue || ''
      }
      return result
    }
  },

  watch: {
    normalizedQueryObject () {
      this.fetchList()
    }
  },

  mounted () {
    this.fetchList()
  },
  destroyed () {
    this.$store.commit('management/usersList/RESET_STATES')
  },

  methods: {
    firstPage () {
      if (!isEmpty(router.currentRoute.query)) {
        router.push({
          name: 'UserList'
        })
      }
    },
    previousPage () {
      router.push({
        name: 'UserList',
        query: {
          pageToken: this.prevPageToken,
          ascending: this.normalizedQueryObject.ascending,
          sort: this.normalizedQueryObject.sortKey,
          queryKey: this.normalizedQueryObject.queryKey,
          queryValue: this.normalizedQueryObject.queryValue
        }
      })
    },
    nextPage () {
      router.push({
        name: 'UserList',
        query: {
          pageToken: this.nextPageToken,
          ascending: this.normalizedQueryObject.ascending,
          sort: this.normalizedQueryObject.sortKey,
          queryKey: this.normalizedQueryObject.queryKey,
          queryValue: this.normalizedQueryObject.queryValue
        }
      })
    },
    submitQuery (data) {
      router.push({
        name: 'UserList',
        query: {
          pageToken: '',
          ascending: data.ascending,
          sort: data.sortKey,
          queryKey: data.queryKey,
          queryValue: data.queryValue
        }
      })
    },

    async viewItem (item) {
      router.push({ name: 'UserDetails', params: { id: item.id } })
    },
    showCreateUserView () {
      router.push({
        name: 'UserCreate'
      })
    },

    async triggerChangePasswordAction (data) {
      await this.$store.dispatch('management/userDetails/get', data.id)
      this.isChangePasswordModalVisible = true
    },
    async triggerLockUserAction (data) {
      await this.$store.dispatch('management/userDetails/get', data.id)
      this.isLockUserModalVisible = true
    },

    fetchList () {
      this.$store.dispatch('management/usersList/fetchList', this.normalizedQueryObject)
    }
  }
}
</script>
