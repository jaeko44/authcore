import _ from 'lodash'
import { assert } from 'chai'

import updateProfileForm from '@/store/widgets/account/update_profile_form'

const {
  UPDATE_PROFILE_STARTED,
  UPDATE_PROFILE_COMPLETED,
  UPDATE_PROFILE_FAILED
} = updateProfileForm.mutations
const { state: defaultState } = updateProfileForm

suite('/store/widgets/account/update_profile_form.js', () => {
  suite('mutations', () => {
    let state
    beforeEach(() => {
      state = _.cloneDeep(defaultState)
    })
    test('should have correct default values', () => {
      assert.isFalse(state.loading)
      assert.isFalse(state.done)
      assert.isUndefined(state.error)
    })
    suite('UPDATE_PROFILE_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        UPDATE_PROFILE_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.done)
        assert.isUndefined(state.error)
      })
    })
    suite('UPDATE_PROFILE_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        UPDATE_PROFILE_COMPLETED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isTrue(state.done)
        assert.isUndefined(state.error)
      })
    })
    suite('UPDATE_PROFILE_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        UPDATE_PROFILE_FAILED(state, new Error('imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
      })
      test('not updated username', () => {
        // Preparing
        UPDATE_PROFILE_FAILED(state, {
          response: {
            status: 400,
            statusText: 'Bad Request',
            obj: {
              code: 3,
              error: 'invalid argument',
              message: 'invalid argument'
            }
          }
        })
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
        assert.equal(state.error, 'profile_edit.input.error.not_updated_username')
      })
      test('unknown error', () => {
        // Preparing
        UPDATE_PROFILE_FAILED(state, {
          response: {
            status: 500,
            statusText: 'Internal Server Error',
            obj: {
              code: 2,
              error: 'incorrect code',
              message: 'incorrect code'
            }
          }
        })
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
        assert.equal(state.error, 'error.unknown')
      })
    })
  })
})
