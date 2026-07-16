export function shouldShowCrudCollectionSkeleton({ loading, rowCount }) {
  return loading && rowCount === 0
}
