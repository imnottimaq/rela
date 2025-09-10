const MainContent = () => {
	return (
		<div className="flex flex-col items-center justify-center h-full text-center text-gray-400">
			<div className="bg-white p-12 rounded-2xl shadow-xl flex flex-col items-center">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					className="h-20 w-20 mb-4 text-gray-300"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						strokeWidth={2}
						d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
					/>
				</svg>
				<h2 className="text-xl font-semibold mb-2">Nothing here yet</h2>
				<p className="text-gray-500">
					Start by creating a new task or checking back later!
				</p>
			</div>
		</div>
	);
};

export default MainContent;
