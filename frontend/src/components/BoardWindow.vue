<template>
  <WindowComponent
    :title="`Board: ${boardName}`"
    :buttons="[{ label: 'Close', onClick: close }]"
    v-model:visible="modelVisible"
    :initial-size="{ width: 420, height: 480 }"
    :min-size="{ width: 360, height: 320 }"
    :storage-key="`rela-window-board-${workspaceId}-${boardId}`"
    :menu="windowMenu"
  >
    <div class="content" @dragover="onDragOver" @drop="onDrop">
      <p>
        Board Name:
        <strong v-if="!isEditingBoard" @dblclick="editBoard" title="Double-click to edit">{{ boardName }}</strong>
        <span v-else class="board-name-edit">
          <input
              ref="boardNameInput"
              v-model="newBoardName"
              placeholder="Enter new board name"
              @blur="saveBoard"
              @keydown.enter.prevent="saveBoard"
              @keydown.esc.prevent="cancelEditBoard"
          />
        </span>
      </p>
      <div v-if="isLoading" class="loading">Loading tasks...</div>
      <div v-else-if="error && !isNotFound" class="error">Failed to load tasks.</div>
      <div v-else-if="tasks.length" class="table-wrap">
        <table class="task-table has-shadow">
          <thead>
            <tr>
              <th>Tasks</th>
              <th class="task-actions-header"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="task in tasks"
              :key="task._id || task.id || task.Id"
              class="task-row"
              draggable="true"
              @dragstart="onDragStart(task, $event)"
              @contextmenu.prevent="onTaskRightClick(task, $event)"
            >
              <td>
                <div class="task-title">{{ task.name || task.title || 'Unnamed Task' }}</div>
                <div v-if="task.description" class="task-desc">{{ task.description }}</div>
                <div v-if="task.deadline" class="task-deadline">Deadline: {{ formatDeadline(task.deadline) }}</div>
              </td>
              <td class="task-actions"></td>
            </tr>
          </tbody>
        </table>
        <teleport to="body">
          <ul
              v-if="contextMenu.visible"
              @click.stop
              role="menu"
              class="context-menu window"
              style="position: fixed; width: 180px;"
              :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px', zIndex: contextMenu.zIndex }"
          >
            <li role="menuitem" @click="editTask(contextMenu.selectedTask)"><a>Edit</a></li>
            <li role="menuitem" @click="deleteTask(contextMenu.selectedTask)"><a>Delete</a></li>
          </ul>
        </teleport>
      </div>
      <div v-else class="no-tasks">No tasks found for this board.</div>
      <div class="status-bar" style="position:absolute;bottom:0;">
        <p class="status-bar-field"><p>Workspace: <strong>{{ workspaceName }}</strong></p></p>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed, ref, reactive, onMounted, onUnmounted, nextTick } from 'vue';
import WindowComponent from './common/WindowComponent.vue';
import { useBoardTasks } from '../composables/useBoards';
import { workspaceApi } from '../utils/http';

const props = defineProps({
  workspaceId: { type: [String, Number], required: true },
  workspaceName: { type: String, default: '' },
  board: { type: Object, required: true },
  visible: { type: Boolean, default: true },
});

const emit = defineEmits(['update:visible', 'close']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const localBoard = ref({ ...(props.board || {}) });

const isEditingBoard = ref(false);
const newBoardName = ref('');
const boardNameInput = ref(null);

const boardId = computed(() => localBoard.value?._id || localBoard.value?.id || localBoard.value?.Id || localBoard.value?.name);
const boardName = computed(() => localBoard.value?.name || String(boardId.value));

const workspaceIdRef = computed(() => props.workspaceId);
const workspaceName = computed(() => props.workspaceName || String(workspaceIdRef.value));

const { tasks, isLoading, error, isNotFound, fetchTasks } = useBoardTasks(workspaceIdRef, boardId);

const contextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  selectedTask: null,
  zIndex: 10000,
});

