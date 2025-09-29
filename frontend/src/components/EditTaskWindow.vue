<template>
  <WindowComponent
      :title="windowTitle"
      v-model:visible="modelVisible"
      :initial-size="{ width: 380, height: 240 }"
      :min-size="{ width: 320, height: 220 }"
      :footer-buttons="footerButtons"
      footer-buttons-align="right"
  >
    <div class="form-content">
      <div class="field-row-stacked">
        <div>
          <input id="task-title-input" type="text" v-model="taskName" @keydown.enter="submit"  placeholder="Name" aria-describedby="task-title-error"/>
          <div v-if="error" class="error-message" id="task-title-error" role="tooltip">{{ error }}</div>
        </div>
        <input id="task-desc-input" type="text" v-model="taskDesc" @keydown.enter="submit"  placeholder="Description"/>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { workspaceApi } from '../utils/http';

const props = defineProps({
  workspaceId: { type: [String, Number], required: true },
  boardId: { type: [String, Number], required: true },
  visible: { type: Boolean, default: false },
  task: { type: Object, default: null },
});

const emit = defineEmits(['update:visible', 'created', 'updated']);

const taskName = ref('');
const taskDesc = ref('');
const isLoading = ref(false);
const error = ref('');

const isEditing = computed(() => !!props.task);
const windowTitle = computed(() => isEditing.value ? 'Edit Task' : 'Create New Task');

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const getTaskId = (task) => task?._id || task?.id || task?.Id;

watch(() => props.visible, (isVisible) => {
  if (isVisible) {
    if (isEditing.value) {
      taskName.value = props.task.name;
      taskDesc.value = props.task.description || '';
    } else {
      taskName.value = '';
      taskDesc.value = '';
    }
    error.value = '';
    isLoading.value = false;
  }
});

const close = () => {
  emit('update:visible', false);
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
    const payload = {
      name: taskName.value,
      description: taskDesc.value,
    };

    if (isEditing.value) {
      const taskId = getTaskId(props.task);
      if (!taskId) {
        error.value = 'Invalid task ID';
        isLoading.value = false;
        return;
      }
      payload.board = props.task.board;
      const { data } = await workspaceApi.updateTask(props.workspaceId, taskId, payload);
      emit('updated', data);
    } else {
      payload.board = props.boardId;
      const { data } = await workspaceApi.createTask(props.workspaceId, payload);
      emit('created', data);
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
