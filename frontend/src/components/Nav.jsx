import { NavLink } from "react-router-dom";
import Logo from "../assets/logo.svg";
import { RiLogoutBoxRLine } from "react-icons/ri";

export default function Nav({ handleLogout }) {
	function navLinkClass({ isActive }) {
		return [
			"underline-offset-4",
			"decoration-2",
			"transition-colors",
			isActive
				? "underline text-main-600"
				: "text-gray-800 hover:underline hover:decoration-gray-400",
		].join(" ");
	}

	return (
		<nav className="bg-white shadow-md">
			<div className="flex justify-between items-center px-6 py-4">
				<div className="flex justify-between items-center gap-8">
					<NavLink
						to="/"
						className="flex items-center space-x-2 cursor-pointer"
					>
						<img
							src={Logo}
							alt=""
							className="h-8 w-8 object-contain"
						/>
						<h1 className="text-2xl font-bold text-gray-800">
							Rela
						</h1>
					</NavLink>

					<div className="flex items-center gap-4 text-lg text-gray-800">
						<NavLink
							to="/boards"
							className={navLinkClass}
						>
							Your tasks
						</NavLink>

						<NavLink
							to="/"
							className={navLinkClass}
						>
							Team tasks
						</NavLink>
					</div>
				</div>

				<div className="flex items-center space-x-4">
					<div className="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-500">
						U
					</div>

					<button
						onClick={handleLogout}
						className="px-4 py-2 text-sm text-white bg-red-600 rounded-lg shadow hover:bg-red-700 transition cursor-pointer"
					>
						Logout <RiLogoutBoxRLine className="inline ml-1" />
					</button>
				</div>
			</div>
		</nav>
	);
}
