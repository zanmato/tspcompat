import { defineStore } from 'pinia'

export const useFavoritesStore = defineStore('favorites', {
  state: () => ({
    favorites: {},
  }),
  getters: {
    isFavorite: (state) => (id) => {
      return id in state.favorites;
    },
  },
  actions: {
    toggleFavorite(sign) {
      if (sign.id in this.favorites) {
        delete this.favorites[sign.id];
        return;
      }

      this.favorites[sign.id] = sign;
    },
  },
  persist: {
    enabled: true
  }
})