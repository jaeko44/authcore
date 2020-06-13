<template>
  <!-- text-center class for align two images into center -->
  <div>
    <div class="text-center">
      <svg
        width="66"
        height="66"
        viewBox="0 0 66 66"
      >
        <g
          class="svg-stroke-primary"
          fill="none"
          stroke-width="3"
        >
          <circle
            cx="33"
            cy="33"
            r="33"
            stroke="none"
          />
          <circle
            cx="33"
            cy="33"
            r="31.5"
            fill="none"
          />
        </g>
      </svg>
      <div class="h-0">
        <svg
          width="39.796"
          height="39.796"
          viewBox="0 0 39.796 39.796"
          v-bind:class="{ tick: true, 'tick-shown': tickShown }"
        >
          <path
            class="svg-stroke-primary"
            d="M0,49.967H17.567l-.911-32.711"
            transform="translate(37.454 -10.08) rotate(45)"
            fill="none"
            stroke="#1100fa"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="3"
          />
        </svg>
      </div>
    </div>
    <h6
      v-if="!excludeWording"
      class="text-center mt-3"
    >
      {{ wording }}
    </h6>
  </div>
</template>

<script>
/** SuccessfulShowUp serves to represents successful status for action.
 * The component aligns into center by default. To change the horizontal alignment for the component, set the parent DOM element with `display: flex` and control the alignment using `justify-content` styling.
 **/
import { i18n } from '@/i18n-setup'

export default {
  name: 'SuccessfulShowUp',

  props: {
    wording: {
      type: String,
      default: i18n.t('description.successful')
    },
    excludeWording: {
      type: Boolean,
      default: false
    },
    noAnimation: {
      type: Boolean,
      default: false
    }
  },

  data () {
    return {
      tickShown: false
    }
  },

  created () {
    this.showTick()
  },

  methods: {
    showTick () {
      if (this.noAnimation) {
        this.tickShown = true
      }
      setTimeout(() => {
        this.tickShown = true
      }, 1000)
    }
  }
}
</script>

<style>
.h-0 {
    height: 0;
}
</style>
