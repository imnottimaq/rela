import { ref } from 'vue';

export const aboutRelaVisible = ref(false);

export const showAboutRelaWindow = () => {
  aboutRelaVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
          new CustomEvent("rela:focus-window", { detail: "rela-window-about" })
      );
    } catch (_) {
      // no-op
    }
  }
};
export const hideAboutRelaWindow = () => {
  aboutRelaVisible.value = false;
};

export function useAboutRelaWindow() {

  return {
    aboutRelaVisible,
    showAboutRelaWindow,
    hideAboutRelaWindow,
  };
}
