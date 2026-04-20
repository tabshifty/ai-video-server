export function getRouteTransitionName(route) {
  if (route?.meta?.public) {
    return 'fade-slide'
  }
  return undefined
}
