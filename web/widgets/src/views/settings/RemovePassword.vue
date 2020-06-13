<template>
  <confirmation-layout
    :title="$t('remove_password.title')"
    :description="$t('remove_password.description')"
    :button-text="$t('remove_password.button.remove')"
    @click="removePassword"
  />
</template>

<script>
import ConfirmationLayout from '@/components/ConfirmationLayout.vue'

export default {
  name: 'RemovePassword',
  components: {
    ConfirmationLayout
  },
  props: {},

  data () {
    return {}
  },

  computed: {},

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {
    removePassword () {
      console.log('test')
    }
  }
}
</script>
