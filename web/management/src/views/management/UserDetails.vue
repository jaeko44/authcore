<template>
  <general-layout
    show-bar
    :back-title="$t('user_list.title')"
    :back-action="routeBack"
  >
    <div>
      <b-container class="mt-5">
        <b-row class="mb-5">
          <b-col>
            <h1 class="my-0 font-weight-bold">{{ $t('user_details.title') }}</h1>
          </b-col>
        </b-row>
        <b-row align-v="center" class="mb-5">
          <b-col class="mr-auto">
            <b-row align-v="center">
              <!-- TODO: profile icon - ref: https://gitlab.com/blocksq/authcore/issues/69 -->
              <b-col cols="auto" class="details-monogram" :data-letter="firstCharProfileName"></b-col>
              <b-col cols="auto" class="mt-3">
                <b-row class="h5 my-0 font-weight-light font:1.375rem">
                  {{ user.profileName }}
                </b-row>
                <b-row class="small text-danger">
                  {{ $t(locked) }}
                </b-row>
              </b-col>
            </b-row>
          </b-col>
          <b-col class="text-right" cols="3">
            <b-dropdown right class="w-100" variant="outline-primary" :text="$t('user_details.button.actions')" toggle-class="btn-block btn-general">
              <b-dropdown-item-button
                @click="isLockUserModalVisible = !isLockUserModalVisible"
              >
                <lock-user-modal-pane
                  :visible.sync="isLockUserModalVisible"
                  :user="user"
                />
                <div v-if="user.locked" class="text-muted">
                  {{ $t('user_details.actions.unlock_user') }}
                </div>
                <div v-else class="text-muted">
                  {{ $t('user_details.actions.lock_user') }}
                </div>
              </b-dropdown-item-button>
              <b-dropdown-item-button
                @click="isDeleteUserModalVisible = !isDeleteUserModalVisible"
              >
                <delete-user-modal-pane
                  :visible.sync="isDeleteUserModalVisible"
                  :user="user"
                  @updated="routeBack"
                />
                <span class="text-muted">
                  {{ $t('user_details.actions.delete_user') }}
                </span>
              </b-dropdown-item-button>
            </b-dropdown>
          </b-col>
        </b-row>
        <b-row>
          <b-col cols="12">
            <navbar>
              <navbar-item
                :to="{ name: 'UserDetailsProfile' }"
              >
                {{ $t('user_details.sub_title.profile') }}
              </navbar-item>
              <navbar-item
                :to="{ name: 'UserDetailsEvents' }"
              >
                {{ $t('user_details.sub_title.events') }}
              </navbar-item>
              <navbar-item
                :to="{ name: 'UserDetailsSecurity' }"
              >
                {{ $t('user_details.sub_title.security') }}
              </navbar-item>
              <navbar-item
                :to="{ name: 'UserDetailsDevices' }"
              >
                {{ $t('user_details.sub_title.devices') }}
              </navbar-item>
              <navbar-item
                :to="{ name: 'UserDetailsRoles' }"
              >
                {{ $t('user_details.sub_title.roles') }}
              </navbar-item>
              <navbar-item
                :to="{ name: 'UserDetailsMetadata' }"
              >
                {{ $t('user_details.sub_title.metadata') }}
              </navbar-item>
            </navbar>
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <!-- View for the navbar -->
            <router-view
              :user="user"
            />
          </b-col>
        </b-row>
      </b-container>
    </div>
  </general-layout>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

import LockUserModalPane from '@/components/modalpane/LockUserModalPane.vue'
import DeleteUserModalPane from '@/components/modalpane/DeleteUserModalPane.vue'
import GeneralLayout from '@/components/GeneralLayout.vue'
import Navbar from '@/components/Navbar.vue'
import NavbarItem from '@/components/NavbarItem.vue'

export default {
  name: 'UserDetails',
  components: {
    LockUserModalPane,
    DeleteUserModalPane,
    GeneralLayout,
    Navbar,
    NavbarItem
  },

  props: {
    id: {
      type: Number,
      required: true
    }
  },

  data () {
    return {
      isLockUserModalVisible: false,
      isDeleteUserModalVisible: false
    }
  },

  computed: {
    ...mapState('management/userDetails', [
      'user',
      'loading',
      'error'
    ]),
    firstCharProfileName () {
      if (this.user.profileName) {
        return this.user.profileName[0].toUpperCase()
      }
      return ''
    },
    locked () {
      return this.user.locked ? 'model.user.locked' : 'common.blank'
    }
  },

  async mounted () {
    this.$store.dispatch('management/userDetails/get', this.id)
  },

  destroyed () {
    this.RESET_STATES()
  },

  methods: {
    ...mapMutations('management/userDetails', [
      'RESET_STATES'
    ]),
    routeBack () {
      this.$router.push({
        name: 'UserList'
      })
    }
  }
}
</script>
