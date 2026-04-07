export function useNodeEditor() {
  const isOpen = ref(false)
  const editingNodeId = ref<string | null>(null)

  function open(nodeId: string) {
    editingNodeId.value = nodeId
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
    editingNodeId.value = null
  }

  return { isOpen, editingNodeId, open, close }
}
