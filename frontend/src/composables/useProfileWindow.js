import { ref } from "vue";

const profileVisible = ref(false);

export const showProfileWindow = () => {
  profileVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", { detail: "rela-window-profile" })
      );
    } catch (_) {}
  }
};

export const hideProfileWindow = () => {
  profileVisible.value = false;
};

export function useProfileWindow() {
  return {
    profileVisible,
    showProfileWindow,
    hideProfileWindow,
  };
}
