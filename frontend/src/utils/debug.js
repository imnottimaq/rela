export const debugAPI = {
	logRequest: (method, url, data, headers) => {
		if (import.meta.env.DEV) {
			console.log("API Request:", {
				method,
				url,
				data: data ? JSON.parse(data) : undefined,
				headers: {
					...headers,
					"Authorization": headers?.["Authorization"]
						? "[REDACTED]"
						: undefined,
				},
			});
		}
	},

	logResponse: (response, data) => {
		if (import.meta.env.DEV) {
			console.log("API Response:", {
				status: response.status,
				statusText: response.statusText,
				url: response.url,
				data,
			});
		}
	},

	logError: (error) => {
		if (import.meta.env.DEV) {
			console.error("API Error Details:", {
				name: error.name,
				message: error.message,
				status: error.status,
				statusText: error.statusText,
			});

			if (error.body) {
				console.error("API Error Body:", error.body);
			}

			console.error("Full Error Object:", error);
		}
	},
};
