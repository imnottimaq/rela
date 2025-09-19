import { ref } from "vue";
import { authApi, onAuthStateChange } from "../utils/http";

// Reactive state with global lifetime (one instance across app)
export const workspaces = ref([]);
export const loadingWorkspaces = ref(false);
export const workspacesError = ref("");

// Open workspace windows registry
export const openWorkspaceWindows = ref([]); // [{ id, name, visible }]

// Fetch user workspaces
export const refreshWorkspaces = async () => {
  if (loadingWorkspaces.value) return;
  loadingWorkspaces.value = true;
  workspacesError.value = "";
  try {
    const { data } = await authApi.getUserWorkspaces();
    const list = Array.isArray(data?.workspaces) ? data.workspaces : [];
    workspaces.value = list;
  } catch (err) {
    console.error("Failed to load workspaces", err);
    workspaces.value = [];
    workspacesError.value = "Failed to load workspaces";
  } finally {
    loadingWorkspaces.value = false;
  }
};

// Manage workspace windows
export const openWorkspaceWindow = (ws) => {
  if (!ws) return;
  const id = ws._id || ws.id || ws.Id || ws.name;
  const name = ws.name || String(id);
  const existing = openWorkspaceWindows.value.find((w) => w.id === id);
  if (existing) {
    existing.visible = true;
    return;
  }
  openWorkspaceWindows.value.push({ id, name, visible: true });
};

export const closeWorkspaceWindow = (id) => {
  const idx = openWorkspaceWindows.value.findIndex((w) => w.id === id);
  if (idx !== -1) {
    openWorkspaceWindows.value.splice(idx, 1);
  }
};

// React to auth state changes so list stays in sync
onAuthStateChange((hasToken) => {
  if (hasToken) {
    refreshWorkspaces();
  } else {
    workspaces.value = [];
    openWorkspaceWindows.value = [];
  }
});

export function useWorkspaces() {
  return {
    workspaces,
    loadingWorkspaces,
    workspacesError,
    refreshWorkspaces,
    openWorkspaceWindows,
    openWorkspaceWindow,
    closeWorkspaceWindow,
  };
}

