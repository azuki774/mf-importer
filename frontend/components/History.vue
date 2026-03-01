<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";

const page = ref(1);
const perPage = ref(20);
const perPageOptions = [10, 20, 50];

const detailsKey = computed(
  () => `/api/details?limit=${perPage.value}&offset=${(page.value - 1) * perPage.value}`
);

const { data: detailsData } = await useFetch<ImportRecord[]>(() => detailsKey.value, {
  key: () => `details-${page.value}-${perPage.value}`,
});

const totalCount = ref(0);

onMounted(async () => {
  try {
    const res = await $fetch<{ count: number }>("/api/details/count");
    totalCount.value = res.count ?? 0;
  } catch {
    totalCount.value = 0;
  }
});
const totalPages = computed(() => Math.ceil(totalCount.value / perPage.value) || 1);

const record_list = computed(() => {
  const data = detailsData.value;
  if (data == null) return [];
  return data.map((d) => {
    const r = { ...d };
    r.registDate = d.registDate.slice(0, 19);
    if (d.importJudgeDate != null) r.importJudgeDate = d.importJudgeDate.slice(0, 19);
    if (d.importDate != null) r.importDate = d.importDate.slice(0, 19);
    return r;
  });
});

function goToPage(p: number) {
  const next = Math.max(1, Math.min(p, totalPages.value));
  page.value = next;
}

function onPerPageChange(ev: Event) {
  const v = Number((ev.target as HTMLSelectElement).value);
  if (v > 0) {
    perPage.value = v;
    page.value = 1;
  }
}

async function showPatchDialog(id: number): Promise<void> {
  const userResponse: boolean = confirm("このデータを再判定対象にしますか");
  if (userResponse === true) {
    const paramStr = "?id=" + id + "&ope=reset";
    await $fetch("/api/detail" + paramStr, { method: "PATCH" });
    location.reload();
  }
}
</script>

<template>
  <section class="container">
    <h3>取り込み履歴</h3>
    <div class="mb-3 mt-3">
      <input
        type="button"
        class="sendbutton btn btn-primary"
        onclick="location.href='./rules'"
        value="ルール設定を表示"
      >
    </div>

    <div class="d-flex justify-content-between align-items-center mb-2">
      <div class="d-flex align-items-center gap-2">
        <label class="form-label mb-0">表示件数</label>
        <select
          class="form-select form-select-sm"
          style="width: auto"
          :value="perPage"
          @change="onPerPageChange"
        >
          <option
            v-for="n in perPageOptions"
            :key="n"
            :value="n"
          >
            {{ n }}件
          </option>
        </select>
      </div>
      <span class="text-muted small">全 {{ totalCount }} 件</span>
    </div>

    <table class="table small bordered striped table-bordered">
      <thead class="table-info">
        <tr>
          <th scope="col">利用日時</th>
          <th scope="col">名前</th>
          <th scope="col">料金</th>
          <th scope="col">登録日時</th>
          <th scope="col">取り込み判定日時</th>
          <th scope="col">取り込み日時</th>
          <th scope="col">再判定</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="record in record_list"
          :key="record.id"
        >
          <td>{{ record.useDate }}</td>
          <td>{{ record.name }}</td>
          <td>{{ record.price }}</td>
          <td>{{ record.registDate }}</td>
          <td>{{ record.importJudgeDate }}</td>
          <td>{{ record.importDate }}</td>
          <td>
            <button
              class="btn btn-secondary btn-sm"
              @click="showPatchDialog(record.id)"
            >
              再判定
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="mt-4 p-3 border rounded bg-light" aria-label="ページ切り替え">
      <p class="mb-2 mb-md-0 small fw-bold text-secondary">
        ページ切り替え
      </p>
      <div class="d-flex flex-wrap align-items-center justify-content-center gap-2 gap-md-3">
        <button
          type="button"
          class="btn btn-outline-primary btn-sm"
          :disabled="page <= 1"
          aria-label="前のページ"
          @click="goToPage(page - 1)"
        >
          前のページ
        </button>
        <span class="px-2 text-nowrap small">
          {{ page }} / {{ totalPages }} ページ
        </span>
        <button
          type="button"
          class="btn btn-outline-primary btn-sm"
          :disabled="page >= totalPages"
          aria-label="次のページを表示"
          @click="goToPage(page + 1)"
        >
          次のページを表示
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
</style>
