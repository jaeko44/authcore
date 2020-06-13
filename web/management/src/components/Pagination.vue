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
    buttonClass: {
      type: String,
      default: 'btn-pagination'
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

<style scoped lang="scss">
.btn-pagination {
    color: $bsq-text-grey;
    &:hover {
        color: $bsq-text-grey;
        background-color: $bsq-grey-light;
    }
    &:disabled {
        background-color: transparent;
        border-color: transparent;
        &:hover {
            color: $btn-link-disabled-color;
        }
    }
}
</style>
