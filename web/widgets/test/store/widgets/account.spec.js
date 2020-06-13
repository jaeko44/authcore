import _ from 'lodash'
import { assert } from 'chai'

import account from '@/store/widgets/account'

const {
  SET_PROFILE_COMPLETED,
  GET_PROFILE_STARTED,
  GET_PROFILE_COMPLETED,
  GET_PROFILE_FAILED
} = account.mutations
const { state: defaultState } = account

suite('/store/widgets/account.js', () => {
  suite('mutations', () => {
    let state
    beforeEach(() => {
      state = _.cloneDeep(defaultState)
    })
    test('should have correct default values', () => {
      assert.isObject(state.user)
      assert.isFalse(state.loading)
      assert.isFalse(state.done)
      assert.isUndefined(state.error)
    })
    suite('GET_PROFILE_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        GET_PROFILE_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.done)
      })
    })
    suite('GET_PROFILE_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        GET_PROFILE_COMPLETED(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isTrue(state.done)
        assert.isUndefined(state.error)
      })
    })
    suite('SET_PROFILE_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        const user = {
          profileName: 'profileName',
          displayName: 'displayName',
          username: 'username',
          primary_email: 'user@example.com',
          primary_phone: '+85298765432'
        }
        SET_PROFILE_COMPLETED(state, user)
        // Testing
        assert.isObject(state.user)
        assert.equal(state.user.handle, 'username')
      })
      test('should use primary email for handle', () => {
        // Preparing
        const user = {
          profileName: 'profileName',
          primary_email: 'user@example.com',
          primary_phone: '+85298765432'
        }
        SET_PROFILE_COMPLETED(state, user)
        // Testing
        assert.isObject(state.user)
        assert.equal(state.user.handle, 'user@example.com')
      })
      test('should use primary phone for handle', () => {
        // Preparing
        const user = {
          profileName: 'profileName',
          primary_phone: '+85298765432'
        }
        SET_PROFILE_COMPLETED(state, user)
        // Testing
        assert.isObject(state.user)
        assert.equal(state.user.handle, '+85298765432')
      })
    })
    suite('GET_PROFILE_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        GET_PROFILE_FAILED(state, new Error('an imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
      })
    })
  })
})
