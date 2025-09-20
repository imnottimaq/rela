import { ref } from "vue";
import { onAuthStateChange } from "../utils/http";
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
  const idx = openBoardWindows.value.findIndex((w) => w.id === id);
  if (idx !== -1) {
    openBoardWindows.value.splice(idx, 1);
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
  if (typeof window === "undefined") return;
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
        const visible = Boolean(parsed?.visible);
        if (!visible) continue;

        const id = `${wsId}:${boardId}`;
        if (openBoardWindows.value.some((w) => w.id === id)) continue;

        const ws = (workspaces?.value || []).find(
          (w) => (w._id || w.id || w.Id || w.name) == wsId
        );
        const workspaceName = ws?.name || String(wsId);

        openBoardWindows.value.push({
          id,
          name: String(boardId),
          workspaceId: wsId,
          workspaceName,
          visible: true,
        });
      } catch (_) {
        // ignore parse errors
      }
    }
  } catch (_) {
    // ignore
  }
}

// React to auth changes similar to workspaces
onAuthStateChange((hasToken) => {
  if (hasToken) {
    // Defer to allow workspaces to refresh first
    setTimeout(() => restoreBoardWindowsFromStorage(), 0);
  } else {
    openBoardWindows.value = [];
  }
});
