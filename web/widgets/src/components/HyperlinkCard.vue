<template>
  <b-card :no-body="noBody" class="mb-3">
    <div class="">
      <b-link v-if="direction === 'bottom'" :to="to" class="stretched-link" @click="clickEvent"></b-link>
      <h5 v-if="title !== null">{{ title }}</h5>
      <slot name="custom-description">
        <b-row align-v="center">
          <b-col :cols="direction === 'right' ? 7 : 12" :sm="direction === 'right' ? 9 : 12">
            <div class="mb-2">
              <slot name="description"></slot>
            </div>
          </b-col>
          <b-col :cols="direction === 'right' ? 5 : 12" :sm="direction === 'right' ? 3 : 12">
            <b-button v-if="direction === 'right'" block class="text-danger text-nowrap" variant="link" :to="to" @click="clickEvent">
              {{ linkWording }}
            </b-button>
            <div v-else class="text-primary">
              {{ linkWording }}
            </div>
          </b-col>
        </b-row>
      </slot>
    </div>
  </b-card>
</template>

<script>
export default {
  name: 'HyperlinkCard',

  props: {
    to: {
      type: [String, Object],
      default: null
    },
    title: {
      type: String,
      default: null
    },
    linkWording: {
      type: String,
      default: null
    },
    direction: {
      type: String,
      default: 'bottom'
    },
    noBody: {
      type: Boolean,
      default: false
    }
  },

  computed: {},

  methods: {
    clickEvent () {
      this.$emit('click')
    }
  }
}
</script>
