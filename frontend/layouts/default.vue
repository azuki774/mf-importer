<script setup lang="ts">
const route = useRoute()

const navItems = [
  { label: '取り込み履歴', to: '/' },
  { label: 'ルール設定', to: '/rules' },
]
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50">
    <header class="bg-white border-b border-gray-200 sticky top-0 z-30">
      <div class="max-w-6xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        <NuxtLink to="/" class="text-lg font-semibold text-gray-900 hover:text-primary-600 transition-colors">
          mf-importer
        </NuxtLink>
        <nav class="flex items-center gap-1">
          <NuxtLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors"
            :class="route.path === item.to
              ? 'bg-primary-50 text-primary-700'
              : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'"
          >
            {{ item.label }}
          </NuxtLink>
        </nav>
      </div>
    </header>

    <main class="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-6">
      <slot />
    </main>

    <footer class="border-t border-gray-200 bg-white">
      <div class="max-w-6xl mx-auto px-4 sm:px-6 py-4 text-center text-xs text-gray-400">
        mf-importer
      </div>
    </footer>
  </div>
</template>
