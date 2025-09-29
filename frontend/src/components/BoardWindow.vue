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
      <p>Workspace: <strong>{{ workspaceName }}</strong></p>
      <p>Board Name: <strong>{{ boardName }}</strong></p>
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
    </div>

    <EditTaskWindow
        :workspace-id="workspaceId"
        :board-id="boardId"
        v-model:visible="taskWindowVisible"
        :task="taskToEdit"
        @created="handleTaskCreated"
        @updated="handleTaskUpdated"
    />
  </WindowComponent>
</template>

<script setup>
import { computed, ref, reactive, onMounted, onUnmounted } from 'vue';
import WindowComponent from './WindowComponent.vue';
import EditTaskWindow from './EditTaskWindow.vue';
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
const taskWindowVisible = ref(false);
const taskToEdit = ref(null);

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
  taskToEdit.value = null;
  taskWindowVisible.value = true;
};

const editTask = (task) => {
  if (!task) return;
  taskToEdit.value = task;
  taskWindowVisible.value = true;
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

const handleTaskCreated = (newTask) => {
  if (newTask) {
    tasks.value.unshift(newTask);
  }
  if (isNotFound.value) {
    isNotFound.value = false;
  }
};

const handleTaskUpdated = () => {
  fetchTasks();
};

const getTaskId = (task) => task?._id || task?.id || task?.Id;

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


onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('rela:task-moved', onTaskMoved);
    document.addEventListener('click', hideContextMenu);
  }
});

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('rela:task-moved', onTaskMoved);
    document.removeEventListener('click', hideContextMenu);
  }
});

const windowMenu = computed(() => {
  const taskItems = (tasks.value || []).map(task => ({
    type: 'button',
    label: task.name || task.title || 'Unnamed Task',
    onClick: () => alert(`Selected task: ${task.name || task.title || 'Unnamed Task'}`),
  }));

  return [
    {
      label: 'Tasks',
      items: [
        {
          type: 'button',
          label: 'New Task',
          onClick: openCreateTaskWindow,
        },
        {
          type: 'button',
          label: 'Refresh Tasks',
          onClick: fetchTasks,
          divider: tasks.value.length > 0,
        },
        ...(tasks.value.length > 0 ? taskItems : []),
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
.task-actions-header { width: 40px; }
.task-actions { text-align: left; vertical-align: middle; width: 50px; }
.context-menu{
  border-radius: 0;
}

</style>
