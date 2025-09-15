import MainContent from "../components/MainContent";
import Nav from "../components/Nav";

export default function Dashboard({ token, handleLogout }) {
	return (
		<div className="flex flex-col h-screen">
			<Nav handleLogout={handleLogout} />

			<div className="flex flex-1 overflow-hidden bg-gray-100">
				<main className="flex-1 overflow-auto p-8">
					<MainContent />
				</main>
			</div>
		</div>
	);
}
