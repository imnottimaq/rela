import React, { useState } from "react";
import {
	DndContext,
	DragOverlay,
	closestCorners,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from "@dnd-kit/core";
import { arrayMove, sortableKeyboardCoordinates } from "@dnd-kit/sortable";

import Nav from "../components/Nav";
import Container from "../components/Board";
import { TaskItem } from "../components/TaskItem";

export default function Boards({ token, handleLogout }) {
	const [items, setItems] = useState({
		root: ["1", "2", "3"],
		container1: ["4", "5", "6"],
		container2: ["7", "8", "9"],
		container3: [],
	});
	const [activeId, setActiveId] = useState();

	const sensors = useSensors(
		useSensor(PointerSensor),
		useSensor(KeyboardSensor, {
			coordinateGetter: sortableKeyboardCoordinates,
		})
	);

    const containerClasses = "flex-1 bg-gray-50 rounded-lg p-4 shadow-md min-w-[250px] max-w-[400px] max-h-full overflow-y-auto"

	return (
		<div className="flex flex-col h-screen bg-gray-100">
			<Nav handleLogout={handleLogout} />

			<div className="flex flex-1 overflow-x-auto p-4 gap-4">
				<DndContext
					sensors={sensors}
					collisionDetection={closestCorners}
					onDragStart={handleDragStart}
					onDragOver={handleDragOver}
					onDragEnd={handleDragEnd}
				>
					<Container
						id="root"
						items={items.root}
						activeId={activeId}
						className={containerClasses}
					/>
					<Container
						id="container1"
						items={items.container1}
						activeId={activeId}
						className={containerClasses}
					/>
					<Container
						id="container2"
						items={items.container2}
						activeId={activeId}
						className={containerClasses}
					/>
					<Container
						id="container3"
						items={items.container3}
						activeId={activeId}
						className={containerClasses}
					/>

					<DragOverlay>
						{activeId ? (
							<TaskItem
								id={activeId}
								className="bg-main-200 border-main-400 shadow-lg scale-110 transition-transform duration-200 ease-in-out cursor-grabbing"
							/>
						) : null}
					</DragOverlay>
				</DndContext>
			</div>
		</div>
	);

	function findContainer(id) {
		if (id in items) return id;

		return Object.keys(items).find((key) => items[key].includes(id));
	}

	function handleDragStart(event) {
		const { active } = event;
		const { id } = active;

		setActiveId(id);
	}

	function handleDragOver(event) {
		const { active, over, draggingRect } = event;
		const id = active.id;
		const overId = over?.id;

		if (!over) return;

		const activeContainer = findContainer(id);
		const overContainer = findContainer(overId);

		if (
			!activeContainer ||
			!overContainer ||
			activeContainer === overContainer
		)
			return;

		setItems((prev) => {
			const activeItems = prev[activeContainer];
			const overItems = prev[overContainer];

			const activeIndex = activeItems.indexOf(id);
			const overIndex = overItems.indexOf(overId);

			let newIndex = overIndex >= 0 ? overIndex : overItems.length;

			if (draggingRect && over.rect) {
				const isBelowLastItem =
					overIndex === overItems.length - 1 &&
					draggingRect.offsetTop >
						over.rect.offsetTop + over.rect.height;

				if (isBelowLastItem) newIndex += 1;
			}

			return {
				...prev,
				[activeContainer]: activeItems.filter((item) => item !== id),
				[overContainer]: [
					...overItems.slice(0, newIndex),
					activeItems[activeIndex],
					...overItems.slice(newIndex),
				],
			};
		});
	}

	function handleDragEnd(event) {
		const { active, over } = event;
		const { id } = active;
		const { id: overId } = over;

		const activeContainer = findContainer(id);
		const overContainer = findContainer(overId);

		if (
			!activeContainer ||
			!overContainer ||
			activeContainer !== overContainer
		) {
			return;
		}

		const activeIndex = items[activeContainer].indexOf(active.id);
		const overIndex = items[overContainer].indexOf(overId);

		if (activeIndex !== overIndex) {
			setItems((items) => ({
				...items,
				[overContainer]: arrayMove(
					items[overContainer],
					activeIndex,
					overIndex
				),
			}));
		}

		setActiveId(null);
	}
}
