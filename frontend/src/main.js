import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import piniaPersist from "pinia-plugin-persist";
import "./style.scss";

import Words from "./components/Words.vue";
import Sign from "./components/Sign.vue";
import Categories from "./components/Categories.vue";
import CategoryWords from "./components/CategoryWords.vue";
import Favorites from "./components/Favorites.vue";
import About from "./components/About.vue";
import messages from "@intlify/unplugin-vue-i18n/messages";
import { setupI18n, setI18nLanguage, SUPPORT_LOCALES } from "./i18n";

import App from "./App.vue";

const pinia = createPinia();
pinia.use(piniaPersist);

const routes = [
  { path: "/", component: Words },
  { path: "/signs/:id", component: Sign, name: "Sign" },
  { path: "/favorites", component: Favorites },
  { path: "/categories", component: Categories },
  { path: "/categories/:id", component: CategoryWords, name: "CategoryWords" },
  { path: "/about", component: About },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

const i18n = setupI18n({ locale: "sv", legacy: false, messages });

router.beforeEach(async (to, from, next) => {
  let paramsLocale = to.params.locale || "sv";

  // use locale if paramsLocale is not in SUPPORT_LOCALES
  if (!SUPPORT_LOCALES.includes(paramsLocale)) {
    paramsLocale = "sv";
  }

  // load locale messages
  if (!i18n.global.availableLocales.includes(paramsLocale)) {
    await loadLocaleMessages(i18n, paramsLocale);
  }

  // set i18n language
  setI18nLanguage(i18n, paramsLocale);

  return next();
});

createApp(App).use(router).use(i18n).use(pinia).mount("#app");
