import { ref } from "vue";

export const createWorkspaceVisible = ref(false);

export const showCreateWorkspaceWindow = () => {
  createWorkspaceVisible.value = true;
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

