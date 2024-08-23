<script setup lang="ts">
import type { Rule } from "@/interfaces";
const config = useRuntimeConfig(); // nuxt.config.ts に書いてあるコンフィグを読み出す
const rule_list = ref<Rule[]>()
const asyncData = await useFetch(
  "/api/rules",
  {
    key: `/api/rules`,
  }
);

const data = asyncData.data.value as Rule[];
rule_list.value = data;

</script>

<template>
  <section class="container">
    <h3>抽出連携ルール</h3>
    <table class="table small bordered striped table-bordered">
      <thead class="table-warning">
        <tr>
          <th scope="col">ID</th>
          <th scope="col">判定フィールド名</th>
          <th scope="col">値</th>
          <th scope="col">完全一致</th>
          <th scope="col">カテゴリID</th>
        </tr>
      </thead>
      <tbody>
        <th scope="row"></th>
        <tr v-for="rule in rule_list">
          <td>{{ rule.id }}</td>
          <td>{{ rule.fieldName }}</td>
          <td>{{ rule.value }}</td>
          <td>{{ rule.exactMatch }}</td>
          <td>{{ rule.categoryId }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style lang="css"></style>
