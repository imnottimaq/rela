/**
 * Map backend error response to user-friendly message.
 * @param {Object} result - API result object from safeFetch
 * @param {"login"|"register"|"default"} context - type of API call
 * @returns {string} Friendly message for user
 */
export const getAPIErrorMessage = (result, context = "default") => {
	if (!result || !result.error)
		return "Something went wrong. Please try again.";

	const { status, body, message } = result.error;

	switch (context) {
		case "login":
			switch (status) {
				case 400:
					return (
						"Something went wrong. Check your email or password."
					);
				default:
					return "Something went wrong. Please try again.";
			}

		case "register":
			switch (status) {
				case 400:
					return (
						"Invalid input. Check your details and try again."
					);
				default:
					return "Something went wrong. Please try again.";
			}

		case "default":
		default:
			return (
				body?.error ||
				message ||
				"Something went wrong. Please try again."
			);
	}
};
