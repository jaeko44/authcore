<template>
  <div>
    <change-password-modal-pane
      :visible.sync="isChangePasswordModalVisible"
      :user="user"
      @updated="updatedRefresh"
    />
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row align-v="center" class="p:2.5rem">
          <b-col class="mr-auto">
            <h4 class="font-weight-bold">{{ $t('model.user.password') }}</h4>
          </b-col>
          <b-col class="text-right">
            <b-button
              class="px-4 btn-general"
              variant="outline-primary"
              :disabled="!canChangePassword"
              @click="isChangePasswordModalVisible = !isChangePasswordModalVisible"
            >
              {{ buttonWording }}
            </b-button>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { isRole } from '@/utils/permission'
import ChangePasswordModalPane from '@/components/modalpane/ChangePasswordModalPane.vue'

export default {
  name: 'PasswordPane',
  components: {
    ChangePasswordModalPane
  },
  props: {
    user: {
      type: Object,
      required: true
    }
  },

  data () {
    return {
      isChangePasswordModalVisible: false
    }
  },

  computed: {
    ...mapState('currentUser', {
      adminUser: 'user'
    }),

    buttonWording () {
      if (this.user.passwordAuthentication) {
        return this.$t('user_details_security.button.change_password')
      }
      return this.$t('user_details_security.button.set_password')
    },
    canChangePassword () {
      return isRole(this.adminUser, 'authcore.admin')
    }
  },

  methods: {
    // TODO: Decide behaviour for success updated
    updatedRefresh () {}
  }
}
</script>
