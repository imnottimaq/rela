import { ref } from "vue";

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

