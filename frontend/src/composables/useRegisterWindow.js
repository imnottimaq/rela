import { ref } from "vue";

export const registerVisible = ref(false);

export const showRegisterWindow = () => {
  registerVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", { detail: "rela-window-register" })
      );
    } catch (_) {}
  }
};

export const hideRegisterWindow = () => {
  registerVisible.value = false;
};

export function useRegisterWindow() {
  return {
    registerVisible,
    showRegisterWindow,
    hideRegisterWindow,
  };
}
