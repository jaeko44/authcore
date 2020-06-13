<template>
  <b-row>
    <b-col class="text-center">
      <i18n path="description.protected_by" tag="div" id="protected_by">
        <template #start>
          <span class="align-middle">Protected by</span>
        </template>
        <template #logo>
          <img alt="Authcore" src="@/assets/logo.svg" class="footer-img" />
        </template>
        <template #name>
          <span class="ml-1 align-middle footer-font authcore-name">Authcore</span>
        </template>
      </i18n>
    </b-col>
  </b-row>
</template>

<style lang="scss">
$footer-size: 16px;

.footer-font {
    font-size: $footer-size;
}
.footer-img {
    height: $footer-size;
}
</style>

<script>
export default {
  name: 'WidgetFooter'
}
</script>