const onTaskRightClick = (task, event) => {
  contextMenu.visible = true;
  contextMenu.x = event.clientX;
  contextMenu.y = event.clientY;
  contextMenu.selectedTask = task;
  if (typeof window !== 'undefined' && window.__relaWindowLayerManager__) {
    contextMenu.zIndex = window.__relaWindowLayerManager__.maxLayer + 1;
  }
};

const hideContextMenu = () => {
  contextMenu.visible = false;
};

const openCreateTaskWindow = () => {
  if (typeof window !== 'undefined') {
    const detail = { workspaceId: workspaceIdRef.value, boardId: boardId.value };
    window.dispatchEvent(new CustomEvent('rela:open-create-task-window', { detail }));
  }
};

const editTask = (task) => {
  if (!task) return;
  if (typeof window !== 'undefined') {
    const detail = { workspaceId: workspaceIdRef.value, task };
    window.dispatchEvent(new CustomEvent('rela:open-edit-task-window', { detail }));
  }
  hideContextMenu();
};

const deleteTask = (task) => {
  if (!task) return;
  if (confirm(`Are you sure you want to delete task: ${task.name || 'Unnamed Task'}?`)) {
    alert(`Deleting task: ${task.name || 'Unnamed Task'}`);
    // Here you would call an API to delete the task
    // And then remove it from the local 'tasks' array
  }
  hideContextMenu();
};

const close = () => {
  emit('update:visible', false);
  emit('close');
};

const getTaskId = (task) => task?._id || task?.id || task?.Id;

const formatDeadline = (timestamp) => {
  if (!timestamp) return '';
  const date = new Date(timestamp * 1000);
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

const onDragStart = (task, ev) => {
  try {
    const payload = {
      taskId: getTaskId(task),
      fromBoardId: boardId.value,
      workspaceId: workspaceIdRef.value,
    };
    ev.dataTransfer?.setData('application/x-rela-task', JSON.stringify(payload));
    ev.dataTransfer?.setData('text/plain', JSON.stringify(payload));
    if (ev.dataTransfer) ev.dataTransfer.effectAllowed = 'move';
  } catch (_) {}
};

const onDragOver = (ev) => {
  try {
    const types = Array.from(ev.dataTransfer?.types || []);
    if (types.includes('application/x-rela-task') || types.includes('text/plain')) {
      ev.preventDefault();
      if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'move';
    }
  } catch (_) {}
};

const onDrop = async (ev) => {
  let raw = '';
  try {
    raw = ev.dataTransfer?.getData('application/x-rela-task') || ev.dataTransfer?.getData('text/plain') || '';
  } catch (_) {}
  if (!raw) return;
  let data;
  try { data = JSON.parse(raw); } catch (_) { return; }
  const { taskId, fromBoardId, workspaceId } = data || {};
  const targetBoardId = boardId.value;
  if (!taskId || !workspaceId || !targetBoardId) return;
  if (String(workspaceId) !== String(workspaceIdRef.value)) return;
  if (String(fromBoardId) === String(targetBoardId)) return;
  try {
    await workspaceApi.updateTask(workspaceId, taskId, { board: targetBoardId });
    if (typeof window !== 'undefined') {
      try {
        window.dispatchEvent(new CustomEvent('rela:task-moved', { detail: { taskId, fromBoardId, toBoardId: targetBoardId, workspaceId } }));
      } catch (_) {}
    }
    fetchTasks();
  } catch (err) {
    console.error('Failed to move task to another board', err);
  }
};

const onTaskMoved = (e) => {
  try {
    const d = e?.detail || {};
    if (String(d.workspaceId) !== String(workspaceIdRef.value)) return;
    if (String(d.fromBoardId) === String(boardId.value)) {
      tasks.value = (tasks.value || []).filter((t) => String(getTaskId(t)) !== String(d.taskId));
    }
    if (String(d.toBoardId) === String(boardId.value)) {
      fetchTasks();
    }
  } catch (_) {}
};

const onTaskCreated = (e) => {
  const detail = e.detail || {};
  if (String(detail.workspaceId) === String(workspaceIdRef.value) && String(detail.boardId) === String(boardId.value)) {
    fetchTasks();
  }
};

const onTaskUpdated = (e) => {
  const detail = e.detail || {};
  if (String(detail.workspaceId) === String(workspaceIdRef.value) && String(detail.boardId) === String(boardId.value)) {
    fetchTasks();
  }
};

const editBoard = async () => {
  isEditingBoard.value = true;
  newBoardName.value = boardName.value;
  await nextTick();
  boardNameInput.value?.focus();
};

const cancelEditBoard = () => {
  isEditingBoard.value = false;
};

const saveBoard = async () => {
  if (!isEditingBoard.value) return;
  isEditingBoard.value = false;
  const oldName = boardName.value;
  const newName = newBoardName.value.trim();
  if (newName && newName !== oldName) {
    try {
      await workspaceApi.updateBoard(workspaceIdRef.value, boardId.value, { name: newName });
      localBoard.value.name = newName;
      if (typeof window !== 'undefined') {
        try {
          window.dispatchEvent(new CustomEvent('rela:board-updated', { detail: { boardId: boardId.value, workspaceId: workspaceIdRef.value, board: localBoard.value } }));
        } catch (_) {}
      }
    } catch (err) {
      console.error('Failed to update board name', err);
      alert('Failed to update board name.');
    }
  }
};

const deleteBoard = async () => {
  if (confirm(`Are you sure you want to delete board: ${boardName.value}?`)) {
    try {
      await workspaceApi.deleteBoard(workspaceIdRef.value, boardId.value);
      if (typeof window !== 'undefined') {
        try {
          window.dispatchEvent(new CustomEvent('rela:board-deleted', { detail: { boardId: boardId.value, workspaceId: workspaceIdRef.value } }));
        } catch (_) {}
      }
      close();
    } catch (err) {
      console.error('Failed to delete board', err);
      alert('Failed to delete board.');
    }
  }
};


onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('rela:task-moved', onTaskMoved);
    window.addEventListener('rela:task-created', onTaskCreated);
    window.addEventListener('rela:task-updated', onTaskUpdated);
    document.addEventListener('click', hideContextMenu);
  }
});

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('rela:task-moved', onTaskMoved);
    window.removeEventListener('rela:task-created', onTaskCreated);
    window.removeEventListener('rela:task-updated', onTaskUpdated);
    document.removeEventListener('click', hideContextMenu);
  }
});

