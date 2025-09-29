import { ref } from 'vue';

export const editWorkspaceVisible = ref(false);
export const editingWorkspace = ref(null);

export const showEditWorkspaceWindow = (workspace) => {
  if (!workspace) {
    console.error('showEditWorkspaceWindow requires a workspace object');
    return;
  }
  editingWorkspace.value = workspace;
  editWorkspaceVisible.value = true;
  if (typeof window !== 'undefined') {
    try {
      const storageKey = `rela-window-edit-workspace-${workspace?._id || workspace?.id}`;
      window.dispatchEvent(new CustomEvent('rela:focus-window', { detail: storageKey }));
    } catch (_) {
      // no-op
    }
  }
};

export const hideEditWorkspaceWindow = () => {
  editWorkspaceVisible.value = false;
  editingWorkspace.value = null;
};

export function useEditWorkspaceWindow() {
  return {
    editWorkspaceVisible,
    editingWorkspace,
    showEditWorkspaceWindow,
    hideEditWorkspaceWindow,
  };
}
