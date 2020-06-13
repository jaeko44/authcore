<template>
  <modify-password
    :title="$t('change_password.title')"
    :button-text="$t('change_password.button.change_password')"
    :password.sync="password"
    :confirm-password.sync="confirmPassword"
    :loading="loading"
    :done="done"
    :password-error="passwordError"
    @submit.prevent="changePassword"
    @score="updatePasswordScore"
  />
</template>

<script>
import { mapState, mapActions } from 'vuex'

import successCallbackMixin from '@/mixins/successCallback'

import ModifyPassword from '@/components/ModifyPassword.vue'

export default {
  name: 'ChangePassword',
  mixins: [successCallbackMixin],
  components: {
    ModifyPassword
  },

  props: {
    callbackAction: {
      default: 'password_changed'
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('password', [
      'loading',
      'done',
      'error'
    ]),
    password: {
      get () {
        return this.$store.state.password.newPassword
      },
      set (value) {
        this.$store.commit('password/SET_NEW_PASSWORD', value)
      }
    },
    confirmPassword: {
      get () {
        return this.$store.state.password.confirmNewPassword
      },
      set (value) {
        this.$store.commit('password/SET_CONFIRM_NEW_PASSWORD', value)
      }
    },
    passwordScore: {
      get () {
        return this.$store.state.password.passswordScore
      },
      set (value) {
        this.$store.commit('password/SET_PASSWORD_SCORE', value)
      }
    },
    passwordError () {
      return this.error && this.error.newPassword
    }
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {
    this.$store.commit('password/RESET')
  },

  methods: {
    ...mapActions('password', [
      'changePassword'
    ]),
    updatePasswordScore (score) {
      this.passwordScore = score
    }
  }
}
</script>
