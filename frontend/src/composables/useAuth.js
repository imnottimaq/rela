import { ref } from "vue";
import { workspaces, openWorkspaceWindows } from "./useWorkspaces";
import { openBoardWindows } from "./useBoards";
import {
  getAccessToken,
  setAuthTokens,
  clearAuthTokens,
  onAuthStateChange,
  authApi,
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

const logout = async () => {
  // 1. Call the backend to invalidate the http-only cookie.
  try {
    await authApi.logoutUser();
  } catch (error) {
    // Log the error but proceed with client-side cleanup anyway,
    // as the user's intent is to log out.
    console.error("Logout API call failed:", error);
  }

  // 2. Disable window persistence to prevent race conditions on cleanup
  try { if (typeof window !== "undefined") window.__relaDisableWindowPersistence = true; } catch (_) {}

  // 3. Clear auth tokens from storage and API client instance
  clearAuthTokens();

  // 4. Reset any in-memory state
  try {
    workspaces.value = [];
    openWorkspaceWindows.value = [];
    openBoardWindows.value = [];
  } catch (_) {}

  // 5. Clean up localStorage, preserving essential windows
  try {
    if (typeof window !== "undefined" && window.localStorage) {
      const keys = Object.keys(window.localStorage);
      const windowsToPreserve = new Set(["rela-window-login", "rela-window-register", "rela-window-main"]);
      for (const k of keys) {
        // Remove all persisted window states except for the ones we want to keep
        if (k && k.startsWith("rela-window-") && !windowsToPreserve.has(k)) {
          try {
            window.localStorage.removeItem(k);
          } catch (_) {}
        }
      }
    }
  } catch (e) {
    // ignore storage errors
  }

  // 6. Re-enable persistence after the UI settles
  try {
    setTimeout(() => {
      if (typeof window !== "undefined") window.__relaDisableWindowPersistence = false;
    }, 50); // A small delay
  } catch (_) {}
};

export function useAuth() {
  return {
    isAuthenticated,
    handleAuthSuccess,
    logout,
  };
}
