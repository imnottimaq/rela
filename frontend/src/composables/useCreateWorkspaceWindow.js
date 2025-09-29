import { ref } from "vue";

export const createWorkspaceVisible = ref(false);

export const showCreateWorkspaceWindow = () => {
  createWorkspaceVisible.value = true;
  if (typeof window !== "undefined") {
    try {
      window.dispatchEvent(
        new CustomEvent("rela:focus-window", {
          detail: "rela-window-create-workspace",
        })
      );
    } catch (_) {}
  }
};

export const hideCreateWorkspaceWindow = () => {
  createWorkspaceVisible.value = false;
};

export function useCreateWorkspaceWindow() {
  return {
    createWorkspaceVisible,
    showCreateWorkspaceWindow,
    hideCreateWorkspaceWindow,
  };
}
