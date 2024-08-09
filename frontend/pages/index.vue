<script setup lang="ts">
import type { ImportRecord } from "@/interfaces";
const record_name_list = ref<string[]>(['テスト明細1', 'テスト明細2', 'テスト明細3']);
const record_list = ref<ImportRecord[]>()
// const asyncData = await useFetch("http://172.19.250.172:20010/",
//   {
//     transform: (data: any): string => {
//       const jsname = data[0].name;
//       return jsname;
//     }
//   }
// );
// const res = asyncData.data;
const response = await $fetch("http://172.19.250.172:20010/") as ImportRecord[];
record_list.value = response // 取得した値を代入
console.log(`${record_list.value[1].name}`)
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
