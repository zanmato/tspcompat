<script setup>
import { useI18n } from "vue-i18n";
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <div class="header">
    <div class="header-content">
      <h1>{{ t("about") }}</h1>
    </div>
  </div>
  <section class="inner-content about">
    <i18n-t keypath="aboutText" tag="p">
      <template #source>
        <a href="https://teckensprakslexikon.su.se/" target="_blank" rel="nofollow">Svenskt Teckenspråkslexikon</a>
      </template>
    </i18n-t>
  </section>
</template>
<style lang="scss">
.inner-content.about {
  padding: 0 15px;
}
</style>
