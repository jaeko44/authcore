<template>
  <widget-layout-v2
    :title="$t('manage_authenticator_app.title')"
    :description="$t('manage_authenticator_app.description')"
    @back-button="$router.push({ name: 'MFAList' })"
  >
    <div>
      <b-row class="mb-3">
        <b-col>
          <successful-show-up exclude-wording no-animation />
        </b-col>
      </b-row>
      <b-row>
        <b-col>
          <b-button
            block
            variant="danger"
            :to="{ name: 'RemoveAuthenticatorApp', params: { factorId: factorId } }"
          >
            {{ $t('manage_authenticator_app.button.remove') }}
          </b-button>
        </b-col>
      </b-row>
    </div>
  </widget-layout-v2>
</template>

<script>
import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import SuccessfulShowUp from '@/components/SuccessfulShowUp.vue'

export default {
  name: 'ManageAuthenticatorApp',
  components: {
    WidgetLayoutV2,
    SuccessfulShowUp
  },
  props: {
    factorId: {
      type: Number,
      required: true
    }
  }
}
</script>
