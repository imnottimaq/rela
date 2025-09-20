import { ref } from "vue";

export const loginVisible = ref(false);

export const showLoginWindow = () => {
  loginVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", { detail: "rela-window-login" })
      );
    } catch (_) {
      // no-op
    }
  }
};

export const hideLoginWindow = () => {
  loginVisible.value = false;
};

export function useLoginWindow() {
  return {
    loginVisible,
    showLoginWindow,
    hideLoginWindow,
  };
}
