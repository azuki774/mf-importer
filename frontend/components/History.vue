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
    if (d.importDate != undefined) {
      d.importDate = d.importDate.slice(0, 19);
    }
  }

  record_list.value = data
}

async function showPatchDialog(id: number): Promise<void> {
  const userResponse: boolean = confirm("このデータを再判定対象にしますか");
  if (userResponse == true) {
    console.log('patch: id=' + id);
    const asyncDataDeleteBtn = await useAsyncData(
      `patchDetail `,
      (): Promise<any> => {
        const param = { 'id': id, 'ope': 'reset' };
        const paramStr = "?id=" + param['id'] + "&ope=" + param['ope'];
        const localurl = "/api/detail" + paramStr
        console.log(localurl)
        const response = $fetch(localurl,
          {
            method: "PATCH"
          }
        );
        return response;
      }
    );
    location.reload()
  }
};

</script>

<template>
  <section class="container">
    <h3>取り込み履歴</h3>
    <div class="mb-3 mt-3">
      <input type="button" class="sendbutton btn btn-primary" onclick="location.href='./rules'" value="ルール設定を表示">
    </div>
    <table class="table small bordered striped table-bordered">
      <thead class="table-info">
        <tr>
          <th scope="col">利用日時</th>
          <th scope="col">名前</th>
          <th scope="col">料金</th>
          <th scope="col">登録日時</th>
          <th scope="col">取り込み判定日時</th>
          <th scope="col">取り込み日時</th>
          <th scope="col">再判定</th>
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
          <td><button class="btn btn-secondary btn-sm" @click="showPatchDialog(record.id)">再判定</button></td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style lang="css"></style>
