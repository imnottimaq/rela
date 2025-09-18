import { ref } from "vue";

export const loginVisible = ref(false);

export const showLoginWindow = () => {
  loginVisible.value = true;
  window.location.hash = "#login-window";
};

export const hideLoginWindow = () => {
  loginVisible.value = false;
    window.location.hash = "";
};

export function useLoginWindow() {
  return {
    loginVisible,
    showLoginWindow,
    hideLoginWindow,
  };
}
