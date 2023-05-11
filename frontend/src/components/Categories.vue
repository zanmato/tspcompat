<script setup>
import { reactive } from "vue";
import { RecycleScroller } from "vue-virtual-scroller";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
import ChevronRightIcon from "bootstrap-icons/icons/chevron-right.svg?component";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { settings } from "@/api.js";

const state = reactive({
  categories: [],
});

const fetchCategories = () => {
  fetch(settings.baseURL + "/api/categories")
    .then((response) => response.json())
    .then((data) => {
      state.categories = data;
    });
};

fetchCategories();

const router = useRouter();
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <div class="header">
    <div class="header-content">
      <h1>{{ t("categories") }}</h1>
    </div>
  </div>
  <RecycleScroller
    class="scroller"
    :items="state.categories"
    :item-size="40"
    key-field="0"
    v-slot="{ item }"
  >
    <div
      class="category"
      @click="router.push({ name: 'CategoryWords', params: { id: item[0] } })"
    >
      <span>{{ item[1] }}</span>
      <div class="count">{{ item[2] }}</div>
      <ChevronRightIcon />
    </div>
  </RecycleScroller>
</template>

<style lang="scss" scoped>
.scroller {
  height: calc(100vh - 61px - 80px);
  width: 100%;
  padding: 0 15px;
}

.category {
  height: 40px;
  line-height: 40px;
  border-bottom: 1px solid var(--separator-color);
  text-align: left;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;

  span {
    flex: 1;
    text-overflow: ellipsis;
    white-space: nowrap;
    overflow: hidden;
    margin-right: 10px;
  }

  .count {
    height: 24px;
    line-height: 14px;
    padding: 4px 10px;
    font-size: 0.8rem;
    border: 1px solid var(--tint-color);
    color: var(--tint-color);
    border-radius: 5px;
  }

  .bi {
    margin: 0 5px;
  }
}
</style>
