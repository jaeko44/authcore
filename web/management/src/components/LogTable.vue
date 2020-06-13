<template>
  <div id="log-table">
    <b-table
      class="theme-full"
      cellspacing="0"
      :items="logData"
      :fields="computedFields"
    >
      <template v-if="includeUser" v-slot:cell(username)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(action)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(result)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(ip)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(device)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(createdAt)="data">
        <with-loading-content :loading="loading">
          {{ data.value }}
        </with-loading-content>
      </template>
    </b-table>
    <b-row>
      <b-col cols="6"></b-col>
      <b-col cols="6" class="text-right">
        <pagination
          class="list-with-table-padding-right"
          :previous-page-token="previousPageToken"
          :next-page-token="nextPageToken"
          @previous-click="previousPage"
          @next-click="nextPage"
        />
      </b-col>
    </b-row>
  </div>
</template>

<script>
import WithLoadingContent from '@/components/WithLoadingContent.vue'
import Pagination from '@/components/Pagination.vue'

export default {
  name: 'LogTable',
  components: {
    WithLoadingContent,
    Pagination
  },
  props: {
    logData: {
      type: Array,
      default: () => []
    },
    totalItems: {
      type: Number,
      default: 0
    },
    rowsPerPage: {
      type: Number,
      default: 0
    },
    includeUser: {
      type: Boolean,
      default: false
    },
    loading: {
      type: Boolean,
      default: false
    },
    previousPageToken: {
      type: String
    },
    nextPageToken: {
      type: String
    }
  },

  data () {
    return {
      fields: [
        { key: 'action', label: 'Event' },
        { key: 'result', label: 'Status' },
        { key: 'ip', label: 'IP' },
        { key: 'device', label: 'Device' },
        { key: 'createdAt', label: 'When' }
      ]
    }
  },

  computed: {
    computedFields: function () {
      const fields = this.fields
      if (this.includeUser) {
        fields.unshift({ key: 'username', label: 'User' })
      }
      return fields
    }
  },

  methods: {
    previousPage () {
      this.$emit('previous-click')
    },
    nextPage () {
      this.$emit('next-click')
    }
  }
}
</script>
