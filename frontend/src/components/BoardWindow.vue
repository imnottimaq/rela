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
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="task in tasks"
              :key="task._id || task.id || task.Id"
              class="task-row"
              draggable="true"
              @dragstart="onDragStart(task, $event)"
            >
              <td>
                <div class="task-title">{{ task.name || task.title || 'Unnamed Task' }}</div>
                <div v-if="task.description" class="task-desc">{{ task.description }}</div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="no-tasks">No tasks found for this board.</div>
    </div>

    <CreateTaskWindow
      :workspace-id="workspaceId"
      :board-id="boardId"
      v-model:visible="createTaskVisible"
      @created="handleTaskCreated"
    />
  </WindowComponent>
</template>

<script setup>
import { computed, ref } from 'vue';
import WindowComponent from './WindowComponent.vue';
import CreateTaskWindow from './CreateTaskWindow.vue';
import { useBoardTasks } from '../composables/useBoards';
import { workspaceApi } from '../utils/http';
import { onMounted, onUnmounted } from 'vue';

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
const createTaskVisible = ref(false);

const boardId = computed(() => localBoard.value?._id || localBoard.value?.id || localBoard.value?.Id || localBoard.value?.name);
const boardName = computed(() => localBoard.value?.name || String(boardId.value));

const workspaceIdRef = computed(() => props.workspaceId);
const workspaceName = computed(() => props.workspaceName || String(workspaceIdRef.value));

const { tasks, isLoading, error, isNotFound, fetchTasks } = useBoardTasks(workspaceIdRef, boardId);

const close = () => {
  emit('update:visible', false);
  emit('close');
};

const handleTaskCreated = (newTask) => {
  if (newTask) {
    tasks.value.unshift(newTask);
  }
  // In case this was the first task
  if (isNotFound.value) {
    isNotFound.value = false;
  }
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
    // Fallback for some UAs
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
    await workspaceApi.updateTask(taskId, { board: targetBoardId });
    // Notify all BoardWindow instances to update their local task lists
    if (typeof window !== 'undefined') {
      try {
        window.dispatchEvent(new CustomEvent('rela:task-moved', { detail: { taskId, fromBoardId, toBoardId: targetBoardId, workspaceId } }));
      } catch (_) {}
    }
    // Refresh our list as target window
    fetchTasks();
  } catch (err) {
    console.error('Failed to move task to another board', err);
  }
};

const onTaskMoved = (e) => {
  try {
    const d = e?.detail || {};
    if (String(d.workspaceId) !== String(workspaceIdRef.value)) return;
    // If this window is the source board, remove the task optimistically
    if (String(d.fromBoardId) === String(boardId.value)) {
      tasks.value = (tasks.value || []).filter((t) => String(getTaskId(t)) !== String(d.taskId));
    }
    // If this window is the destination board, ensure list is in sync
    if (String(d.toBoardId) === String(boardId.value)) {
      fetchTasks();
    }
  } catch (_) {}
};

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener('rela:task-moved', onTaskMoved);
  }
});

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('rela:task-moved', onTaskMoved);
  }
});

const windowMenu = computed(() => {
  const taskItems = (tasks.value || []).map(task => ({
    type: 'button',
    label: task.name || task.title || 'Unnamed Task',
    // placeholder onClick
    onClick: () => alert(`Selected task: ${task.name || task.title || 'Unnamed Task'}`),
  }));

  return [
    {
      label: 'Tasks',
      items: [
        {
          type: 'button',
          label: 'New Task',
          onClick: () => { createTaskVisible.value = true; },
        },
        {
          type: 'button',
          label: 'Refresh Tasks',
          onClick: fetchTasks,
          divider: tasks.value.length > 0,
        },
        ...(tasks.value.length > 0 ? taskItems : [{ label: 'No tasks yet', disabled: true }]),
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
/*
.table-wrap { overflow: auto; }

.task-table thead th {
  text-align: left;
  background: #f1f1f1;
  border-bottom: 1px solid #ddd;
  padding: 8px;
  font-weight: 600;
}
.task-table tbody td { padding: 10px 8px; border-bottom: 1px solid #eee; }
.task-row:hover td { background: #fafafa; }
*/
.task-table { width: 100%}
.task-title { font-weight: 600; color: #222; }
.task-desc { font-size: 12px; color: #666; margin-top: 2px; }

</style>
