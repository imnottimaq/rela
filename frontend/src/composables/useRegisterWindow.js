import { ref } from "vue";

export const registerVisible = ref(false);

export const showRegisterWindow = () => {
  registerVisible.value = true;
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
