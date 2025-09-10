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

const createApi = () => {
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

            // just call the actual fetch without debug
            return api[method](url, requestOptions);
        };
    });

    return wrappedApi;
};

// const api = createApiWithDebug();
const api = createApi();

export const userAPI = {
	login: (email, password) =>
		api.post(
			"users/login",
			{ email, password },
			{ headers: { "Content-Type": "application/json" } }
		),

	register: (name, email, password) =>
		api.post(
			"users/create",
			{ name, email, password },
			{ headers: { "Content-Type": "application/json" } }
		),

	delete: (email, password, token) =>
		api.delete(
			"users/delete",
			{ email, password },
			{ headers: { Authorization: token } }
		),

	refresh: () => api.get("users/refresh"),
};

export const boardAPI = {
	getAll: (token) =>
		api.get("boards", {}, { headers: { Authorization: token } }),

	create: (name, token) =>
		api.post("boards", { name }, { headers: { Authorization: token } }),

	update: (boardId, name, token) =>
		api.patch(
			`boards/${boardId}`,
			{ name },
			{ headers: { Authorization: token } }
		),

	delete: (boardId, token) =>
		api.delete(
			`boards/${boardId}`,
			{},
			{ headers: { Authorization: token } }
		),
};

export const taskAPI = {
	getAll: (token) =>
		api.get("tasks", {}, { headers: { Authorization: token } }),

	create: (taskData, token) =>
		api.post("tasks", { taskData }, { headers: { Authorization: token } }),

	update: (taskId, taskData, token) =>
		api.patch(
			`tasks/${taskId}`,
			{ taskData },
			{ headers: { Authorization: token } }
		),

	delete: (taskId, token) =>
		api.delete(
			`tasks/${taskId}`,
			{},
			{ headers: { Authorization: token } }
		),
};

export default api;
