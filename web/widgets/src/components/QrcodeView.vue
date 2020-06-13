<template>
  <canvas></canvas>
</template>

<script>
import QRCode from 'qrcode'

export default {
  props: {
    content: String,
    width: {
      type: Number,
      default: 150
    },
    errorCorrectionLevel: {
      type: String,
      default: 'Q'
    }
  },
  mounted () {
    QRCode.toCanvas(this.$el, this.content, {
      width: this.width,
      errorCorrectionLevel: this.errorCorrectionLevel
    })
  },

  watch: {
    content: function () {
      QRCode.toCanvas(this.$el, this.content, {
        width: this.width,
        errorCorrectionLevel: this.errorCorrectionLevel
      })
    }
  }
}
</script>
