import { FaRegClipboard } from "react-icons/fa";

export default function MainContent() {
	return (
		<div className="flex flex-col items-center justify-center h-full text-center text-gray-400">
			<FaRegClipboard className="h-20 w-20 mb-4 text-gray-300" />
			<h2 className="text-xl font-semibold mb-2">Nothing here yet</h2>
			<p>Start by creating a new board or checking back later!</p>
		</div>
	);
}
