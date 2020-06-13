<!-- WithLoadingContent provides Facebook-like loading content in bar format.
     Mainly use in table cell, content provides in default slot is the content
     after loading is completed. -->
<template>
  <div v-if="loading" class="ph-item">
    <div class="ph-row">
      <div class="ph-col-12"></div>
    </div>
  </div>
  <div v-else>
    <slot name="default"></slot>
  </div>
</template>

<script>
export default {
  name: 'WithLoadingContent',
  components: {},

  props: {
    loading: {
      type: Boolean,
      default: false
    },
    row: {
      type: Object
    }
  }
}
</script>
