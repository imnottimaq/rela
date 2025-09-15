import { useState, useEffect } from "react";
import {
	BrowserRouter as Router,
	Routes,
	Route,
	Navigate,
} from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Boards from "./pages/Boards";

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
				{/* Public routes */}
				<Routes>
					<Route
						path="/login"
						element={<Login setToken={setToken} />}
					/>
					<Route
						path="/register"
						element={<Register setToken={setToken} />}
					/>

					{/* Protected route */}
					<Route
						path="/boards"
						element={
							token ? (
								<Boards
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
								to={token ? "/boards" : "/login"}
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
