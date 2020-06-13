// isRole returns whether the user have the specified role.
export function isRole (user, role) {
  return user.roles && user.roles.some(item => item.name === role)
}

// isAdmin returns whether the user is admin or not.
export function isAdmin (user) {
  return isRole(user, 'authcore.admin') || isRole(user, 'authcore.editor')
}
