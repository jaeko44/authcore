<template>
  <widget-template
    :back-enabled="backEnabled"
    :back-action="goBack"
    :title="title"
    :headerError="headerError"
  >
    <div :key="contentKey">
      <slot name="default"></slot>
      <b-row class="mt-4">
        <b-col class="text-center">
          <b-button
            v-if="button"
            block
            :variant="buttonVariant"
            @click="action"
          >
            <span>{{ buttonText }}</span>
          </b-button>
        </b-col>
      </b-row>
    </div>
  </widget-template>
</template>

<script>
import router from '@/router'
import { i18n } from '@/i18n-setup'

import WidgetTemplate from '@/components/WidgetTemplate.vue'

export default {
  name: 'ConfirmationTemplate',
  components: {
    WidgetTemplate
  },

  props: {
    title: {
      type: String,
      default: ''
    },
    backEnabled: {
      type: Boolean,
      default: false
    },
    button: {
      type: Boolean,
      default: true
    },
    buttonVariant: {
      type: String,
      default: ''
    },
    buttonText: {
      type: String,
      default: i18n.t('button.ok')
    },
    action: {
      type: Function,
      default: () => {}
    },
    backAction: {
      type: Function,
      default: null
    },
    headerError: {
      type: [String, Error],
      default: undefined
    }
  },

  data () {
    return {
      contentKey: ''
    }
  },

  computed: {},

  watch: {},

  created () {},
  mounted () {},
  updated () {
    // Update the div key to for animation transition from the slot.
    // Match with normal practice in Vue.
    const child = this.$slots.default.filter(child => child.tag)[0]
    this.contentKey = child.key
  },
  destroyed () {},

  methods: {
    goBack () {
      if (this.backAction === null) {
        router.go(-1)
      } else {
        this.backAction()
      }
    }
  }
}
</script>
