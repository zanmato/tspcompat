import { defineStore } from "pinia";
import { settings } from "@/api.js";

export const useWordsStore = defineStore("words", {
  state: () => ({
    words: [],
  }),
  actions: {
    async fetchWords() {
      if (this.words.length > 0) {
        return;
      }

      const res = await fetch(settings.baseURL + "/api/words");
      this.words = await res.json();
    },
  },
});
