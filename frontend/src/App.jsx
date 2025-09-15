import { useState, useEffect } from "react";
import {
	BrowserRouter as Router,
	Routes,
	Route,
	Navigate,
} from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";

function App() {
	const [token, setToken] = useState(localStorage.getItem("token") || "");

	useEffect(() => {
		if (token) {
			localStorage.setItem("token", token);
		}
	}, [token]);

	const handleLogout = () => {
		setToken("");
		localStorage.removeItem("token");
	};

	return (
		<Router>
			<div className="min-h-screen bg-gray-100">
				<Routes>
					{/* Public routes */}
					<Route
						path="/login"
						element={
							token ? (
								<Navigate to="/dashboard" replace />
							) : (
								<Login setToken={setToken} />
							)
						}
					/>
					<Route
						path="/register"
						element={
							token ? (
								<Navigate to="/dashboard" replace />
							) : (
								<Register setToken={setToken} />
							)
						}
					/>

					{/* Protected route */}
					<Route
						path="/dashboard"
						element={
							token ? (
								<Dashboard
									token={token}
									handleLogout={handleLogout}
								/>
							) : (
								<Navigate to="/login" replace />
							)
						}
					/>

					{/* Default redirect */}
					<Route
						path="*"
						element={
							<Navigate
								to={token ? "/dashboard" : "/login"}
								replace
							/>
						}
					/>
				</Routes>
			</div>
		</Router>
	);
}

export default App;
