<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";
const config = useRuntimeConfig(); // nuxt.config.ts に書いてあるコンフィグを読み出す
const record_list = ref<ImportRecord[]>()
const asyncData = await useFetch(
  "/api/details",
  {
    key: `/api/details`,
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
    <div class="mb-3 mt-3">
      <input type="button" class="sendbutton btn btn-primary" onclick="location.href='./rules'" value="ルール設定を表示">
    </div>

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
