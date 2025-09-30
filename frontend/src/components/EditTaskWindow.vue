<template>
  <WindowComponent
      :title="windowTitle"
      v-model:visible="visible"
      :initial-size="{ width: 380, height: 320 }"
      :min-size="{ width: 320, height: 280 }"
      :footer-buttons="footerButtons"
      footer-buttons-align="right"
  >
    <div class="form-content">
      <div class="field-row-stacked">
        <div>
          <a>Name: </a>
          <input id="task-title-input" type="text" v-model="taskName" @keydown.enter="submit"  placeholder="Name" aria-describedby="task-title-error"/>
          <div v-if="error" class="error-message" id="task-title-error" role="tooltip">{{ error }}</div>
          <br>
        </div>
        <a>Description: </a>
        <input id="task-desc-input" type="text" v-model="taskDesc" @keydown.enter="submit"  placeholder="Description"/>
        <br>
        <a>Deadline: </a>
        <input id="task-deadline-input" type="date" v-model="taskDeadline" placeholder="Deadline" />
        <br>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue';
import WindowComponent from './common/WindowComponent.vue';
import { authApi, workspaceApi } from '../utils/http';

const visible = ref(false);
const isEditing = ref(false);
const taskName = ref('');
const taskDesc = ref('');
const taskDeadline = ref('');
const assignedTo = ref(null);
const workspaceMembers = ref([]);
const currentUser = ref(null);
const isLoading = ref(false);
const error = ref('');

const workspaceId = ref(null);
const boardId = ref(null);
const task = ref(null);

const windowTitle = computed(() => isEditing.value ? 'Edit Task' : 'Create New Task');

const getTaskId = (task) => task?._id || task?.id || task?.Id;

const fetchWorkspaceMembers = async (wId) => {
  try {
    const { data } = await workspaceApi.getWorkspaceInfo(wId);
    const owner = data.owner;
    const members = data.members || [];
    if (owner && !members.some(m => m.id === owner.id)) {
      members.unshift(owner);
    }
    workspaceMembers.value = members;
  } catch (err) {
    console.error('Failed to fetch workspace members:', err);
  }
};
const fetchCurrentUser = async () => {
  try {
    const { data } = await authApi.getUserInfo();
    currentUser.value = data;
  } catch (err) {
    console.error('Failed to fetch current user:', err);
  }
};

const resetState = () => {
  taskName.value = '';
  taskDesc.value = '';
  taskDeadline.value = '';
  assignedTo.value = null;
  error.value = '';
  isLoading.value = false;
  workspaceId.value = null;
  boardId.value = null;
  task.value = null;
  isEditing.value = false;
};

const handleOpen = (wId) => {
  fetchCurrentUser();
  fetchWorkspaceMembers(wId);
  visible.value = true;
};

const openForCreate = (e) => {
  const detail = e.detail || {};
  if (!detail.workspaceId || !detail.boardId) return;

  resetState();
  workspaceId.value = detail.workspaceId;
  boardId.value = detail.boardId;
  handleOpen(detail.workspaceId);
};

const openForEdit = (e) => {
  const detail = e.detail || {};
  if (!detail.workspaceId || !detail.task) return;

  resetState();
  isEditing.value = true;
  workspaceId.value = detail.workspaceId;
  task.value = detail.task;
  
  taskName.value = detail.task.name;
  taskDesc.value = detail.task.description || '';
  taskDeadline.value = detail.task.deadline ? new Date(detail.task.deadline * 1000).toISOString().split('T')[0] : '';
  assignedTo.value = detail.task.assignedTo || null;
  boardId.value = detail.task.board;

  handleOpen(detail.workspaceId);
};

onMounted(() => {
  window.addEventListener('rela:open-create-task-window', openForCreate);
  window.addEventListener('rela:open-edit-task-window', openForEdit);
});

onUnmounted(() => {
  window.removeEventListener('rela:open-create-task-window', openForCreate);
  window.removeEventListener('rela:open-edit-task-window', openForEdit);
});

const close = () => {
  visible.value = false;
};

const submit = async () => {
  if (!taskName.value.trim()) {
    error.value = 'Task title is required.';
    return;
  }
  if (isLoading.value) return;

  isLoading.value = true;
  error.value = '';

  try {
    let deadline = null;
    if (taskDeadline.value) {
      const [year, month, day] = taskDeadline.value.split('-').map(Number);
      // Date.UTC uses 0-indexed months, so subtract 1 from the month.
      deadline = Math.floor(Date.UTC(year, month - 1, day) / 1000);
    }

    const payload = {
      name: taskName.value,
      description: taskDesc.value,
      deadline: deadline,
      assignedTo: assignedTo.value,
    };

    let eventName;
    let eventDetail;

    if (isEditing.value) {
      const taskId = getTaskId(task.value);
      if (!taskId) {
        error.value = 'Invalid task ID';
        isLoading.value = false;
        return;
      }
      payload.board = task.value.board;
      const { data } = await workspaceApi.updateTask(workspaceId.value, taskId, payload);
      eventName = 'rela:task-updated';
      eventDetail = { task: data, workspaceId: workspaceId.value, boardId: data.board };
    } else {
      payload.board = boardId.value;
      const { data } = await workspaceApi.createTask(workspaceId.value, payload);
      eventName = 'rela:task-created';
      eventDetail = { task: data, workspaceId: workspaceId.value, boardId: boardId.value };
    }
    
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent(eventName, { detail: eventDetail }));
    }

    close();
  } catch (err) {
    console.error(`Failed to ${isEditing.value ? 'update' : 'create'} task:`, err);
    error.value = err.response?.data?.error || 'An unknown error occurred.';
  } finally {
    isLoading.value = false;
  }
};

const footerButtons = computed(() => [
  {
    label: 'Cancel',
    onClick: close,
    disabled: isLoading.value,
  },
  {
    label: isEditing.value ? 'Save' : 'Create',
    onClick: submit,
    disabled: !taskName.value.trim(),
    loading: isLoading.value,
  },
]);
</script>

<style scoped>

</style>
