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

onMounted(
  (): void => {
    const data = asyncData.data.value as Rule[];
    if (data != undefined) {
      rule_list.value = data;
    }
  }
)


const postButton = async (): Promise<void> => {
  const param = { 'fieldName': ruleFieldSelector.value, 'value': ruleFilterValue.value, 'categoryId': ruleCategoryID.value, 'exactMatch': ruleExactCheck.checked }
  // exactMatch: true -> 1, false -> 0
  if (param['exactMatch'] == true) {
    param['exactMatch'] = 1;
  } else {
    param['exactMatch'] = 0;
  }

  const localurl = "/api/rules"
  // create body
  const reqbody = `
  {
  "fieldName":"${param.fieldName}",
  "value":"${param.value}",
  "categoryId":${param.categoryId},
  "exactMatch":${param.exactMatch}
  }
  `
  console.log(reqbody)
  const result = await $fetch(localurl,
    {
      method: "POST",
      body: reqbody
    }
  )

  location.reload()
};

</script>

<template>
  <section class="container">
    <div class="mb-3 mt-3">
      <input type="button" class="sendbutton btn btn-primary" onclick="location.href='../'" value="トップに戻る">
    </div>

    <h3>抽出連携ルール</h3>

    <h4>追加</h4>
    <div class="mb-2">
      <select class="form-select" id="ruleFieldSelector">
        <option value="name" selected>name</option>
        <option value="m_category">m_category</option>
      </select>
    </div>

    <div class="mb-3">
      <label class="form-label">フィルタ値</label>
      <input class="form-control" id="ruleFilterValue">
    </div>

    <div class="mb-1">
      <div class="form-check">
        <input type="checkbox" class="form-check-input" value="" id="ruleExactCheck">
        <label for="formCheckDefault" class="form-check-label">
          完全一致
        </label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label">カテゴリID</label>
      <input type="number" class="form-control" id="ruleCategoryID">
    </div>

    <div class="mb-3">
      <button type="submit" @click="postButton" name="postButton" class="btn btn-primary">追加する</button>
    </div>

    <h4>一覧</h4>
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
