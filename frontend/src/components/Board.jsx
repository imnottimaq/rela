import { useDroppable } from "@dnd-kit/core";
import {
    SortableContext,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import SortableItem from "./TaskItem";

export default function Container({ id, items, activeId, className }) {

	const { setNodeRef } = useDroppable({
		id,
	});

	return (
		<SortableContext
			id={id}
			items={items}
			strategy={verticalListSortingStrategy}
		>
			<div ref={setNodeRef} className={className || "flex-1 m-3 p-3 bg-main-200 rounded-lg"}>
				{items.map((id) => (
					<SortableItem
						key={id}
						id={id}
						isDragging={activeId === id}
					/>
				))}
			</div>
		</SortableContext>
	);
}