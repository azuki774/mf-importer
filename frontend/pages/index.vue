<script setup lang="ts">
const record_name_list = ref<string[]>(['テスト明細1', 'テスト明細2', 'テスト明細3']);
const asyncData = await useFetch("https://www.jma.go.jp/bosai/forecast/data/forecast/130000.json",
  {
    transform: (data: any): string => {
      const jsname = data[0].publishingOffice;
      return jsname;
    }
  }
);
const res = asyncData.data;
console.log(`${res.value}`)
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
        <tr>
          <td>2024-01-23</td>
          <td>テスト明細0</td>
          <td>2024-01-25 12:00</td>
          <td>2024-01-25 15:00</td>
          <td>NULL</td>
        </tr>
        <tr v-for="record_name in record_name_list" :key="record_name">
          <td>2024-01-23</td>
          <td>{{ record_name }}</td>
          <td>2024-01-25 12:00</td>
          <td>2024-01-25 15:00</td>
          <td>NULL</td>
        </tr>
        <tr>
          <td>2024-01-23</td>
          <td>{{ res }}</td>
          <td>2024-01-25 12:00</td>
          <td>2024-01-25 15:00</td>
          <td>NULL</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>
