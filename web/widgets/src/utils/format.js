import moment from 'moment'

const datetimeGeneralFormat = 'YYYY-MM-DD HH:mm'

export function formatDatetime (datetime) {
  return moment(datetime).format(datetimeGeneralFormat)
}
