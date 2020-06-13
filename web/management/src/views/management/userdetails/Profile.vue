<template>
  <b-row no-gutters class="mt-5">
    <b-col cols="12" class="border rounded">
      <!-- Start of content part, padding needed to be set individually as to support table inside the content -->
      <b-row class="p:2.5rem">
        <b-col>
          <h4 class="font-weight-bold">{{ $t('user_details_profile.title') }}</h4>
          <div class="text-grey mb-4">{{ $t('user_details_profile.description') }}</div>
          <b-form @submit.prevent="update">
            <b-row class="mb-4">
              <b-col md="7">
                <b-bsq-input
                  separate
                  v-model="userProfile.name"
                  :label="$t('model.user.name')"
                />
              </b-col>
            </b-row>
            <b-row class="mb-4">
              <b-col md="7">
                <b-bsq-input
                  separate
                  v-model="userProfile.username"
                  :state="usernameErrorState"
                  aria-describedby="username"
                  :label="$t('model.user.username')"
                >
                  <b-form-invalid-feedback id="username">
                    {{ $t(usernameErrorContent) }}
                  </b-form-invalid-feedback>
                </b-bsq-input>
              </b-col>
            </b-row>
            <b-row class="mb-4">
              <b-col md="7">
                <b-form-contact-input
                  separate
                  v-model="userProfile.email"
                  :label="$t('model.user.email')"
                  autocorrect="off"
                  autocomplete="off"
                  autocapitalize="off"
                  :spellcheck="false"
                />
                <b-form-checkbox
                  v-model="userProfile.emailVerified"
                  switch
                  :disabled="userProfile.email === null || userProfile.email === ''"
                >
                  {{ $t('model.user.verified') }}
                </b-form-checkbox>
              </b-col>
            </b-row>
            <b-row class="mb-4">
              <b-col md="7">
                <b-form-contact-input
                  separate
                  v-model="userProfile.phoneNumber"
                  :label="$t('model.user.phone')"
                  autocorrect="off"
                  autocomplete="off"
                  autocapitalize="off"
                  :spellcheck="false"
                />
                <b-form-checkbox
                  v-model="userProfile.phoneNumberVerified"
                  switch
                  :disabled="userProfile.phoneNumber === null || userProfile.phoneNumber === ''"
                >
                  {{ $t('model.user.verified') }}
                </b-form-checkbox>
              </b-col>
            </b-row>
            <b-row class="mb-4">
              <b-col md="3">
                <b-button block class="btn-general" type="submit" variant="primary">
                  {{ $t('user_details_profile.button.save') }}
                </b-button>
              </b-col>
            </b-row>
          </b-form>
        </b-col>
      </b-row>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

export default {
  name: 'UserDetailsProfile',
  components: {},
  props: {
    user: {
      type: Object,
      required: true
    }
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/userDetails', [
      'loading'
    ]),
    ...mapState('management/userDetails/updateUserForm', [
      'error'
    ]),
    userProfile: {
      get () {
        return this.$store.state.management.userDetails.updateUserForm.user
      },
      set (value) {
        this.$store.commit('management/userDetails/updateUserForm/SET_USER', value)
      }
    },
    ...mapState('management/userDetails/updateUserForm', {
      updateProfileDone: state => state.done
    }),
    usernameErrorState () {
      return this.error.username === undefined ? null : false
    },
    usernameErrorContent () {
      return this.error.username === undefined ? 'common.blank' : this.error.username
    },
    isLoadingStatus () {
      return !this.authenticated || this.loading
    }
  },

  watch: {
    user () {
      this.$scrollTo('body')
      this.refresh()
    },
    userProfile: {
      handler (newVal) {
        if (newVal.email === '' || newVal.email === null) {
          this.userProfile.emailVerified = false
        }
        if (newVal.phoneNumber === '' || newVal.phoneNumber === null) {
          this.userProfile.phoneNumberVerified = false
        }
      },
      deep: true
    }
  },

  mounted () {
    this.refresh()
  },

  destroyed () {
    this.RESET_STATES()
  },

  methods: {
    ...mapMutations('management/userDetails/updateUserForm', [
      'RESET_STATES'
    ]),
    refresh () {
      this.userProfile = this.user
    },
    async update (e) {
      await this.$store.dispatch('management/userDetails/updateUserForm/update', this.userProfile)
    }
  }
}
</script>
