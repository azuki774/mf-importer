<!-- <script setup lang="ts">
import type { City } from "@/interfaces";
//都市情報リストをステートとして用意。
useState<Map<number, City>>(
  "cityList",
  (): Map<number, City> => {
    const cityListInit = new Map<number, City>();
    cityListInit.set(1853909,
      {
        id: 1853909,
        name: "大阪",
        q: "Osaka"
      });
    cityListInit.set(1859171,
      {
        id: 1859171,
        name: "神戸",
        q: "Kobe"
      });
    cityListInit.set(1862627,
      {
        id: 1862627,
        name: "姫路",
        q: "Himeji"
      });
    return cityListInit;
  }
);
</script> -->

<template>
  <header>
    <h1>fetchサンプル</h1>
  </header>
  <main>
    <NuxtPage />
  </main>
</template>
