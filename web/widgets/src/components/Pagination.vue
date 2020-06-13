<template>
  <div>
    <button id="previous" class="btn btn-sm mr-1" :class="buttonClass" @click="previousButton" :disabled="previousButtonEnabled">
      <i v-if="!previousButtonText" class="fas fa-chevron-left" />
      <span v-else>{{ previousButtonText }}</span>
    </button>
    <button id="next" class="btn btn-sm ml-1" :class="buttonClass" @click="nextButton" :disabled="nextButtonEnabled">
      <i v-if="!nextButtonText" class="fas fa-chevron-right" />
      <span v-else>{{ nextButtonText }}</span>
    </button>
  </div>
</template>

<script>
export default {
  name: 'Pagination',
  components: {},

  props: {
    previousPageToken: {
      type: String,
      default: ''
    },
    nextPageToken: {
      type: String,
      default: ''
    },
    previousButtonText: {
      type: String,
      default: ''
    },
    nextButtonText: {
      type: String,
      default: ''
    },
    totalItems: {
      type: Number,
      default: -1
    },
    buttonClass: {
      type: String
    }
  },

  computed: {
    previousButtonEnabled () {
      return !this.previousPageToken ? 'disabled' : false
    },
    nextButtonEnabled () {
      return !this.nextPageToken ? 'disabled' : false
    }
  },

  mounted () {},

  methods: {
    previousButton () {
      this.$emit('previous-click')
    },
    nextButton () {
      this.$emit('next-click')
    }
  }
}
</script>
