import axios from "axios";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "");
const ACCESS_TOKEN_KEY = "rela.access_token";
const REFRESH_TOKEN_KEY = "rela.refresh_token";
const DEFAULT_REFRESH_PATH = "/users/refresh";
const REFRESH_PATH = import.meta.env.VITE_API_REFRESH_PATH || DEFAULT_REFRESH_PATH;

const authStateListeners = new Set();
const RETRY_FLAG = "__relaRetry403";

const notifyAuthStateChange = (hasToken) => {
  authStateListeners.forEach((listener) => {
    try {
      listener(Boolean(hasToken));
    } catch (listenerError) {
      console.error("Auth state listener error", listenerError);
    }
  });
};

export const onAuthStateChange = (listener) => {
  if (typeof listener !== "function") {
    return () => {};
  }
  authStateListeners.add(listener);
  return () => {
    authStateListeners.delete(listener);
  };
};

const apiClient = axios.create({
  baseURL: API_BASE_URL || "",
  withCredentials: true
});

let refreshRequest = null;

const getSafeStorage = () => {
  if (typeof window === "undefined") {
    return null;
  }
  return window.localStorage;
};

export const getAccessToken = () => {
  const storage = getSafeStorage();
  return storage ? storage.getItem(ACCESS_TOKEN_KEY) : null;
};

export const getRefreshToken = () => {
  const storage = getSafeStorage();
  return storage ? storage.getItem(REFRESH_TOKEN_KEY) : null;
};

const setDefaultAuthHeader = (token) => {
  if (token) {
    apiClient.defaults.headers.common.Authorization = `Bearer ${token}`;
    apiClient.defaults.headers.common["X-Authorization"] = token;
    return;
  }
  delete apiClient.defaults.headers.common.Authorization;
  delete apiClient.defaults.headers.common["X-Authorization"];
};

