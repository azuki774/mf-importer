<script setup lang="ts">
import type { City } from "@/interfaces";
//ルートオブジェクトを用意、
const route = useRoute();
//都市情報リストをステートから取得。
const cityList = useState<Map<number, City>>("cityList");
//ルートパラメータをもとに該当都市データを取得。
const selectedCity = computed(
    (): City => {
        const idNo = Number(route.params.id);
        return cityList.value.get(idNo) as City;
    }
);
//天気情報のテンプレート変数を用意。
const weatherDescription = ref("");

const weatherInfoUrl = "https://api.openweathermap.org/data/2.5/weather"
const params: {
    lang: string;
    q: string;
    appid: string;
} = {
    lang: "ja",
    q: selectedCity.value.q,
    appid: "ab61e67ff2ea6eeb72f80b3f55ef4ccd"
}

const queryParams = new URLSearchParams(params);
const urlFull = `${weatherInfoUrl}?${queryParams}`;
//URLに非同期でアクセスしてデータを取得。
const response = await $fetch(urlFull) as any;
//天気情報JSONから天気データを取得し、テンプレート変数に格納。
const weatherArray = response.weather;
const weather = weatherArray[0];
weatherDescription.value = weather.description
</script>


<template>
    <section>
        <h2>{{ selectedCity.name }}の天気</h2>
        <p>{{ weatherDescription }}</p>
    </section>
    <p>リストに<NuxtLink v-bind:to="{ name: 'index' }">戻る</NuxtLink>
    </p>
</template>
