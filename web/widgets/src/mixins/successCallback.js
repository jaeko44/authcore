export default {
  watch: {
    done () {
      if (this.callbackAction !== undefined) {
        const callbackMessage = {
          action: this.callbackAction
        }
        if (this.user !== undefined) {
          callbackMessage.current_user = this.user
        }
        this.postMessage('AuthCore_onSuccess', callbackMessage)
      }
    }
  }
}
