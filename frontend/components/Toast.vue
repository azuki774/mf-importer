<script setup lang="ts">
const props = defineProps<{
  message: string
  type?: 'success' | 'error'
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

watch(() => props.visible, (v) => {
  if (v) {
    setTimeout(() => emit('close'), 3000)
  }
})

const colorClass = computed(() =>
  props.type === 'error'
    ? 'bg-red-50 text-red-800 border-red-200'
    : 'bg-green-50 text-green-800 border-green-200'
)
</script>

<template>
  <Teleport to="body">
    <Transition name="slide">
      <div
        v-if="visible"
        class="fixed top-4 right-4 z-50 max-w-sm w-full border rounded-lg shadow-lg px-4 py-3 text-sm font-medium"
        :class="colorClass"
      >
        <div class="flex items-center justify-between gap-2">
          <span>{{ message }}</span>
          <button
            class="shrink-0 opacity-60 hover:opacity-100 transition-opacity"
            @click="emit('close')"
          >
            ✕
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}
.slide-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}
.slide-leave-to {
  opacity: 0;
  transform: translateX(16px);
}
</style>
