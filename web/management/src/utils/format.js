import moment from 'moment'
import { parsePhoneNumberFromString } from 'libphonenumber-js'

const datetimeGeneralFormat = 'HH:mm:ss, DD MMM Y'

/**
 *  Format the phone into international string format for easier reading.
 **/
export function formatToPhoneString (phone) {
  // TODO: Hardcoded for Hong Kong, ref: https://gitlab.com/blocksq/authcore/issues/14
  if (phone) {
    const phoneNumber = parsePhoneNumberFromString(phone, 'HK')
    if (phoneNumber !== undefined) {
      return phoneNumber.formatInternational()
    }
  }
  return phone
}

export function formatDatetime (datetime) {
  return moment(datetime).format(datetimeGeneralFormat)
}
