import moment from 'moment'
import { parsePhoneNumberFromString } from 'libphonenumber-js'

import { formatDatetime } from './format'

function parsePhoneToString (phone) {
  // TODO: Hardcoded for Hong Kong, ref: https://gitlab.com/blocksq/authcore/issues/14
  if (phone !== undefined) {
    const phoneNumber = parsePhoneNumberFromString(phone, 'HK')
    if (phoneNumber !== undefined) {
      return phoneNumber.formatInternational()
    }
  }
  return ''
}

export function marshalDatetime (datetime, format) {
  return moment(datetime).format(format)
}

export function marshalContact (contact) {
  function getType (type) {
    switch (type) {
      case undefined: return 'email'
      case 'EMAIL': return 'email'
      case 'PHONE': return 'phone'
      default: throw new Error(`undefined type ${type}`)
    }
  }
  let valueString = ''
  if (getType(contact.type) === 'phone') {
    valueString = parsePhoneToString(contact.value)
  } else {
    valueString = contact.value
  }
  return {
    id: parseInt(contact.id, 10),
    type: getType(contact.type),
    value: valueString,
    isPrimary: contact.primary,
    isVerified: contact.verified || false
  }
}

export function marshalUser (user) {
  let profileName, handle
  if (typeof user.display_name === 'string' && user.display_name !== '') {
    profileName = user.display_name
  } else if (typeof user.username === 'string' && user.username !== '') {
    profileName = user.username
  } else if (typeof user.primary_email === 'string' && user.primary_email !== '') {
    profileName = user.primary_email
  } else if (typeof user.primary_phone === 'string' && user.primary_phone !== '') {
    profileName = user.primary_phone
  } else {
    // Should be impossible state
    profileName = 'No data shown'
  }
  if (typeof user.username === 'string' && user.username !== '') {
    handle = user.username
  } else if (typeof user.primary_email === 'string' && user.primary_email !== '') {
    handle = user.primary_email
  } else if (typeof user.primary_phone === 'string' && user.primary_phone !== '') {
    handle = user.primary_phone
  } else {
    // Should be impossible state
    handle = 'No handle'
  }
  const primaryEmailVerifiedStatus = user.primary_email_verified !== '0001-01-01T00:00:00Z'
  const primaryPhoneVerifiedStatus = user.primary_phone_verified !== '0001-01-01T00:00:00Z'

  return {
    handle: handle,
    profileName: profileName,
    displayName: user.display_name,
    username: user.username,
    primaryEmail: user.primary_email,
    primaryEmailVerified: user.primary_email_verified,
    primaryEmailVerifiedStatus: primaryEmailVerifiedStatus,
    primaryPhone: parsePhoneToString(user.primary_phone),
    primaryPhoneVerified: user.primary_phone_verified,
    primaryPhoneVerifiedStatus: primaryPhoneVerifiedStatus,
    smsAuthentication: user.sms_authentication || false,
    totpAuthentication: user.totp_authentication || false,
    passwordAuthentication: user.password_authentication || false,
    language: user.language || ''
  }
}

export function marshalDevice (device) {
  let lastSeenAt = ''
  let expiredAt = ''
  if (device.last_seen_at !== '0001-01-01T00:00:00Z') {
    lastSeenAt = marshalDatetime(device.last_seen_at, 'HH:mm:ss, DD MMM Y')
  }
  if (device.expired_at !== '0001-01-01T00:00:00Z') {
    expiredAt = marshalDatetime(device.expired_at, 'HH:mm:ss, DD MMM Y')
  }
  return {
    id: parseInt(device.id, 10),
    userAgent: device.user_agent,
    lastSeenAt: lastSeenAt,
    lastSeenIp: device.last_seen_ip,
    lastSeenLocation: device.last_seen_location,
    expiredAt: expiredAt,
    isCurrent: device.is_current || false
  }
}

export function marshalSecondFactor (secondFactor) {
  return {
    id: parseInt(secondFactor.id, 10),
    userId: parseInt(secondFactor.user_id, 10),
    type: secondFactor.type || 'SMS',
    content: {
      identifier: secondFactor.content.identifier,
      phoneNumber: parsePhoneToString(secondFactor.content.phone_number),
      used: secondFactor.content.used
    },
    createdAt: formatDatetime(secondFactor.created_at),
    lastUsedAt: secondFactor.last_used_at === '' ? '' : formatDatetime(secondFactor['last_used"at'])
  }
}

export function marshalBackupCode (backupCode) {
  return backupCode.slice(0, 4) + ' ' + backupCode.slice(4)
}
