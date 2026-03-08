<script setup lang="ts">
const { rules, addRule, deleteRule } = useRules()

const form = reactive({
  fieldName: 'name',
  value: '',
  exactMatch: false,
  categoryId: undefined as number | undefined,
})

const submitting = ref(false)
const deleteTarget = ref<number | null>(null)
const toastState = reactive({ visible: false, message: '', type: 'success' as 'success' | 'error' })

async function handleSubmit() {
  if (!form.value || form.categoryId == null) return
  submitting.value = true
  try {
    await addRule({
      fieldName: form.fieldName,
      value: form.value,
      categoryId: form.categoryId,
      exactMatch: form.exactMatch,
    })
    form.value = ''
    form.exactMatch = false
    form.categoryId = undefined
    toastState.message = 'ルールを追加しました'
    toastState.type = 'success'
    toastState.visible = true
  } catch {
    toastState.message = '追加に失敗しました'
    toastState.type = 'error'
    toastState.visible = true
  } finally {
    submitting.value = false
  }
}

function requestDelete(id: number) {
  deleteTarget.value = id
}

async function executeDelete() {
  const id = deleteTarget.value
  deleteTarget.value = null
  if (id == null) return
  try {
    await deleteRule(id)
    toastState.message = 'ルールを削除しました'
    toastState.type = 'success'
    toastState.visible = true
  } catch {
    toastState.message = '削除に失敗しました'
    toastState.type = 'error'
    toastState.visible = true
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Add Rule Form -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm">
      <div class="px-4 sm:px-6 py-4 border-b border-gray-200">
        <h2 class="text-base font-semibold text-gray-900">ルール追加</h2>
      </div>
      <form class="px-4 sm:px-6 py-5 space-y-4" @submit.prevent="handleSubmit">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">判定フィールド</label>
            <select
              v-model="form.fieldName"
              class="w-full text-sm border border-gray-300 rounded-md px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            >
              <option value="name">name</option>
              <option value="m_category">m_category</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">フィルタ値</label>
            <input
              v-model="form.value"
              type="text"
              class="w-full text-sm border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="値を入力"
            >
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">カテゴリID</label>
            <input
              v-model.number="form.categoryId"
              type="number"
              class="w-full text-sm border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="例: 101"
            >
          </div>
          <div class="flex items-end">
            <label class="flex items-center gap-2 cursor-pointer select-none">
              <input
                v-model="form.exactMatch"
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              >
              <span class="text-sm text-gray-700">完全一致</span>
            </label>
          </div>
        </div>
        <div>
          <button
            type="submit"
            :disabled="submitting || !form.value || form.categoryId == null"
            class="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <span v-if="submitting" class="inline-block h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
            追加する
          </button>
        </div>
      </form>
    </div>

    <!-- Rule List -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm">
      <div class="px-4 sm:px-6 py-4 border-b border-gray-200">
        <h2 class="text-base font-semibold text-gray-900">ルール一覧</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              <th class="px-4 py-3">ID</th>
              <th class="px-4 py-3">判定フィールド名</th>
              <th class="px-4 py-3">値</th>
              <th class="px-4 py-3 text-center">完全一致</th>
              <th class="px-4 py-3">カテゴリID</th>
              <th class="px-4 py-3 text-center">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr
              v-for="rule in rules"
              :key="rule.id"
              class="hover:bg-gray-50 transition-colors"
            >
              <td class="px-4 py-2.5 tabular-nums text-gray-500">{{ rule.id }}</td>
              <td class="px-4 py-2.5 text-gray-900">{{ rule.fieldName }}</td>
              <td class="px-4 py-2.5 text-gray-700">{{ rule.value }}</td>
              <td class="px-4 py-2.5 text-center">
                <span
                  class="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
                  :class="rule.exactMatch ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'"
                >
                  {{ rule.exactMatch ? '完全' : '部分' }}
                </span>
              </td>
              <td class="px-4 py-2.5 tabular-nums text-gray-700">{{ rule.categoryId }}</td>
              <td class="px-4 py-2.5 text-center">
                <button
                  class="text-xs font-medium text-red-600 hover:text-red-800 transition-colors"
                  @click="requestDelete(rule.id)"
                >
                  削除
                </button>
              </td>
            </tr>
            <tr v-if="rules.length === 0">
              <td colspan="6" class="px-4 py-8 text-center text-gray-400">ルールがありません</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <ConfirmDialog
    :open="deleteTarget != null"
    title="ルール削除"
    message="このルールを削除しますか？"
    confirm-label="削除する"
    @confirm="executeDelete"
    @cancel="deleteTarget = null"
  />

  <Toast
    :visible="toastState.visible"
    :message="toastState.message"
    :type="toastState.type"
    @close="toastState.visible = false"
  />
</template>
