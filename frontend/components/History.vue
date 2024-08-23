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
if (data != undefined) { // 取得済の場合のみ
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
}


</script>

<template>
  <section class="container">
    <h3>取り込み履歴</h3>
    <table class="table small bordered striped table-bordered">
      <thead class="table-info">
        <tr>
          <th scope="col">利用日時</th>
          <th scope="col">名前</th>
          <th scope="col">料金</th>
          <th scope="col">登録日時</th>
          <th scope="col">取り込み判定日時</th>
          <th scope="col">取り込み日時</th>
        </tr>
      </thead>
      <tbody>
        <th scope="row"></th>
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

<style lang="css"></style>
