export const validateEmail = (email) => {
	const emailRegex = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i;
	return emailRegex.test(email);
};

export const validatePassword = (password) => {
	if (!password || password.length < 8) return false;

	const hasUpper = /[A-Z]/.test(password);
	const hasLower = /[a-z]/.test(password);
	const hasDigit = /\d/.test(password);
	const hasSpecial = /[@$!%*?&]/.test(password);

	return hasUpper && hasLower && hasDigit && hasSpecial;
};

export const validateName = (name) => {
	return !!name && name.trim().length > 0;
};

export const validateEntityName = (name) => {
	const entityRegex = /^[A-Za-z]+(?:[ '-][A-Za-z]+)*$/;
	return entityRegex.test(name);
};
