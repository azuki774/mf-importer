<script setup lang="ts">
function buildMfLink(lastMonth: boolean): string {
  const today = new Date()
  let year = today.getFullYear()
  let month = today.getMonth() + 1

  if (lastMonth) {
    month -= 1
    if (month === 0) {
      month = 12
      year -= 1
    }
  }

  return `https://moneyforward.com/cf/csv?from=${year}%2F${month}%2F01&month=${month}&year=${year}`
}

function goToMf(lastMonth: boolean) {
  window.location.href = buildMfLink(lastMonth)
}
</script>

<template>
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm">
    <div class="px-4 sm:px-6 py-4 border-b border-gray-200">
      <h2 class="text-base font-semibold text-gray-900">MoneyForward CSV 取り込み</h2>
    </div>
    <div class="px-4 sm:px-6 py-4 flex flex-wrap gap-3">
      <button
        class="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 transition-colors"
        @click="goToMf(false)"
      >
        今月分
      </button>
      <button
        class="inline-flex items-center px-4 py-2 text-sm font-medium text-primary-700 bg-primary-50 border border-primary-200 rounded-md hover:bg-primary-100 transition-colors"
        @click="goToMf(true)"
      >
        先月分
      </button>
    </div>
  </div>
</template>
