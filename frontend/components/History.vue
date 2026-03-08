<script setup lang="ts">
const page = ref(1)
const perPage = ref(20)
const perPageOptions = [10, 20, 50]

const { records, status, totalCount, totalPages, resetDetail } = useDetails(perPage, page)

const confirmTarget = ref<number | null>(null)
const toastState = reactive({ visible: false, message: '', type: 'success' as 'success' | 'error' })

function onPerPageChange(v: number) {
  perPage.value = v
  page.value = 1
}

function requestReset(id: number) {
  confirmTarget.value = id
}

async function executeReset() {
  const id = confirmTarget.value
  confirmTarget.value = null
  if (id == null) return
  try {
    await resetDetail(id)
    toastState.message = '再判定対象に設定しました'
    toastState.type = 'success'
    toastState.visible = true
  } catch {
    toastState.message = '操作に失敗しました'
    toastState.type = 'error'
    toastState.visible = true
  }
}
</script>

<template>
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm">
    <div class="px-4 sm:px-6 py-4 border-b border-gray-200">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <h2 class="text-base font-semibold text-gray-900">取り込み履歴</h2>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-500">表示件数</label>
            <select
              class="text-sm border border-gray-300 rounded-md px-2 py-1 bg-white focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              :value="perPage"
              @change="onPerPageChange(Number(($event.target as HTMLSelectElement).value))"
            >
              <option v-for="n in perPageOptions" :key="n" :value="n">{{ n }}件</option>
            </select>
          </div>
          <span class="text-sm text-gray-400">全 {{ totalCount }} 件</span>
        </div>
      </div>
    </div>

    <div v-if="status === 'pending'" class="px-6 py-12 text-center">
      <div class="inline-block h-6 w-6 border-2 border-primary-600 border-t-transparent rounded-full animate-spin" />
      <p class="mt-2 text-sm text-gray-500">読み込み中…</p>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            <th class="px-4 py-3">利用日時</th>
            <th class="px-4 py-3">名前</th>
            <th class="px-4 py-3 text-right">料金</th>
            <th class="px-4 py-3">登録日時</th>
            <th class="px-4 py-3">取り込み判定日時</th>
            <th class="px-4 py-3">取り込み日時</th>
            <th class="px-4 py-3 text-center">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="record in records"
            :key="record.id"
            class="hover:bg-gray-50 transition-colors"
          >
            <td class="px-4 py-2.5 whitespace-nowrap text-gray-700">{{ record.useDate }}</td>
            <td class="px-4 py-2.5 text-gray-900">{{ record.name }}</td>
            <td class="px-4 py-2.5 text-right tabular-nums text-gray-700">{{ record.price.toLocaleString() }}</td>
            <td class="px-4 py-2.5 whitespace-nowrap text-gray-500 text-xs">{{ record.registDate }}</td>
            <td class="px-4 py-2.5 whitespace-nowrap text-gray-500 text-xs">{{ record.importJudgeDate || '—' }}</td>
            <td class="px-4 py-2.5 whitespace-nowrap text-gray-500 text-xs">{{ record.importDate || '—' }}</td>
            <td class="px-4 py-2.5 text-center">
              <button
                class="text-xs font-medium text-primary-600 hover:text-primary-800 transition-colors"
                @click="requestReset(record.id)"
              >
                再判定
              </button>
            </td>
          </tr>
          <tr v-if="records.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-gray-400">データがありません</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="px-4 sm:px-6 py-3 border-t border-gray-200">
      <Pagination
        :page="page"
        :total-pages="totalPages"
        @update:page="page = $event"
      />
    </div>
  </div>

  <ConfirmDialog
    :open="confirmTarget != null"
    title="再判定"
    message="このデータを再判定対象にしますか？"
    confirm-label="再判定する"
    @confirm="executeReset"
    @cancel="confirmTarget = null"
  />

  <Toast
    :visible="toastState.visible"
    :message="toastState.message"
    :type="toastState.type"
    @close="toastState.visible = false"
  />
</template>
