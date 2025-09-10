const MainContent = () => {
	return (
		<div className="flex items-center justify-center h-full">
			<div className="text-center text-gray-500">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					className="h-16 w-16 mx-auto mb-4"
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
				<p>There is nothing here :(</p>
			</div>
		</div>
	);
};

export default MainContent;
