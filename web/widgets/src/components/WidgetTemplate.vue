<template>
  <div>
    <b-row align-h="center">
      <b-col cols="12">
        <slot name="header">
          <b-row>
            <b-col cols="12" class="px-0">
              <b-alert
                :show="errorDisplay"
                variant="danger"
              >
                <span v-if="typeof headerError === 'object'">
                  {{ $t('error.unknown') }}
                </span>
                <span v-else>
                  {{ $t(headerError) }}
                </span>
              </b-alert>
            </b-col>
          </b-row>
          <b-row
            v-if="!internal"
            class="py-2 border-bottom border-grey-light"
          >
            <b-col>
              <img class="logo-size mr-3" alt="" :src="logo || require('@/assets/logo.svg')" />
              <span class="text-grey-medium align-middle">{{ company || "Authcore" }}</span>
            </b-col>
          </b-row>
          <div v-if="title !== ''">
            <b-row v-if="backEnabled" :class="{ 'border-bottom': !loading, 'border-grey-medium': !loading }" align-v="center">
              <b-col cols="auto" class="col-auto-with-back-button">
                <!-- Only apply py-2 on button to increase clickable zone for back button -->
                <b-button
                  variant="link"
                  @click="goBack"
                  class="border-0 p-2 text-decoration-none text-grey-dark font-size-inherit"
                >
                  <i class="ac-icon ac-left-arrow h5 my-0"></i>
                </b-button>
              </b-col>
              <b-col class="h5 text-center px-0 my-0">
                {{ title }}
              </b-col>
              <!-- Balancing for the center of the title -->
              <b-col cols="auto" class="pl-0">
                <b-button
                  variant="link"
                  @click="goBack"
                  class="border-0 p-2 text-decoration-none text-grey-dark invisible font-size-inherit"
                >
                  <i class="ac-icon ac-left-arrow h5 my-0"></i>
                </b-button>
              </b-col>
            </b-row>
            <b-row v-else :class="{ 'border-bottom': !loading, 'border-grey-medium': !loading }" class="py-2">
              <b-col class="h5 text-center px-0 my-0">
                {{ title }}
              </b-col>
            </b-row>
          </div>
          <b-row v-if="computedDescription" class="py-2" :class="{ 'mb-3': !sinkToContent, 'border-bottom': !sinkToContent, 'border-grey-light': !sinkToContent }">
            <b-col class="text-center">
              <slot name="description">{{ description }}</slot>
            </b-col>
          </b-row>
          <b-row v-else :class="{ 'mb-3': !sinkToContent }">
            <b-col></b-col>
          </b-row>
        </slot>
        <transition
          name="prompt-from-bottom"
          mode="out-in"
        >
          <slot name="default"></slot>
        </transition>
      </b-col>
    </b-row>
    <slot name="footer">
      <widget-footer
        v-if="!internal"
        class="mt-3"
      />
    </slot>
  </div>
</template>

<script>
import { mapState } from 'vuex'

import router from '@/router'

import WidgetFooter from '@/components/WidgetFooter.vue'

export default {
  name: 'WidgetTemplate',
  components: {
    WidgetFooter
  },

  props: {
    title: {
      type: String,
      default: ''
    },
    description: {
      type: String,
      default: ''
    },
    sinkToContent: {
      type: Boolean,
      default: false
    },
    logoEnabled: {
      type: Boolean,
      default: false
    },
    backEnabled: {
      type: Boolean,
      default: false
    },
    loading: {
      type: Boolean,
      default: false
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
    return {}
  },

  computed: {
    ...mapState('client', [
      'logo',
      'company',
      'internal'
    ]),
    // Use computedDescription to decide whether description styling is required
    computedDescription () {
      return this.$slots.description !== undefined || this.description !== ''
    },
    errorDisplay () {
      const errorAllowed = [
        'error.unknown',
        'sign_in.description.error.used_contact_in_system',
        'manage_social_logins.description.error.used_contact_in_system'
      ]
      if (typeof this.headerError === 'string') {
        const errorShown = errorAllowed.includes(this.headerError)
        if (!errorShown) {
          console.warn(`The error is not allowed to be shown in header, the error key is ${this.headerError}.`)
        }
        return errorShown
      }
      return this.headerError !== undefined
    }
  },

  mounted () {},
  updated () {},

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
