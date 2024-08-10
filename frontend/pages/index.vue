<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";
const record_list = ref<ImportRecord[]>()
// const response = await $fetch("http://172.19.250.172:20010/") as ImportRecord[]; // TODO: アドレス
const asyncData = await useAsyncData(
  `api`,
  (): Promise<any> => {
    const url = "http://172.19.250.172:20010/";
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
          <th>登録日時</th>
          <th>取り込み判定日時</th>
          <th>取り込み日時</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="record in record_list" :key="record_list">
          <td>{{ record.use_date }}</td>
          <td>{{ record.name }}</td>
          <td>{{ record.regist_date }}</td>
          <td>{{ record.import_judge_date }}</td>
          <td>{{ record.import_date }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>
