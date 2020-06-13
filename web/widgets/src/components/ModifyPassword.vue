<template>
  <widget-layout-v2
    :title="title"
    :description="description"
  >
    <b-row>
      <b-col>
        <b-form v-on:submit="$emit('submit', $event)">
          <b-form-group class="mb-0">
            <b-bsq-input
              password
              v-bind:value="password"
              v-on:input="$emit('update:password', $event)"
              :state="passwordError ? false : null"
              :disabled="loading || done"
              :label="$t(`modify_password.input.label.password`)"
              aria-describedby="passwordInvalidFeedback"
              autocomplete="new-password"
              type="password"
            />
            <b-form-invalid-feedback v-if="password === ''" id="passwordInvalidFeedback">
              {{ $t('general.blank') }}
            </b-form-invalid-feedback>
            <div v-else class="my-3">
              <password-strength-indicator
                :password="password"
                @score="updateScore"
              />
            </div>
          </b-form-group>
          <b-form-group class="mb-0">
            <b-bsq-input
              password
              v-bind:value="confirmPassword"
              v-on:input="$emit('update:confirmPassword', $event)"
              :state="passwordError ? false : null"
              :disabled="loading || done"
              :label="$t(`modify_password.input.label.confirm_password`)"
              aria-describedby="confirmPasswordInvalidFeedback"
              autocomplete="new-password"
              type="password"
            />
            <b-form-invalid-feedback id="confirmPasswordInvalidFeedback">
              {{ passwordError || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-form-group>
          <b-button
            block
            type="submit"
            variant="primary"
          >
            {{ buttonText }}
          </b-button>
        </b-form>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

export default {
  name: 'ModifyPassword',
  components: {
    WidgetLayoutV2,
    PasswordStrengthIndicator
  },
  props: {
    title: {
      type: String,
      required: true
    },
    description: {
      type: String,
      required: false
    },
    buttonText: {
      type: String,
      required: true
    },
    password: {
      type: String
    },
    confirmPassword: {
      type: String
    },
    loading: {
      type: Boolean,
      default: false
    },
    done: {
      type: Boolean,
      default: false
    },
    passwordError: {
      type: String
    }
  },

  methods: {
    updateScore (score) {
      this.$emit('score', score)
    }
  }
}
</script>
