/**
 * Sanitize DataFlow manifest for display in YAML editor.
 * Removes read-only Kubernetes metadata fields that clutter the UI.
 * Preserves name, namespace, labels, annotations.
 */
export function sanitizeManifestForDisplay(manifest) {
  if (!manifest || typeof manifest !== 'object') return manifest
  const copy = JSON.parse(JSON.stringify(manifest))
  if (copy.metadata) {
    const meta = copy.metadata
    delete meta.uid
    delete meta.resourceVersion
    delete meta.generation
    delete meta.creationTimestamp
    delete meta.managedFields
  }
  if (copy.status) {
    delete copy.status
  }
  return copy
}

/**
 * Merge manifest for update - ensures resourceVersion and uid are preserved for Kubernetes API.
 */
export function mergeManifestForUpdate(parsed, original) {
  if (!original?.metadata) return parsed
  return {
    ...parsed,
    metadata: {
      ...parsed.metadata,
      resourceVersion: original.metadata.resourceVersion,
      uid: original.metadata.uid,
    },
  }
}
