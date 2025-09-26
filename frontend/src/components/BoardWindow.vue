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
    <div class="content">
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
