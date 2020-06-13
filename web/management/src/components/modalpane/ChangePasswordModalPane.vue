<template>
  <b-bsq-modal
    v-bind:visible="visible"
    v-on:change="$emit('update:visible', $event)"
    :header-title="headerTitle"
    :button-title="buttonTitle"
    button-variant="primary"
    @ok="modalSubmit"
  >
    <div class="mt-3 mb-5">
      <i18n :path="descriptionKey" tag="div" class="mb-3">
        <span class="font-weight-bold" place="username">{{ user.profileName }}</span>
      </i18n>
      <b-bsq-input
        v-model="password"
        :disabled="loading"
        :label="$t('model.user.password')"
        type="password"
        class="my-3" />
      <b-bsq-input
        v-model="confirmPassword"
        :disabled="loading"
        :label="$t('model.user.confirm_password')"
        type="password"
        class="my-3" />
    </div>
  </b-bsq-modal>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'ChangePasswordModalPane',
  model: {
    prop: 'visible',
    event: 'change'
  },
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    user: {
      type: Object,
      required: true
    }
  },

  data () {
    return {
      password: '',
      confirmPassword: ''
    }
  },

  computed: {
    ...mapState('management/modalPane/changePassword', [
      'loading'
    ]),
    headerTitle () {
      if (this.user.passwordAuthentication) {
        return this.$t('change_password_modal_pane.title.change_password')
      }
      return this.$t('change_password_modal_pane.title.set_password')
    },
    buttonTitle () {
      if (this.user.passwordAuthentication) {
        return this.$t('change_password_modal_pane.button.change_password')
      }
      return this.$t('change_password_modal_pane.button.set_password')
    },
    descriptionKey () {
      if (this.user.passwordAuthentication) {
        return 'change_password_modal_pane.description.change_password'
      }
      return 'change_password_modal_pane.description.set_password'
    }
  },

  destoryed () {
    this.refresh()
  },

  methods: {
    ...mapActions('management/modalPane/changePassword', [
      'changePassword'
    ]),
    refresh () {
      this.password = ''
      this.confirmPassword = ''
    },
    async modalSubmit () {
      await this.changePassword({
        password: this.password,
        confirmPassword: this.confirmPassword
      })
      this.refresh()
      // Dismiss the modal when the action finished
      this.$emit('update:visible', false)
    }
  }
}
</script>
