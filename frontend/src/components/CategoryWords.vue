<script setup>
import { reactive } from "vue";
import { RecycleScroller } from "vue-virtual-scroller";
import VueSelect from "vue-select";
import ChevronRightIcon from "bootstrap-icons/icons/chevron-right.svg?component";
import { useRouter, useRoute } from "vue-router";
import SearchIcon from "bootstrap-icons/icons/search.svg?component";
import CloseIcon from "bootstrap-icons/icons/x-lg.svg?component";
import ChevronLeftIcon from "bootstrap-icons/icons/chevron-left.svg?component";
import { useI18n } from "vue-i18n";
import { settings } from "@/api.js";

const state = reactive({
  words: [],
});

const route = useRoute();

const fetchWords = () => {
  fetch(
    settings.baseURL + "/api/words?" +
      new URLSearchParams({
        category_id: route.params.id,
      })
  )
    .then((response) => response.json())
    .then((data) => {
      state.words = data;
    });
};

fetchWords();

const router = useRouter();

let debounceTimer = null;

const searchData = reactive({
  results: [],
});

const onSearch = (search, loading) => {
  if (search.trim().length <= 1) {
    return;
  }

  clearTimeout(debounceTimer);

  debounceTimer = setTimeout(() => {
    searchWords(loading, search, this);
  }, 250);
};

const onSelect = (item) => {
  router.push({ name: 'Sign', params: { id: item[1] } })
};

const searchWords = (loading, search, vm) => {
  loading(true);
  fetch(
    settings.baseURL + "/api/words?" +
      new URLSearchParams({
        q: search,
        category_id: route.params.id,
      })
  )
    .then((response) => response.json())
    .then((data) => {
      searchData.results = data;
    })
    .finally(() => {
      loading(false);
    });
};

const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <div class="header">
    <div class="header-content">
      <button class="back" type="button" @click="$router.go(-1)">
        <ChevronLeftIcon />
      </button>
      <h1>{{ t("words") }}</h1>
      <button style="visibility: hidden"><ChevronLeftIcon /></button>
    </div>
    <VueSelect
      ref="search"
      class="search-bar"
      :filterable="false"
      :clearable="true"
      :placeholder="t('search')"
      :options="searchData.results"
      :select-on-key-codes="[]"
      :get-option-key="(opt) => opt[0]"
      :get-option-label="(opt) => opt[2]"
      @search="onSearch"
      @option:selected="onSelect"
    >
      <template #open-indicator>
        <span class="search-icon">
          <SearchIcon />
          <CloseIcon />
        </span>
      </template>
      <template #no-options="{ loading }">
        <span v-if="loading" variant="dark" label="Loading..."></span>
        <span v-else>{{ t("noResults") }}</span>
      </template>
    </VueSelect>
  </div>
  <RecycleScroller
    class="scroller"
    :items="state.words"
    :item-size="40"
    key-field="0"
    v-slot="{ item }"
  >
    <div
      class="word"
      @click="router.push({ name: 'Sign', params: { id: item[1] } })"
    >
      <span>{{ item[2] }}</span>
      <small class="unusual" v-if="item[3]">{{ t("unusualSign") }}</small>
      <ChevronRightIcon />
    </div>
  </RecycleScroller>
</template>

<style lang="scss" scoped>
@import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
@import "vue-select/dist/vue-select.css";

.header {
  height: 120px;
  width: 100%;
  border-bottom: 1px solid var(--separator-color);
  padding: 10px 15px;
  flex-wrap: wrap;

  .header-content {
    height: 60px;
  }

  .search-bar {
    width: 100%;
    .search-icon {
      .bi-search {
        display: inline;
      }

      .bi-x-lg {
        display: none;
      }
    }

    &.vs--open .search-icon {
      .bi-search {
        display: none;
      }

      .bi-x-lg {
        cursor: pointer;
        display: inline;
      }
    }
  }
}

.scroller {
  // Subtract tab bar and header
  height: calc(100vh - 61px - 120px - 18px);
  width: 100%;
  padding: 0 15px;
}
</style>
