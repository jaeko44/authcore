<template>
  <b-container fluid class="py-4">
    <b-row align-h="center">
      <b-col cols="12">
        <slot name="header">
          <large-icon-header
            v-if="largeHeader"
            :logo="logo"
            :company="company"
            :title="title"
            :description="description"
            :logo-enabled="logoEnabled"
            :company-enabled="companyEnabled"
            :back-button-enabled="backButtonEnabled"
            @back-button="onBackButton"
          />
          <small-icon-header
            v-else
            :logo="logo"
            :company="company"
            :title="title"
            :description="description"
            :logo-enabled="logoEnabled"
            :company-enabled="companyEnabled"
            :back-button-enabled="backButtonEnabled"
            @back-button="onBackButton"
          >
            <template #description>
              <slot name="description"></slot>
            </template>
          </small-icon-header>
        </slot>
        <transition
          name="prompt-from-bottom"
          mode="out-in"
        >
          <slot name="default"></slot>
        </transition>
      </b-col>
    </b-row>
    <widget-footer
      v-if="!internal"
      class="mt-5"
    />
    <slot name="footer">
    </slot>
  </b-container>
</template>

<script>
import { mapState } from 'vuex'

import router from '@/router'

import LargeIconHeader from '@/components/layout/LargeIconHeader.vue'
import SmallIconHeader from '@/components/layout/SmallIconHeader.vue'
import WidgetFooter from '@/components/WidgetFooter.vue'

export default {
  name: 'WidgetLayoutV2',
  components: {
    LargeIconHeader,
    SmallIconHeader,
    WidgetFooter
  },

  props: {
    largeHeader: {
      type: Boolean,
      default: false
    },
    title: {
      type: String,
      default: ''
    },
    description: {
      type: String,
      default: ''
    },
    backButtonEnabled: {
      type: Boolean,
      default: true
    },
    alert: {
      type: String,
      default: ''
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('preferences', [
      'logo',
      'company',
      'internal'
    ]),
    // TODO: Set enable into a props
    logoEnabled () {
      return !!this.logo
    },
    companyEnabled () {
      return !!this.company
    }
  },

  methods: {
    onBackButton () {
      if (this.$listeners['back-button']) {
        this.$emit('back-button')
      } else {
        router.go(-1)
      }
    }
  }
}
</script>
