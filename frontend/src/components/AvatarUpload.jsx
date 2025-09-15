export default function EmailField({
    setAvatar,
	avatarPreview,
	setAvatarPreview,
	isLoading,
}) {
	return (
		<div className="w-full">
			<div className="flex flex-col items-center">
				<label className="block text-gray-700 font-medium mb-2">
					Avatar (optional)
				</label>
				<div className="relative w-24 h-24 rounded-full overflow-hidden border-2 border-gray-300 flex items-center justify-center cursor-pointer hover:border-main-500 transition">
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
		</div>
	);
}
