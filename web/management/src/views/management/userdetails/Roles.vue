<template>
  <div>
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col class="mr-auto">
            <h4 class="font-weight-bold">{{ $t('user_details_roles.title') }}</h4>
            <div class="text-grey">{{ $t('user_details_roles.description') }}</div>
          </b-col>
          <b-col class="text-right">
            <b-dropdown right variant="outline-primary" :text="$t('user_details_roles.text.assign_role')" toggle-class="btn-block btn-general">
              <b-dropdown-item-button
                v-for="item in roles"
                :key="item.name"
                @click="assignRole(item.id)"
              >
                {{ item.name }}
              </b-dropdown-item-button>
            </b-dropdown>
          </b-col>
        </b-row>

        <data-table
          :items="roleAssignments"
          :fields="roleFields"
          :empty-text="$t('user_details_roles.text.no_roles_assigned')"
        >
          <template v-slot="slotProps">
            <b-button
              class="btn-general"
              variant="outline-danger"
              @click="unassignRole(slotProps.data.id)"
            >
              {{ $t('user_details_roles.button.unassign') }}
            </b-button>
          </template>
        </data-table>
      </b-col>
    </b-row>
  </div>
</template>

<script>
import { mapState, mapMutations, mapActions } from 'vuex'

import DataTable from '@/components/DataTable.vue'

export default {
  name: 'UserDetailsRoles',
  components: {
    DataTable
  },
  props: {
    user: {
      type: Object,
      required: true
    }
  },

  computed: {
    ...mapState('management/rolesList', [
      'roles'
    ]),
    ...mapState('management/userDetails/role', [
      'roleAssignments'
    ]),
    roleFields () {
      return [
        { key: 'name', label: this.$t('model.role.name'), class: 'pl:2.5rem' },
        { key: 'description', label: this.$t('model.role.description') },
        { key: 'actions', label: '', class: 'pr:2.5rem text-right' }
      ]
    }
  },

  watch: {
    user () {
      this.refresh()
    }
  },

  mounted () {
    this.refresh()
  },
  destroyed () {
    this.RESET_STATES()
  },

  methods: {
    ...mapMutations('management/userDetails/role', [
      'RESET_STATES'
    ]),
    ...mapActions('management/userDetails/role', [
      'fetchRoleAssignmentsList',
      'assignRole',
      'unassignRole'
    ]),
    refresh () {
      this.fetchRoleAssignmentsList(this.user.id)
    }
  }
}
</script>
