import { ref, watch } from "vue";
import { onAuthStateChange, workspaceApi } from "../utils/http";
import { workspaces } from "./useWorkspaces";

// Global registry of open board windows
// Item shape: { id, name, workspaceId, workspaceName, visible }
export const openBoardWindows = ref([]);

const storageKeyForBoard = (workspaceId, board) => {
  const bId = board?._id || board?.id || board?.Id || board?.name;
  const ws = workspaceId ?? board?.workspaceId;
  if (!bId || !ws) return null;
  return `rela-window-board-${ws}-${bId}`;
};

export const openBoardWindow = (workspace, board) => {
  if (!workspace || !board) return;
  const workspaceId = workspace?._id || workspace?.id || workspace?.Id || workspace;
  const workspaceName = workspace?.name || String(workspaceId);
  const boardId = board?._id || board?.id || board?.Id || board?.name;
  const name = board?.name || String(boardId);
  if (!workspaceId || !boardId) return;

  const id = `${workspaceId}:${boardId}`;
  const existing = openBoardWindows.value.find((w) => w.id === id);
  if (existing) {
    existing.visible = true;
    if (typeof window !== "undefined") {
      try {
        const key = storageKeyForBoard(workspaceId, { _id: boardId });
        if (key) {
          window.dispatchEvent(
            new CustomEvent("rela:focus-window", { detail: key })
          );
        }
      } catch (_) {}
    }
    return;
  }
  openBoardWindows.value.push({ id, name, workspaceId, workspaceName, visible: true });
};

export const closeBoardWindow = (id) => {
  const win = openBoardWindows.value.find((w) => w.id === id);
  if (win) {
    win.visible = false;
  }
};

export function useBoards() {
  return {
    openBoardWindows,
    openBoardWindow,
    closeBoardWindow,
    restoreBoardWindowsFromStorage,
  };
}

// Restore board windows based on WindowComponent's persisted state
export function restoreBoardWindowsFromStorage() {
  if (typeof window === "undefined" || !workspaces.value || workspaces.value.length === 0) {
    return false;
  }

  let restoredSomething = false;
  try {
    const keys = Object.keys(window.localStorage || {});
    for (const key of keys) {
      const prefix = "rela-window-board-";
      if (!key.startsWith(prefix)) continue;
      const tail = key.slice(prefix.length);
      const [wsId, boardId] = tail.split("-");
      if (!wsId || !boardId) continue;
      try {
        const raw = window.localStorage.getItem(key);
        if (!raw) continue;
        const parsed = JSON.parse(raw);
        const visible = parsed?.visible !== false;

        const id = `${wsId}:${boardId}`;
        if (openBoardWindows.value.some((w) => w.id === id)) continue;

        const ws = (workspaces?.value || []).find(
          (w) => (w._id || w.id || w.Id || w.name) == wsId
        );
        const workspaceName = ws?.name || String(wsId);

        const board = (ws?.boards || []).find(
          (b) => (b._id || b.id || b.Id || b.name) == boardId
        );
        const boardName = board?.name || String(boardId);

        openBoardWindows.value.push({
          id,
          name: boardName,
          workspaceId: wsId,
          workspaceName,
          visible,
        });
        restoredSomething = true;
      } catch (_) {
        // ignore parse errors
      }
    }
  } catch (_) {
    // ignore
  }
  return restoredSomething;
}

// React to auth changes. This is complex because we must wait for workspaces to be
// loaded before we can restore board windows (as we need the names).
onAuthStateChange((hasToken) => {
  if (hasToken) {
    // Attempt to restore immediately. If workspaces aren't loaded, this will fail gracefully.
    const restored = restoreBoardWindowsFromStorage();

    // If restoration didn't happen, it means workspaces were not ready. We set up a watcher.
    if (!restored) {
      const unwatch = watch(workspaces, () => {
        // When workspaces are populated, try restoring again.
        if (restoreBoardWindowsFromStorage()) {
          // If we succeed, we can stop watching.
          unwatch();
        }
      }, { deep: true });
    }
  } else {
    // If the user logs out, clear all open board windows.
    openBoardWindows.value = [];
  }
});

// New composable for fetching tasks for a specific board
export function useBoardTasks(workspaceIdRef, boardIdRef) {
  const tasks = ref([]);
  const isLoading = ref(false);
  const error = ref(null);
  const isNotFound = ref(false);

  const fetchTasks = async () => {
    const boardId = boardIdRef.value;
    const workspaceId = workspaceIdRef.value;
    if (!boardId || !workspaceId) {
      tasks.value = [];
      return;
    }
    isLoading.value = true;
    error.value = null;
    isNotFound.value = false;
    try {
      const { data } = await workspaceApi.getTasks(workspaceId, boardId);
      const list = Array.isArray(data?.tasks) ? data.tasks : (Array.isArray(data) ? data : []);
      tasks.value = list;
      if (tasks.value.length === 0) {
        isNotFound.value = true;
      }
    } catch (err) {
      console.error(`Failed to fetch tasks for board ${boardId} in workspace ${workspaceId}:`, err);
      error.value = err;
      tasks.value = [];
      if (err.response?.status === 404) {
        isNotFound.value = true;
      }
    } finally {
      isLoading.value = false;
    }
  };

  watch([workspaceIdRef, boardIdRef], fetchTasks, { immediate: true });

  return {
    tasks,
    isLoading,
    error,
    isNotFound,
    fetchTasks,
  };
}
