import { useState } from "react";
import { userAPI } from "../api";
import { getAPIErrorMessage } from "../api/errors";
import PasswordField from "../components/PasswordField";
import EmailField from "../components/EmailField";

export default function Login({ setToken, switchToRegister }) {
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError("");
		setIsLoading(true);

		const result = await userAPI.login(email, password);

		if (!result.ok) {
			setError(getAPIErrorMessage(result, "login"));
			setIsLoading(false);
		}
		if (!result.data?.token) {
			setError("Something went wrong. Please try again.");
			setIsLoading(false);
			return;
		}

		setToken(result.data.token);
		localStorage.setItem("token", result.data.token);
		setIsLoading(false);
	};

	return (
		<div className="flex items-center justify-center min-h-screen bg-gradient-to-r from-main-50 to-main-100">
			<div className="px-10 py-8 mt-4 bg-white shadow-2xl rounded-2xl w-full max-w-md border border-gray-200">
				<h3 className="text-3xl font-extrabold text-center text-gray-800 mb-6">
					Login to your account
				</h3>

				{error && (
					<div className="mt-4 p-3 bg-red-100 text-red-700 rounded-md border border-red-200 shadow-sm">
						{error}
					</div>
				)}

				<form onSubmit={handleSubmit} className="space-y-5">
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

					<div className="flex flex-col sm:flex-row items-center justify-between mt-6 gap-3">
						<button
							className="w-full sm:w-auto px-6 py-3 text-white bg-gradient-to-r from-main-600 to-main-700 rounded-lg shadow-lg hover:from-main-700 hover:to-main-800 disabled:opacity-50 transition"
							disabled={isLoading}
						>
							{isLoading ? "Logging in..." : "Login"}
						</button>
						<button
							type="button"
							className="text-main-600 hover:underline disabled:opacity-50 mt-2 sm:mt-0"
							onClick={switchToRegister}
							disabled={isLoading}
						>
							Register now
						</button>
					</div>
				</form>
			</div>
		</div>
	);
}
