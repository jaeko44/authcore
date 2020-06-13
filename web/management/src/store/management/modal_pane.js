import changePassword from './modal_pane/change_password'
import lockUser from './modal_pane/lock_user'
import deleteUser from './modal_pane/delete_user'
import unlinkOauthFactor from './modal_pane/unlink_oauth_factor'
import unlinkSecondFactor from './modal_pane/unlink_second_factor'

export default {
  namespaced: true,
  modules: {
    changePassword,
    lockUser,
    deleteUser,
    unlinkOauthFactor,
    unlinkSecondFactor
  }
}
