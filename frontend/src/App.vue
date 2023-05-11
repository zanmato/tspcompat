<script setup>
import WordsIcon from "bootstrap-icons/icons/list.svg?component";
import FavoritesIcon from "bootstrap-icons/icons/suit-heart.svg?component";
import SubjectsIcon from "bootstrap-icons/icons/tag.svg?component";
import AboutIcon from "bootstrap-icons/icons/info-circle.svg?component";
import { useI18n } from "vue-i18n";

const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <main>
    <KeepAlive :max="10">
      <router-view></router-view>
    </KeepAlive>
  </main>
  <nav class="tab-bar">
    <router-link to="/">
      <WordsIcon />
      <span>{{ t("words") }}</span>
    </router-link>
    <router-link to="/favorites">
      <FavoritesIcon />
      <span>{{ t("favorites") }}</span>
    </router-link>
    <router-link to="/categories">
      <SubjectsIcon />
      <span>{{ t("categories") }}</span>
    </router-link>
    <router-link to="/about">
      <AboutIcon />
      <span>{{ t("about") }}</span>
    </router-link>
  </nav>
</template>

<style lang="scss" scoped>
.main-view {
  max-width: 320px;
}

.split-view {
  width: auto;
}

.tab-bar {
  position: fixed;
  bottom: 0;
  left: 0;

  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 60px;

  background: var(--background-color);
  border-top: 1px solid var(--separator-color);

  a {
    flex: 1 0;
    line-height: 1.2;
    color: var(--secondary-label-color);
    text-align: center;

    &:disabled {
      opacity: 0.6;
    }

    .bi {
      fill: var(--secondary-label-color);
    }

    &.router-link-active {
      color: var(--tint-color);

      .bi {
        fill: var(--tint-color);
      }
    }

    span {
      display: block;
      font-size: 0.8rem;
    }
  }
}
</style>