const windowMenu = computed(() => {
  const taskItems = (tasks.value || []).map(task => ({
    type: 'button',
    label: task.name || task.title || 'Unnamed Task',
    onClick: () => editTask(task),
  }));

  return [
    {
      label: 'Board',
      items: [
        { type: 'button', label: 'Edit Name', onClick: editBoard },
        { type: 'button', label: 'Delete Board', onClick: deleteBoard, divider: true },
        { type: 'button', label: 'New Task', onClick: openCreateTaskWindow },
        { type: 'button', label: 'Refresh', onClick: fetchTasks },
      ],
    },
  ];
});

</script>

<style scoped>
.content {
  text-align: left;
  padding: 0 12px 12px;
  overflow-y: auto;
  height: 100%;
}
.content h4 {
  margin-top: 1em;
  margin-bottom: 0.5em;
}
.loading,
.error,
.no-tasks {
  padding: 1em;
  text-align: center;
  color: #555;
}
.error {
  color: #d9534f;
}
.task-table { width: 100%; border-collapse: collapse; }
.task-title { font-weight: 600; color: #222; }
.task-desc { font-size: 12px; color: #666; margin-top: 2px; }
.task-deadline { font-size: 12px; color: #c0392b; margin-top: 2px; font-weight: 500; }
.task-actions-header { width: 40px; }
.task-actions { text-align: left; vertical-align: middle; width: 50px; }
.context-menu{
  border-radius: 0;
}
p .board-name-edit {
  display: inline-block;
}
.board-name-edit input {
  font-size: 1em;
  font-weight: 600;
  border: 1px solid #888;
  padding: 1px 4px;
  border-radius: 3px;
  width: 200px;
  background: #fefefe;
  color: #222;
}

</style>
