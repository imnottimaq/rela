import { useState } from "react";
import { FaRegEye, FaRegEyeSlash } from "react-icons/fa";

export default function PasswordField({
	label,
	setPassword,
	password,
	isLoading,
	inputClassName,
}) {
	const [showPassword, setShowPassword] = useState(false);

	const toggleShowPassword = (e) => {
		e.preventDefault();
		setShowPassword((prev) => !prev);
	};

	return (
		<div className="w-full">
			{label && (
				<label className="block text-gray-700 font-medium mb-1">
					{label}
				</label>
			)}

			<div
				className={`flex items-center justify-between w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus-within:ring-2 focus-within:ring-main-500 focus-within:border-transparent transition ${
					inputClassName || ""
				}`}
			>
				<input
					type={showPassword ? "text" : "password"}
					placeholder="••••••••"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					required
					disabled={isLoading}
					minLength={8}
					className="flex-1 outline-none bg-transparent"
				/>
				<button
					type="button"
					onClick={toggleShowPassword}
					className="ml-2 text-gray-600 hover:text-gray-900"
				>
					{showPassword ? <FaRegEye /> : <FaRegEyeSlash />}
				</button>
			</div>
		</div>
	);
}
