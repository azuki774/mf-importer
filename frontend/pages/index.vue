<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";
const config = useRuntimeConfig(); // nuxt.config.ts に書いてあるコンフィグを読み出す
const record_list = ref<ImportRecord[]>()
const asyncData = await useAsyncData(
  `api`,
  (): Promise<any> => {
    const url = config.public.apiBaseEndpoint + "/details";
    const response = $fetch(url);
    return response;
  }
);

const data = asyncData.data.value as ImportRecord[];

// fetchデータを整形
for (let d of data) {
  d.registDate = d.registDate.slice(0, 19);
  if (d.importJudgeDate != undefined) { // 2023-09-23T00:00:00+09:00 -> 2023-09-23T00:00:00
    d.importJudgeDate = d.importJudgeDate.slice(0, 19);
  }
  d.importJudgeDate = d.importJudgeDate.slice(0, 19);
  if (d.importDate != undefined) {
    d.importDate = d.importDate.slice(0, 19);
  }
}

record_list.value = data
</script>

<template>
  <h1>mf-importer-web</h1>

  <section class="container">
    <h1>リスト</h1>
    <table class="Lists">
      <thead>
        <tr>
          <th>利用日時</th>
          <th>名前</th>
          <th>料金</th>
          <th>登録日時</th>
          <th>取り込み判定日時</th>
          <th>取り込み日時</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="record in record_list">
          <td>{{ record.useDate }}</td>
          <td>{{ record.name }}</td>
          <td>{{ record.price }}</td>
          <td>{{ record.registDate }}</td>
          <td>{{ record.importJudgeDate }}</td>
          <td>{{ record.importDate }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style lang="css">
.Lists {
  width: 75%;
  text-align: center;
  margin-left: auto;
  margin-right: auto;
  border-collapse: collapse;
  border-spacing: 0;
}

.Lists th {
  padding: 10px;
  background: #e9faf9;
  border: solid 1px #778ca3;
}

.Lists td {
  padding: 10px;
  border: solid 1px #778ca3;
}
</style>
