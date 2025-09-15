import { useState } from "react";
import { userAPI } from "../api";
import { getAPIErrorMessage } from "../api/errors";
import {
	validateEmail,
	validatePassword,
	validateName,
} from "../utils/validation";
import PasswordField from "../components/PasswordField";
import EmailField from "../components/EmailField";
import AvatarUpload from "../components/AvatarUpload";
import { useNavigate } from "react-router-dom";

export default function Register({ setToken }) {
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);

	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [avatar, setAvatar] = useState(null);
	const [avatarPreview, setAvatarPreview] = useState(null);
    
	const navigate = useNavigate();

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError("");
		setIsLoading(true);

		const handleError = (error) => {
			setError(error);
			setIsLoading(false);
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
		navigate("/dashboard");
	};

	return (
		<div className="flex items-center justify-center min-h-screen bg-gradient-to-r from-main-50 to-main-100">
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
							className="w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-main-500 focus:border-transparent transition"
							value={name}
							onChange={(e) => setName(e.target.value)}
							required
							disabled={isLoading}
						/>
					</div>

					<EmailField
						label={"Email"}
						email={email}
						setEmail={setEmail}
						isLoading={isLoading}
					/>

					<PasswordField
						label={"Password"}
						password={password}
						setPassword={setPassword}
						isLoading={isLoading}
					/>

					<AvatarUpload
						setAvatar={setAvatar}
						avatarPreview={avatarPreview}
						setAvatarPreview={setAvatarPreview}
						isLoading={isLoading}
					/>

					<div className="flex flex-col sm:flex-row items-center justify-between mt-6 gap-3">
						<button
							className="w-full sm:w-auto px-6 py-3 text-white bg-gradient-to-r from-main-600 to-main-700 rounded-lg shadow-lg hover:from-main-700 hover:to-main-800 disabled:opacity-50 transition"
							disabled={isLoading}
						>
							{isLoading ? "Creating account..." : "Register"}
						</button>
						<button
							type="button"
							className="text-main-600 hover:underline disabled:opacity-50 mt-2 sm:mt-0"
							onClick={() => navigate("/login")}
							disabled={isLoading}
						>
							Already have an account?
						</button>
					</div>
				</form>
			</div>
		</div>
	);
}
