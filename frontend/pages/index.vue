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
      </tbody>
    </table>
  </section>

</template>


<script setup lang="ts">
const record_name_list = ref<string[]>(['テスト明細1', 'テスト明細2', 'テスト明細3']);
</script>
