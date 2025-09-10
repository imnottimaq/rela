import MainContent from "../components/MainContent";

function Dashboard({ token, handleLogout }) {
	return (
		<div className="flex flex-col h-screen">
			<header className="bg-white shadow">
				<div className="flex justify-between items-center px-6 py-4">
					<h1 className="text-xl font-bold text-gray-800">
						Task Tracker
					</h1>
					<button
						onClick={handleLogout}
						className="px-4 py-2 text-sm text-white bg-red-600 rounded-md hover:bg-red-700"
					>
						Logout
					</button>
				</div>
			</header>

			<div className="flex flex-1 overflow-hidden">
				<main className="flex-1 overflow-auto p-6 bg-gray-100">
					<MainContent />
				</main>
			</div>
		</div>
	);
}

export default Dashboard;
