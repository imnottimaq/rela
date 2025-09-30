import { ref } from "vue";

// Visibility state for the delete account confirmation window
export const deleteAccountConfirmVisible = ref(false);

// Optional callback to run after a successful account deletion
const onConfirmedRef = ref(null);

// User email passed from ProfileWindow
const userEmailRef = ref(null);

export const showConfirmDeleteAccountWindow = (onConfirmed, userEmail) => {
  if (typeof onConfirmed === "function") {
    onConfirmedRef.value = onConfirmed;
  } else {
    onConfirmedRef.value = null;
  }
  userEmailRef.value = userEmail || null;
  deleteAccountConfirmVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", {
          detail: "rela-window-confirm-delete-account",
        })
      );
    } catch (_) {}
  }
};

export const hideConfirmDeleteAccountWindow = () => {
  deleteAccountConfirmVisible.value = false;
  onConfirmedRef.value = null;
  userEmailRef.value = null;
};

export function useConfirmDeleteAccountWindow() {
  return {
    deleteAccountConfirmVisible,
    showConfirmDeleteAccountWindow,
    hideConfirmDeleteAccountWindow,
    // Internal use by the window component to get the callback and email
    __onConfirmedRef: onConfirmedRef,
    __userEmailRef: userEmailRef,
  };
}