export const setAuthTokens = ({ accessToken, refreshToken }) => {
  const storage = getSafeStorage();
  if (!storage) {
    if (typeof accessToken !== "undefined") {
      setDefaultAuthHeader(accessToken || null);
    }
    notifyAuthStateChange(Boolean(accessToken));
    return;
  }
  if (accessToken) {
    storage.setItem(ACCESS_TOKEN_KEY, accessToken);
  }
  if (refreshToken) {
    storage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
  const resultingAccessToken = storage.getItem(ACCESS_TOKEN_KEY);
  setDefaultAuthHeader(resultingAccessToken);
  notifyAuthStateChange(Boolean(resultingAccessToken));
};

export const clearAuthTokens = () => {
  const storage = getSafeStorage();
  if (!storage) {
    setDefaultAuthHeader(null);
    notifyAuthStateChange(false);
    return;
  }
  storage.removeItem(ACCESS_TOKEN_KEY);
  storage.removeItem(REFRESH_TOKEN_KEY);
  setDefaultAuthHeader(null);
  notifyAuthStateChange(false);
};

const applyAuthHeader = (config, token) => {
  if (!token) {
    return config;
  }
  const authHeader = `Bearer ${token}`;
  config.headers = {
    ...(config.headers || {}),
    Authorization: authHeader,
    "X-Authorization": token,
  };
  return config;
};

const requestTokenRefresh = async () => {
  if (refreshRequest) {
    return refreshRequest;
  }

  const url = `${API_BASE_URL || ""}${REFRESH_PATH}`;
  const storedRefreshToken = getRefreshToken();
  const requestConfig = {
    withCredentials: true,
  };
  if (storedRefreshToken) {
    requestConfig.headers = { Authorization: `Bearer ${storedRefreshToken}` };
  }

  refreshRequest = axios
    .get(url, requestConfig)
    .then(({ data }) => {
      const newAccessToken = data?.token;
      const newRefreshToken = data?.refreshToken || storedRefreshToken;
      if (!newAccessToken) {
        throw new Error("Invalid refresh token response");
      }
      setAuthTokens({ accessToken: newAccessToken, refreshToken: newRefreshToken });
      return newAccessToken;
    })
    .catch((error) => {
      clearAuthTokens();
      throw error;
    })
    .finally(() => {
      refreshRequest = null;
    });

  return refreshRequest;
};

apiClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  return applyAuthHeader(config, token);
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const { config, response } = error;
    if (!response || !config) {
      return Promise.reject(error);
    }

    if (response.status === 403 && !config[RETRY_FLAG]) {
      config[RETRY_FLAG] = true;
      try {
        const newToken = await requestTokenRefresh();
        const updatedConfig = applyAuthHeader(config, newToken);
        return apiClient(updatedConfig);
      } catch (refreshError) {
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export const authApi = {
  login(email, password) {
    return apiClient.post("/users/login", { email, password });
  },
  createUser(name, email, password) {
    return apiClient.post("/users/create", { name, email, password });
  },
  deleteUser(payload) {
    return apiClient.delete("/users/delete", { data: payload });
  },
  getUserInfo() {
    return apiClient.get("/users/get_info");
  },
  updateUserInfo(payload) {
    return apiClient.patch("/users/update_info", payload);
  },
  getUserWorkspaces() {
    return apiClient.get("/users/workspaces");
  },
  uploadAvatar(formData) {
    return apiClient.post("/users/upload_avatar", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
  },
  logoutUser(){
    return apiClient.post("/users/logout");
  }
};

export const workspaceApi = {
  createWorkspace(payload) {
    return apiClient.post("/workspaces/create", payload);
  },
  deleteWorkspace(workspaceId) {
    return apiClient.delete(`/workspaces/${workspaceId}/`);
  },
  updateWorkspace(workspaceId, payload) {
    return apiClient.patch(`/workspaces/${workspaceId}/`, payload);
  },
  getBoards(workspaceId) {
    return apiClient.get(`/workspaces/${workspaceId}/boards`);
  },
  getBoard(workspaceId, boardId) {
    return apiClient.get(`/workspaces/${workspaceId}/boards/${boardId}`);
  },
  createBoard(workspaceId, payload) {
    return apiClient.post(`/workspaces/${workspaceId}/boards`, payload);
  },
  updateBoard(workspaceId, boardId, payload) {
    return apiClient.patch(`/workspaces/${workspaceId}/boards/${boardId}`, payload);
  },
  deleteBoard(workspaceId, boardId) {
    return apiClient.delete(`/workspaces/${workspaceId}/boards/${boardId}`);
  },
  getTasks(workspaceId, boardId) {
    return apiClient.get(`/workspaces/${workspaceId}/tasks/${boardId}`);
  },
  createTask(workspaceId, payload) {
    return apiClient.post(`/workspaces/${workspaceId}/tasks`, payload);
  },
  updateTask(workspaceId, taskId, payload) {
    return apiClient.patch(`/workspaces/${workspaceId}/tasks/${taskId}`, payload);
  },
  deleteTask(workspaceId, taskId) {
    return apiClient.delete(`/workspaces/${workspaceId}/tasks/${taskId}`);
  },
  assignTask(workspaceId, payload) {
    return apiClient.post(`/workspaces/${workspaceId}/assign`, payload);
  },
  getMembers(workspaceId) {
    return apiClient.get(`/workspaces/${workspaceId}/members`);
  },
  kickMember(workspaceId, memberId) {
    return apiClient.delete(`/workspaces/${workspaceId}/kick`, {
      data: { userId: memberId },
    });
  },
  getInvite(workspaceId) {
    return apiClient.get(`/workspaces/${workspaceId}/new_invite`);
  },
  getWorkspaceByInviteToken(joinToken) {
    return apiClient.get(`/workspaces/invite/${joinToken}`);
  },
  promoteMember(workspaceId, userId, payload) {
    return apiClient.patch(`/workspaces/${workspaceId}/promote/${userId}`, payload);
  },
  acceptInvite(joinToken) {
    return apiClient.post(`/workspaces/invite/accept/${joinToken}`);
  },
  uploadAvatar(workspaceId,formData) {
    return apiClient.post(`/workspaces/${workspaceId}/upload_avatar`, formData);
  }
};

export const rawApiClient = apiClient;

export default apiClient;
