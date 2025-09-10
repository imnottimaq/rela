import { useState, useEffect } from "react";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";

function App() {
	const [token, setToken] = useState(localStorage.getItem("token") || "");
	const [currentView, setCurrentView] = useState(
		token ? "dashboard" : "login"
	);

	useEffect(() => {
		setCurrentView(token ? "dashboard" : "login");
	}, [token]);

	const handleLogout = () => {
		setToken("");
        localStorage.removeItem("token");
		setCurrentView("login");
	};

	return (
		<div className="min-h-screen bg-gray-100">
			{currentView === "login" && (
				<Login
					setToken={setToken}
					switchToRegister={() => setCurrentView("register")}
				/>
			)}
			{currentView === "register" && (
				<Register
					setToken={setToken}
					switchToLogin={() => setCurrentView("login")}
				/>
			)}
			{currentView === "dashboard" && (
				<Dashboard token={token} handleLogout={handleLogout} />
			)}
		</div>
	);
}

export default App;
