import { defineStore } from 'pinia'

export const useCounterStore = defineStore('counter', {
    state: (): { c: number } => ({ c: 0 }),
    getters: {
        count: (state): number => state.c
    },
    actions: {
        increment(): void {
            this.c++
        }
    }
})
