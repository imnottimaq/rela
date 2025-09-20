import { ref } from "vue";
import { workspaces, openWorkspaceWindows } from "./useWorkspaces";
import { openBoardWindows } from "./useBoards";
import {
  getAccessToken,
  setAuthTokens,
  clearAuthTokens,
  onAuthStateChange,
} from "../utils/http";

const isAuthenticated = ref(Boolean(getAccessToken()));

onAuthStateChange((hasToken) => {
  isAuthenticated.value = hasToken;
});

const handleAuthSuccess = ({ token, refreshToken }) => {
  if (!token) {
    return false;
  }
  // When logging into a new account, purge any UI state persisted from a previous user
  try {
    if (typeof window !== "undefined") {
      window.__relaDisableWindowPersistence = true;
    }
    // Reset in-memory lists so the UI doesn't show stale data between auth flips
    workspaces.value = [];
    openWorkspaceWindows.value = [];
    openBoardWindows.value = [];
    // Remove any persisted window states to avoid restoring from a previous user
    if (typeof window !== "undefined" && window.localStorage) {
      const keys = Object.keys(window.localStorage);
      for (const k of keys) {
        if (k && k.startsWith("rela-window-")) {
          try { window.localStorage.removeItem(k); } catch (_) {}
        }
      }
    }
  } catch (_) {}
  finally {
    try { setTimeout(() => { if (typeof window !== "undefined") window.__relaDisableWindowPersistence = false; }, 0); } catch (_) {}
  }
  setAuthTokens({ accessToken: token, refreshToken });
  return true;
};

const logout = () => {
  console.log("Logout placeholder: send HTTP request later");
  // Clear auth headers/tokens and notify listeners
  try { if (typeof window !== "undefined") window.__relaDisableWindowPersistence = true; } catch (_) {}
  clearAuthTokens();
  // Proactively reset in-memory state for UI
  try {
    workspaces.value = [];
    openWorkspaceWindows.value = [];
    openBoardWindows.value = [];
  } catch (_) {}
  // Then clear the entire localStorage as requested
  try {
    if (typeof window !== "undefined" && window.localStorage) {
      let mainWindow = window.localStorage.getItem("rela-window-main");
      window.localStorage.clear();
        // Restore main window state if it existed before
        if (mainWindow) {
            try { window.localStorage.setItem("rela-window-main", mainWindow); } catch (_) {}
        }
    }
  } catch (e) {
    // ignore storage errors
  }
  // Re-enable persistence after the UI settles
  try { setTimeout(() => { if (typeof window !== "undefined") window.__relaDisableWindowPersistence = false; }, 0); } catch (_) {}
};

export function useAuth() {
  return {
    isAuthenticated,
    handleAuthSuccess,
    logout,
  };
}
