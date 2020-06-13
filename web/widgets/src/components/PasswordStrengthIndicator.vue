<template>
  <div>
    <svg width="100%" height="10px">
      <rect width="33%" height="10px" x="0%" style="fill-opacity: 0.2" v-bind:style="{ fill: strengthFirst }" />
      <rect width="33%" height="10px" x="33.5%" style="fill-opacity: 0.5" v-bind:style="{ fill: strengthSecond }" />
      <rect width="33%" height="10px" x="67%" v-bind:style="{ fill: strengthThird }" />
    </svg>
    <div class="text-left px-4 font-weight-bold">
      {{ $t('password_strength_indicator.title.password_strength') }} {{ $t(passwordStrength.title) }}
    </div>
    <div class="text-left px-4 small">
      {{ $t(passwordStrength.description) }}
    </div>
  </div>
</template>

<script>
import zxcvbn from 'zxcvbn'

export default {
  name: 'PasswordStrengthIndicator',
  components: {},

  props: {
    password: {
      type: String
    },
    disabledColour: {
      type: String,
      default: '#f2f2f2'
    },
    strengthTitle: {
      type: Array,
      default () {
        return [
          'password_strength_indicator.title.too_weak',
          'password_strength_indicator.title.ok',
          'password_strength_indicator.title.strong'
        ]
      }
    },
    strengthDescription: {
      type: Array,
      default () {
        return [
          'password_strength_indicator.description.too_weak',
          'password_strength_indicator.description.ok',
          'password_strength_indicator.description.strong'
        ]
      }
    }
  },

  data () {
    return {}
  },

  computed: {
    passwordStrength () {
      const score = zxcvbn(this.password).score
      let title = ''
      let description = ''
      if (score !== 0) {
        title = score > 3 ? this.strengthTitle[2] : this.strengthTitle[score - 1]
        description = score > 3 ? this.strengthDescription[2] : this.strengthDescription[score - 1]
      } else {
        // Only show message if the password is not null
        if (this.password !== '') {
          title = this.strengthTitle[0]
          description = this.strengthDescription[0]
        }
      }
      this.$emit('score', score)
      return {
        score: score,
        title: title,
        description: description
      }
    },
    strengthFirst () {
      if (this.password !== '') {
        return this.passwordStrength.score >= 0 ? getComputedStyle(document.documentElement).getPropertyValue('--primary') : this.disabledColour
      }
      return this.disabledColour
    },
    strengthSecond () {
      return this.passwordStrength.score >= 2 ? getComputedStyle(document.documentElement).getPropertyValue('--primary') : this.disabledColour
    },
    strengthThird () {
      return this.passwordStrength.score >= 3 ? getComputedStyle(document.documentElement).getPropertyValue('--primary') : this.disabledColour
    }
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {}
}
</script>
