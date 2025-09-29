import { ref } from "vue";
import {authApi} from "../utils/http.js";

// Visibility state for the logout confirmation window
export const logoutConfirmVisible = ref(false);

// Optional callback to run after a successful logout
const onConfirmedRef = ref(null);

export const showConfirmLogoutWindow = (onConfirmed) => {
  if (typeof onConfirmed === "function") {
    onConfirmedRef.value = onConfirmed;
  } else {
    onConfirmedRef.value = null;
  }
  logoutConfirmVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", {
          detail: "rela-window-confirm-logout",
        })
      );
    } catch (_) {}
  }
};

export const hideConfirmLogoutWindow = () => {
  logoutConfirmVisible.value = false;
  onConfirmedRef.value = null;
};

export function useConfirmLogoutWindow() {
  return {
    logoutConfirmVisible,
    showConfirmLogoutWindow,
    hideConfirmLogoutWindow,
    // Internal use by the window component to get the callback
    __onConfirmedRef: onConfirmedRef,
  };
}
