<script setup lang="ts">
const props = defineProps<{
  page: number
  totalPages: number
}>()

const emit = defineEmits<{
  'update:page': [value: number]
}>()

function go(p: number) {
  const next = Math.max(1, Math.min(p, props.totalPages))
  emit('update:page', next)
}
</script>

<template>
  <nav class="flex items-center justify-center gap-2 text-sm" aria-label="ページ切り替え">
    <button
      class="px-3 py-1.5 rounded-md transition-colors"
      :class="page <= 1
        ? 'text-gray-300 cursor-not-allowed'
        : 'text-gray-600 hover:bg-gray-100'"
      :disabled="page <= 1"
      @click="go(page - 1)"
    >
      ← 前
    </button>
    <span class="px-2 text-gray-500 tabular-nums">{{ page }} / {{ totalPages }}</span>
    <button
      class="px-3 py-1.5 rounded-md transition-colors"
      :class="page >= totalPages
        ? 'text-gray-300 cursor-not-allowed'
        : 'text-gray-600 hover:bg-gray-100'"
      :disabled="page >= totalPages"
      @click="go(page + 1)"
    >
      次 →
    </button>
  </nav>
</template>
