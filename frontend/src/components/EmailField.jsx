export default function EmailField({
	label,
	setEmail,
	email,
	isLoading,
	inputClassName,
}) {

	return (
		<div className="w-full">
			<div>
				{label && (
					<label
						className="block text-gray-700 font-medium"
						htmlFor="email"
					>
						{label}
					</label>
				)}
				<input
					type="email"
					placeholder="you@example.com"
					className={`w-full mt-2 px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-main-500 focus:border-transparent transition ${
						inputClassName || ""
					}`}
					value={email}
					onChange={(e) => setEmail(e.target.value)}
					required
					disabled={isLoading}
				/>
			</div>
		</div>
	);
}
