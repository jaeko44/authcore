<template>
  <standard-layout>
    <template #header>
      <navigation-bar ref="header" />
    </template>
    <router-view :style="{ 'margin-top': headerHeight + 'px' }" />
  </standard-layout>
</template>

<script>
import NavigationBar from '@/components/NavigationBar.vue'

export default {
  name: 'ManagementRouterView',
  components: {
    NavigationBar
  },
  props: {},

  data () {
    return {
      headerHeight: 0
    }
  },

  computed: {},

  mounted () {
    this.headerHeight = this.$refs.header.$el.clientHeight
  }
}
</script>
