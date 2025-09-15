import React from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

export function TaskItem({ id, className, children }) {
	return (
		<div
			className={`w-full h-12 flex items-center justify-between border border-gray-400 rounded-md my-2 px-2 bg-white shadow-sm ${
				className || ""
			}`}
		>
			<span>{id}</span>
			{children}
		</div>
	);
}

export default function SortableItem({ id }) {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id });

	const style = {
		transform: transform ? CSS.Transform.toString(transform) : undefined,
		transition,
		opacity: isDragging ? 0.5 : 1,
	};

	return (
		<div ref={setNodeRef} style={style}>
			<TaskItem id={id} className={isDragging ? "bg-main-100" : ""}>
				<button
					{...attributes}
					{...listeners}
					className="cursor-grab active:cursor-grabbing opacity-70 select-none"
				>
					⠿
				</button>
			</TaskItem>
		</div>
	);
}
