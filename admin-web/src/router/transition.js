export function getRouteTransitionName(route) {
  if (route?.meta?.public) {
    return 'fade-slide'
  }
  return undefined
}

export function getRouteViewKey(route) {
  if (route?.path === '/toolbox/china-map') {
    return route.path
  }
  return route?.fullPath || route?.path
}
