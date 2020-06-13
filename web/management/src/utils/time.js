export async function delay (time) {
  return new Promise((resolve, reject) => {
    try {
      setTimeout(resolve, time)
    } catch (err) {
      reject(err)
    }
  })
}
