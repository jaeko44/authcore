import _ from 'lodash'
import { assert } from 'chai'

import resetPasswordForm from '@/store/widgets/account/reset_password_form'

const {
  AUTHENTICATE_HANDLE_STARTED,
  AUTHENTICATE_HANDLE_COMPLETED,
  AUTHENTICATE_HANDLE_FAILED,
  AUTHENTICATE_FACTOR_INIT,
  AUTHENTICATE_FACTOR_STARTED,
  AUTHENTICATE_FACTOR_COMPLETED,
  AUTHENTICATE_FACTOR_FAILED,
  RESET_PASSWORD_STARTED,
  RESET_PASSWORD_COMPLETED,
  RESET_PASSWORD_FAILED
} = resetPasswordForm.mutations
const { state: defaultState } = resetPasswordForm

suite('/store/widgets/account/reset_password_form.js', () => {
  suite('mutations', () => {
    let state
    beforeEach(() => {
      state = _.cloneDeep(defaultState)
    })
    test('should have correct default values', () => {
      assert.isArray(state.challenges)
      assert.isUndefined(state.selectedChallengeMethod)
      assert.isUndefined(state.authorizationToken)
      assert.isFalse(state.authenticateHandleDone)
      assert.isFalse(state.authenticateFactorDone)
      assert.isFalse(state.resetPasswordDone)
      assert.isFalse(state.loading)
      assert.isUndefined(state.error)
    })
    suite('AUTHENTICATE_HANDLE_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_HANDLE_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.authenticateHandleDone)
        assert.isUndefined(state.error)
      })
    })
    suite('AUTHENTICATE_HANDLE_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_HANDLE_COMPLETED(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isTrue(state.authenticateHandleDone)
        assert.isUndefined(state.error)
      })
    })
    suite('AUTHENTICATE_HANDLE_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_HANDLE_FAILED(state, new Error('imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
      })
    })
    suite('AUTHENTICATE_FACTOR_INIT', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_FACTOR_INIT(state, [
          'CHALLENGE'
        ])
        // Testing
        assert.isFalse(state.loading)
        assert.isArray(state.challenges)
        assert.equal(state.selectedChallengeMethod, 'CHALLENGE')
      })
    })
    suite('AUTHENTICATE_FACTOR_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_FACTOR_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.authenticateFactorDone)
        assert.isUndefined(state.error)
      })
    })
    suite('AUTHENTICATE_FACTOR_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_FACTOR_COMPLETED(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isTrue(state.authenticateFactorDone)
      })
    })
    suite('AUTHENTICATE_FACTOR_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        AUTHENTICATE_FACTOR_FAILED(state, new Error('an imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.isFalse(state.authenticateFactorDone)
        assert.exists(state.error)
      })
    })
    suite('RESET_PASSWORD_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        RESET_PASSWORD_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.resetPasswordDone)
        assert.isUndefined(state.error)
      })
    })
    suite('RESET_PASSWORD_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        RESET_PASSWORD_COMPLETED(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isTrue(state.resetPasswordDone)
        assert.isUndefined(state.error)
      })
    })
    suite('RESET_PASSWORD_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        RESET_PASSWORD_FAILED(state, new Error('an imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.isFalse(state.resetPasswordDone)
        assert.exists(state.error)
      })
    })
  })
})
