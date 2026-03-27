import { useIntervalFn } from '@vueuse/core'

export function usePolling(
  callback: () => Promise<void>,
  interval: Ref<number> = ref(15000),
) {
  const isPolling = ref(true)

  const { pause: internalPause, resume: internalResume } = useIntervalFn(
    async () => {
      if (!isPolling.value) return
      try {
        await callback()
      }
      catch (err) {
        console.warn('[usePolling] Callback error:', err)
      }
    },
    interval,
    { immediate: false },
  )

  function pause() {
    isPolling.value = false
    internalPause()
  }

  function resume() {
    isPolling.value = true
    internalResume()
  }

  function setInterval(ms: number) {
    interval.value = ms
  }

  onMounted(() => {
    if (isPolling.value) {
      internalResume()
    }
  })

  onUnmounted(() => {
    internalPause()
  })

  return {
    isPolling: readonly(isPolling),
    pause,
    resume,
    setInterval,
  }
}
