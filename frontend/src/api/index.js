import { createSafeFetch } from "@asouei/safe-fetch";
import { debugAPI } from "../utils/debug";

const createApiWithDebug = () => {
	const api = createSafeFetch({
		baseURL: `${window.location.origin}/api/v1/`,
		timeoutMs: 10000,
		retries: {
			retries: 2,
			baseDelayMs: 300,
		},
	});

	const wrappedApi = {};

	["get", "post", "put", "patch", "delete"].forEach((method) => {
		wrappedApi[method] = async (url, options = {}) => {
			const headers = {
				"Content-Type": "application/json",
				...options.headers,
			};

			const requestOptions = {
				...options,
				headers,
			};

			debugAPI.logRequest(
				method.toUpperCase(),
				url,
				requestOptions.body,
				requestOptions.headers
			);

			try {
				const result = await api[method](url, requestOptions);

				if (result.ok) {
					debugAPI.logResponse(result.response, result.data);
				} else {
					debugAPI.logError(result.error);
				}

				return result;
			} catch (error) {
				debugAPI.logError(error);
				throw error;
			}
		};
	});

	return wrappedApi;
};

const api = createApiWithDebug();

const getAuthHeaders = (token) => ({
	"X-Authorization": `Bearer ${token}`,
	"Content-Type": "application/json",
});

export const userAPI = {
	login: (email, password) =>
		api.post("users/login", {
			body: JSON.stringify({ email, password }),
		}),

	register: (name, email, password) =>
		api.post("users/create", {
			body: JSON.stringify({ name, email, password }),
		}),

	delete: (email, password, token) =>
		api.delete("users/delete", {
			body: JSON.stringify({ email, password }),
			headers: getAuthHeaders(token),
		}),

	refresh: () => api.get("users/refresh"),
};

export const boardAPI = {
	getAll: (token) =>
		api.get("boards", {
			headers: getAuthHeaders(token),
		}),

	create: (name, token) =>
		api.post("boards", {
			body: JSON.stringify({ name }),
			headers: getAuthHeaders(token),
		}),

	update: (boardId, name, token) =>
		api.patch(`boards/${boardId}`, {
			body: JSON.stringify({ name }),
			headers: getAuthHeaders(token),
		}),

	delete: (boardId, token) =>
		api.delete(`boards/${boardId}`, {
			headers: getAuthHeaders(token),
		}),
};

export const taskAPI = {
	getAll: (token) =>
		api.get("tasks", {
			headers: getAuthHeaders(token),
		}),

	create: (taskData, token) =>
		api.post("tasks", {
			body: JSON.stringify(taskData),
			headers: getAuthHeaders(token),
		}),

	update: (taskId, taskData, token) =>
		api.patch(`tasks/${taskId}`, {
			body: JSON.stringify(taskData),
			headers: getAuthHeaders(token),
		}),

	delete: (taskId, token) =>
		api.delete(`tasks/${taskId}`, {
			headers: getAuthHeaders(token),
		}),
};

export default api;
