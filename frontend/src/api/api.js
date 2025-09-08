export const loginUser = async (email, password) => {
	const res = await fetch("/api/users/login", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password }),
	});
	return res.json();
};

export const getBoards = async (token) => {
	const res = await fetch("/api/boards", {
		headers: { Authorization: `Bearer ${token}` },
	});
	return res.json();
};

export const getTasks = async (token) => {
	const res = await fetch("/api/tasks", {
		headers: { Authorization: `Bearer ${token}` },
	});
	return res.json();
};

export const updateTaskBoard = async (id, board, token) => {
	return fetch(`/api/tasks/${id}`, {
		method: "PATCH",
		headers: {
			"Content-Type": "application/json",
			Authorization: `Bearer ${token}`,
		},
		body: JSON.stringify({ board }),
	});
};
