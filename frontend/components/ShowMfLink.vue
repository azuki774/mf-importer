<script setup lang="ts">
const MonthLinkMove = (lastmonth: boolean): void => {
  // 現在日付取得
  const today = new Date();
  let year = today.getFullYear();
  let month = today.getMonth() + 1; // 0から11 -> 1から12
  let lyear = year; // 先月の情報
  let lmonth = month - 1; // 先月の情報

  // xx 年 1月の前月取得の場合
  if (lmonth == 0) {
    lmonth = 12;
    lyear = year - 1;
  }

  if (lastmonth) {
    // 先月のリンク
    const MonthLink = 'https://moneyforward.com/cf/csv?from=' + lyear + '%2F' + lmonth + '%2F01&month=' + lmonth + '&year=' + lyear
    console.log(MonthLink)
    window.location.href = MonthLink;
  } else {
    // 今月のリンク
    const MonthLink = 'https://moneyforward.com/cf/csv?from=' + year + '%2F' + month + '%2F01&month=' + month + '&year=' + year
    console.log(MonthLink)
    window.location.href = MonthLink;
  }

}


</script>

<template>
  <section class="container">
    <h4>取り込みリンク</h4>
    <div class="col-6 col-sm-3 mb-2">
      <input type="button" class="sendbutton btn btn-primary" @click="MonthLinkMove(false)" value="今月分">
    </div>
    <div class="col-6 col-sm-3 mb-2">
      <input type="button" class="sendbutton btn btn-primary" @click="MonthLinkMove(true)" value="先月分">
    </div>

  </section>
</template>

<style lang="css"></style>
