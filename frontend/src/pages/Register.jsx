import { useState } from "react";
import { userAPI } from "../api";
import { getAPIErrorMessage } from "../api/errors";
import {
	validateEmail,
	validatePassword,
	validateName,
} from "../utils/validation";

const Register = ({ setToken, switchToLogin }) => {
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);

	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [avatar, setAvatar] = useState(null);
	const [avatarPreview, setAvatarPreview] = useState(null);

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError("");
		setIsLoading(true);

		const handleError = (error) => {
			setError(error);
			setIsLoading(false);
			return;
		};

		if (!validateName(name)) return handleError("Please enter your name.");
		if (!validateEmail(email)) return handleError("Invalid email address.");
		if (!validatePassword(password))
			return handleError(
				"Password must be at least 8 characters and include uppercase, lowercase, number, and special character."
			);

		const result = await userAPI.register(name, email, password);

		if (!result.ok)
			return handleError(getAPIErrorMessage(result, "register"));
		if (!result.data?.token)
			return setError("Invalid response from server. Please try again.");

		setToken(result.data.token);
        localStorage.setItem("token", result.data.token);

		if (avatar) {
			const uploadRes = await userAPI.uploadAvatar(
				avatar,
				result.data.token
			);
			if (!uploadRes.ok)
				console.error("Avatar upload failed", uploadRes.error);
		}

		setIsLoading(false);
	};

	return (
		<div className="flex items-center justify-center min-h-screen bg-gradient-to-r from-blue-50 to-indigo-50">
			<div className="px-10 py-8 mt-4 bg-white shadow-2xl rounded-2xl w-full max-w-md border border-gray-200">
				<h3 className="text-3xl font-extrabold text-center text-gray-800 mb-6">
					Create a new account
				</h3>

				{error && (
					<div className="mt-4 p-3 bg-red-100 text-red-700 rounded-md border border-red-200 shadow-sm">
						{error}
					</div>
				)}

				<form onSubmit={handleSubmit} className="space-y-5">
					{/* Name */}
					<div>
						<label
							className="block text-gray-700 font-medium"
							htmlFor="name"
						>
							Name
						</label>
						<input
							type="text"
							placeholder="John Doe"
							className="w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
							value={name}
							onChange={(e) => setName(e.target.value)}
							required
							disabled={isLoading}
						/>
					</div>

					{/* Email */}
					<div>
						<label
							className="block text-gray-700 font-medium"
							htmlFor="email"
						>
							Email
						</label>
						<input
							type="email"
							placeholder="you@example.com"
							className="w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
							value={email}
							onChange={(e) => setEmail(e.target.value)}
							required
							disabled={isLoading}
						/>
					</div>

					{/* Password */}
					<div>
						<label className="block text-gray-700 font-medium">
							Password
						</label>
						<input
							type="password"
							placeholder="********"
							className="w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
							required
							disabled={isLoading}
							minLength={8}
						/>
						<p className="text-xs text-gray-400 mt-1">
							Password must be at least 8 characters
						</p>
					</div>

					{/* Avatar Upload */}
					<div className="flex flex-col items-center">
						<label className="block text-gray-700 font-medium mb-2">
							Avatar (optional)
						</label>
						<div className="relative w-24 h-24 rounded-full overflow-hidden border-2 border-gray-300 flex items-center justify-center cursor-pointer hover:border-blue-500 transition">
							{avatarPreview ? (
								<img
									src={avatarPreview}
									alt="Avatar Preview"
									className="w-full h-full object-cover"
								/>
							) : (
								<span className="text-gray-400 text-sm text-center">
									Click to upload
								</span>
							)}
							<input
								type="file"
								accept="image/*"
								className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
								onChange={(e) => {
									const file = e.target.files[0];
									setAvatar(file);
									if (file) {
										const reader = new FileReader();
										reader.onload = () =>
											setAvatarPreview(reader.result);
										reader.readAsDataURL(file);
									}
								}}
								disabled={isLoading}
							/>
						</div>
					</div>

					{/* Buttons */}
					<div className="flex flex-col sm:flex-row items-center justify-between mt-6 gap-3">
						<button
							className="w-full sm:w-auto px-6 py-3 text-white bg-gradient-to-r from-blue-600 to-indigo-600 rounded-lg shadow-lg hover:from-blue-700 hover:to-indigo-700 disabled:opacity-50 transition"
							disabled={isLoading}
						>
							{isLoading ? "Creating account..." : "Register"}
						</button>
						<button
							type="button"
							className="text-blue-600 hover:underline disabled:opacity-50 mt-2 sm:mt-0"
							onClick={switchToLogin}
							disabled={isLoading}
						>
							Already have an account?
						</button>
					</div>
				</form>
			</div>
		</div>
	);
};

export default Register;
