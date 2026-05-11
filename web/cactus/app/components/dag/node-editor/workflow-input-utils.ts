export function workflowInputPath(name: string) {
  return `$.message.value.${name}`
}
