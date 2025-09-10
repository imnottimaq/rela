import { useState } from "react";
import { userAPI } from "../api";
import { getAPIErrorMessage } from "../api/errors";
import {
	validateEmail,
	validatePassword,
	validateName,
} from "../utils/validation";

const Register = ({ setToken, switchToLogin }) => {
	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [isLoading, setIsLoading] = useState(false);

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

        setIsLoading(false);
    };

	return (
		<div className="flex items-center justify-center min-h-screen">
			<div className="px-8 py-6 mt-4 text-left bg-white shadow-lg rounded-md w-full max-w-md">
				<h3 className="text-2xl font-bold text-center">
					Create a new account
				</h3>
				{error && (
					<div className="mt-4 p-2 bg-red-100 text-red-700 rounded">
						{error}
					</div>
				)}
				<form onSubmit={handleSubmit}>
					<div className="mt-4">
						<div>
							<label className="block" htmlFor="name">
								Name
							</label>
							<input
								type="text"
								placeholder="Name"
								className="w-full px-4 py-2 mt-2 border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-600"
								value={name}
								onChange={(e) => setName(e.target.value)}
								required
								disabled={isLoading}
							/>
						</div>
						<div className="mt-4">
							<label className="block" htmlFor="email">
								Email
							</label>
							<input
								type="email"
								placeholder="Email"
								className="w-full px-4 py-2 mt-2 border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-600"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								required
								disabled={isLoading}
							/>
						</div>
						<div className="mt-4">
							<label className="block">Password</label>
							<input
								type="password"
								placeholder="Password"
								className="w-full px-4 py-2 mt-2 border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-600"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								required
								disabled={isLoading}
								minLength={8}
							/>
							<p className="text-xs text-gray-500 mt-1">
								Password must be at least 8 characters long
							</p>
						</div>
						<div className="flex items-baseline justify-between mt-6">
							<button
								className="px-6 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-900 disabled:opacity-50"
								disabled={isLoading}
							>
								{isLoading ? "Creating account..." : "Register"}
							</button>
							<button
								type="button"
								className="text-blue-600 hover:underline disabled:opacity-50"
								onClick={switchToLogin}
								disabled={isLoading}
							>
								Already have an account?
							</button>
						</div>
					</div>
				</form>
			</div>
		</div>
	);
};

export default Register;
