<script setup>
import { reactive } from "vue";
import { useRouter } from "vue-router";
import ChevronLeftIcon from "bootstrap-icons/icons/chevron-left.svg?component";
import HeartIcon from "bootstrap-icons/icons/suit-heart.svg?component";
import HeartFillIcon from "bootstrap-icons/icons/suit-heart-fill.svg?component";
import { useFavoritesStore } from "@/stores/favorites";
import { settings } from "@/api.js";
import { useI18n } from "vue-i18n";

const { t } = useI18n({ useScope: "global" });
const favoritesStore = useFavoritesStore();
const router = useRouter();

const state = reactive({ sign: {} });

const fetchSign = (id) => {
  fetch(`${settings.baseURL}/api/signs/${id}`)
    .then((response) => response.json())
    .then((data) => {
      state.sign = data;
    });
};

fetchSign(router.currentRoute.value.params.id);
</script>
<template>
  <div class="header">
    <div class="header-content">
      <button class="back" type="button" @click="router.go(-1)">
        <ChevronLeftIcon />
      </button>
      <div class="title">
        <h1 v-if="state.sign.words">{{ state.sign.words[0] }}</h1>
        <small>{{ state.sign.id }}</small>
      </div>
      <div class="right-controls">
        <button
          type="button"
          @click="favoritesStore.toggleFavorite(state.sign)">
          <HeartFillIcon v-if="favoritesStore.isFavorite(state.sign.id)" />
          <HeartIcon v-else />
        </button>
      </div>
    </div>
  </div>
  <section class="inner-content">
    <video
      class="sign-video"
      :src="state.sign.video_url"
      playsinline
      muted
      loop
      autoplay
      controls
      controlslist="nodownload" />
    <p class="description">
      {{ state.sign.description }}
    </p>

    <div v-if="state.sign.frequency" class="list">
      <h3>{{ t("comment") }}</h3>
      <ul>
        <li>
          {{ state.sign.frequency }}
        </li>
      </ul>
    </div>

    <div v-if="state.sign.vocable" class="list">
      <h3>{{ t("vocableInCorpus") }}</h3>
      <ul>
        <li>
          <a
            :href="`https://teckensprakskorpus.su.se/#/?q=${state.sign.vocable}`"
            target="_blank"
            rel="nofollow">
            {{ state.sign.vocable }}
          </a>
        </li>
      </ul>
    </div>

    <div v-if="state.sign.transcription" class="list">
      <h3>{{ t("transcription") }}</h3>
      <ul>
        <li class="transcription">
          {{ state.sign.transcription }}
        </li>
      </ul>
    </div>

    <div class="list">
      <h3>{{ t("categories") }}</h3>
      <ul>
        <li v-for="cat in state.sign.categories">
          <router-link :to="{ name: 'CategoryWords', params: { id: cat.id } }">
            {{ cat.name }}
          </router-link>
        </li>
      </ul>
    </div>

    <div class="list">
      <h3>{{ t("words") }}</h3>
      <ul>
        <li v-for="word in state.sign.words">
          {{ word }}
        </li>
      </ul>
    </div>

    <div v-if="state.sign.phrases" class="list">
      <h3>{{ t("phrases") }}</h3>
      <ul>
        <li v-for="ph in state.sign.phrases">
          <video
            class="example-video"
            :src="ph.video_url"
            playsinline
            muted
            loop
            controls
            controlslist="nodownload" />
          {{ ph.phrase }}
        </li>
      </ul>
    </div>
  </section>
</template>
<style lang="scss" scoped>
.header {
  .header-content {
    .title {
      text-align: center;
    }
    h1 {
      font-size: 1.2rem;
    }
  }
}

.sign-video {
  width: 100%;
}

.example-video {
  width: 100%;
}

.header {
  position: fixed;
  width: 100%;
  top: 0;
  z-index: 99;
}

.inner-content {
  margin-top: 80px;

  a {
    color: var(--tint-color);
  }
}

.transcription {
  font-family: FreeSans SWL;
  font-size: 30px;
  padding: 15px 15px 5px;
}
</style>
