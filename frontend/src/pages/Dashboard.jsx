// Dashboard.jsx
import MainContent from "../components/MainContent";

function Dashboard({ token, handleLogout }) {
	return (
		<div className="flex flex-col h-screen">
			{/* Navbar */}
			<nav className="bg-white shadow-md">
				<div className="flex justify-between items-center px-6 py-4">
					{/* Logo / Title */}
					<h1 className="text-2xl font-bold text-gray-800">
						Task Tracker
					</h1>

					{/* User section */}
					<div className="flex items-center space-x-4">
						{/* Placeholder for avatar */}
						<div className="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-500">
							U
						</div>

						{/* Logout button */}
						<button
							onClick={handleLogout}
							className="px-4 py-2 text-sm text-white bg-red-600 rounded-lg shadow hover:bg-red-700 transition"
						>
							Logout
						</button>
					</div>
				</div>
			</nav>

			{/* Main Content */}
			<div className="flex flex-1 overflow-hidden bg-gray-100">
				<main className="flex-1 overflow-auto p-8">
					<MainContent />
				</main>
			</div>
		</div>
	);
}

export default Dashboard;
