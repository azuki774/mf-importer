<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";
const config = useRuntimeConfig(); // nuxt.config.ts に書いてあるコンフィグを読み出す
const record_list = ref<any>()
const asyncData = await useAsyncData(
  `api`,
  (): Promise<any> => {
    const url = config.public.apiBaseEndpoint + "/details";
    const response = $fetch(url);
    return response;
  }
);
const data = asyncData.data;
record_list.value = data.value
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
        <tr v-for="record in record_list" :key="record_list">
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
  width: 100%;
  text-align: center;
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
