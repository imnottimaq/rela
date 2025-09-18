import { ref } from "vue";

export const loginVisible = ref(false);

export const showLoginWindow = () => {
  loginVisible.value = true;
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
