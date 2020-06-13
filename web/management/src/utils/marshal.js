import _ from 'lodash'

import { i18n } from '@/i18n-setup'

import { formatToPhoneString, formatDatetime } from '@/utils/format'

export function marshalContact (contact) {
  function getType (type) {
    switch (type) {
      case undefined: return 'email'
      case 'EMAIL': return 'email'
      case 'PHONE': return 'phone'
      default: throw new Error(`undefined type ${type}`)
    }
  }
  let valueString = contact.value
  if (contact.type === 'PHONE') {
    valueString = formatToPhoneString(contact.value)
  }
  return {
    id: parseInt(contact.id, 10),
    type: getType(contact.type),
    value: valueString,
    isPrimary: contact.primary,
    isVerified: contact.verified || false
  }
}

export function marshalAuditLog (auditLog) {
  return {
    id: parseInt(auditLog.id, 10),
    username: auditLog.username,
    action: i18n.t(auditLog.action),
    target: auditLog.target,
    sessionId: auditLog.session_id,
    device: auditLog.device,
    ip: auditLog.ip,
    description: auditLog.description,
    result: i18n.t(`general.${auditLog.result}`),
    isInternal: auditLog.is_internal,
    createdAt: formatDatetime(auditLog.created_at)
  }
}

export function marshalPermission (permission) {
  return {
    name: permission.name
  }
}

export function marshalRole (role) {
  const roleName = _.last(role.name.split('.'))
  const beautifyName = i18n.t(`model.role.${roleName.toLowerCase()}`)
  const description = i18n.t(`user_details_roles.text.${roleName.toLowerCase()}`)
  return {
    id: parseInt(role.id, 10),
    name: beautifyName,
    description: description
  }
}

export function marshalSecondFactor (device) {
  // Default to be not set, only SMS device has relevant value
  let valueString = '-'
  if (device.type === 'sms_otp' && device.value !== '') {
    valueString = formatToPhoneString(device.value)
  }
  let lastUsedAt = '-'
  let createdAt = '-'
  if (device.last_used_at !== '') {
    createdAt = formatDatetime(device.last_used_at)
  }
  if (device.last_used_at !== '') {
    lastUsedAt = formatDatetime(device.last_used_at)
  }
  const deviceName = i18n.t(`model.second_factor.${device.type.toLowerCase()}`)
  return {
    id: device.id,
    type: device.type,
    name: deviceName,
    value: valueString,
    createdAt: createdAt,
    lastUsedAt: lastUsedAt
  }
}

export function marshalUser (user) {
  let profileName
  if (user.name) {
    profileName = user.name
  } else if (user.preferred_username) {
    profileName = user.preferred_username
  } else if (user.email) {
    profileName = user.email
  } else if (user.phone_number) {
    profileName = user.phone_number
  } else {
    profileName = 'No data shown'
  }
  return {
    id: parseInt(user.id, 10),
    profileName: profileName,
    name: user.name,
    username: user.preferred_username,
    email: user.email,
    emailVerified: user.email_verified,
    phoneNumber: user.phone_number,
    phoneNumberVerified: user.phone_number_verified,
    locked: user.locked,
    roles: user.roles,
    createdAt: formatDatetime(user.created_at),
    lastSeenAt: user.last_seen_at,
    userMetadata: user.user_metadata !== null ? user.user_metadata : '',
    appMetadata: user.app_metadata !== null ? user.app_metadata : ''
  }
}

export function marshalOAuthFactor (oauthFactor) {
  let service = oauthFactor.service || 'FACEBOOK'
  service = i18n.t(`model.oauth_factor.service.${service.toLowerCase()}`)
  return {
    id: oauthFactor.id,
    userId: oauthFactor.user_id,
    service: service,
    oauthUserId: oauthFactor.oauth_user_id,
    createdAt: oauthFactor.created_at === '' ? '' : formatDatetime(oauthFactor.created_at),
    lastUsedAt: oauthFactor.last_used_at === '' ? '' : formatDatetime(oauthFactor.last_used_at),
    metadata: oauthFactor.metadata === '' ? '{}' : oauthFactor.metadata
  }
}

export function marshalSession (session) {
  let lastSeenAt = ''
  let expiredAt = ''
  if (session.last_seen_at !== '0001-01-01T00:00:00Z') {
    lastSeenAt = formatDatetime(session.last_seen_at)
  }
  if (session.expired_at !== '0001-01-01T00:00:00Z') {
    expiredAt = formatDatetime(session.expired_at)
  }
  return {
    id: parseInt(session.id, 10),
    userAgent: session.user_agent,
    lastSeenAt: lastSeenAt,
    lastSeenIp: session.last_seen_ip,
    lastSeenLocation: session.last_seen_location,
    expiredAt: expiredAt
  }
}

export function marshalTemplate (template) {
  let updatedAt = ''
  if (template.updated_at) {
    updatedAt = formatDatetime(template.updated_at)
  }
  return {
    language: template.language,
    name: template.name,
    updatedAt: updatedAt
  }
}
