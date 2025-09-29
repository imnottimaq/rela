import { ref } from "vue";
import {authApi, onAuthStateChange, workspaceApi} from "../utils/http";
import { useClipboard } from '@vueuse/core'

// Reactive state with global lifetime (one instance across app)
export const workspaces = ref([]);
export const loadingWorkspaces = ref(false);
export const workspacesError = ref("");

// Open workspace windows registry
export const openWorkspaceWindows = ref([]); // [{ id, name, visible }]

// Fetch user workspaces
export const refreshWorkspaces = async (options = {}) => {
  const { skipRestore = false } = options || {};
  if (loadingWorkspaces.value) return;
  loadingWorkspaces.value = true;
  workspacesError.value = "";
  try {
    const { data } = await authApi.getUserWorkspaces();
    const list = Array.isArray(data?.workspaces) ? data.workspaces : [];
    workspaces.value = list;
    if (!skipRestore) {
      restoreWorkspaceWindowsFromStorage();
    }
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
    if (typeof window !== "undefined") {
      try {
        const key = id ? `rela-window-workspace-${id}` : null;
        if (key) {
          window.dispatchEvent(
            new CustomEvent("rela:focus-window", { detail: key })
          );
        }
      } catch (_) {}
    }
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

const { copy } = useClipboard()
export const createWorkspaceInviteLink = async (workspaceId) => {
  if (!workspaceId) return;
  const {data} = await workspaceApi.getInvite(workspaceId)
  const token = data?.sqid;
  const invite_url = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") + "/workspaces/add/" + token;
  copy(invite_url)
}

// React to auth state changes so list stays in sync
onAuthStateChange((hasToken) => {
  if (hasToken) {
    refreshWorkspaces();
  } else {
    workspaces.value = [];
    openWorkspaceWindows.value = [];
  }
});

// Helpers to keep window visibility persisted based on WindowComponent storage
const storageKeyForWorkspace = (ws) => {
  const id = ws?._id || ws?.id || ws?.Id || ws?.name;
  return id ? `rela-window-workspace-${id}` : null;
};

const restoreWorkspaceWindowsFromStorage = () => {
  if (typeof window === "undefined") return;
  const existingIds = new Set(openWorkspaceWindows.value.map((w) => w.id));
  for (const ws of workspaces.value || []) {
    const key = storageKeyForWorkspace(ws);
    if (!key) continue;
    try {
      const raw = window.localStorage.getItem(key);
      if (!raw) continue;
      const parsed = JSON.parse(raw);
      const visible = Boolean(parsed?.visible);
      if (visible) {
        const id = ws._id || ws.id || ws.Id || ws.name;
        if (!existingIds.has(id)) {
          openWorkspaceWindows.value.push({ id, name: ws.name || String(id), visible: true });
          existingIds.add(id);
        }
      }
    } catch (_) {
      // ignore parsing errors
    }
  }
};

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
