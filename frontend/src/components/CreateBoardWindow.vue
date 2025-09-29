<template>
  <WindowComponent
    :title="`Create Board`"
    :buttons="[{ label: 'Close', onClick: hideCreateBoardWindow }]"
    v-model:visible="createBoardVisible"
    :storage-key="`rela-window-create-board-${workspaceId}`"
    footer-buttons-align="right"
    :footer-buttons="[
      { label: 'Cancel', onClick: hideCreateBoardWindow },
      { label: 'Create', onClick: onCreate, primary: true, loading: isSubmitting, disabled: isSubmitting }
    ]"
    :initial-size="{ width: 320, height: 200 }"
    :min-size="{ width: 320, height: 200 }"
  >
    <div style="text-align: left; padding: 0 10px;">
      <h1>New board</h1>
      <div class="group" style="width: 100%">
        <label for="board-name">Name</label>
        <input id="board-name" type="text" v-model="name" @keydown.enter.prevent="onCreate" aria-describedby="board-name-error" />
        <div v-if="error" class="error" id="board-name-error" role="tooltip">{{ error }}</div>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref, watch } from 'vue';
import WindowComponent from './common/WindowComponent.vue';
import { workspaceApi } from '../utils/http';
import { useCreateBoardWindow } from '../composables/useCreateBoardWindow.js';

const { createBoardVisible, hideCreateBoardWindow } = useCreateBoardWindow();

const props = defineProps({
  workspaceId: { type: [String, Number], required: true },
});

const emit = defineEmits(['created']);

const name = ref('');
const error = ref('');
const isSubmitting = ref(false);

const reset = () => {
  name.value = '';
  error.value = '';
};

watch(createBoardVisible, (isVisible) => {
  if (!isVisible) {
    reset();
  }
});

const onCreate = async () => {
  error.value = '';
  const trimmed = name.value.trim();
  if (!trimmed) {
    error.value = 'Please enter board name';
    return;
  }
  try {
    if (isSubmitting.value) return;
    isSubmitting.value = true;
    const { data } = await workspaceApi.createBoard(props.workspaceId, { name: trimmed });
    // Always emit something usable; fallback to just the name
    emit('created', data && typeof data === 'object' ? data : { name: trimmed });
    hideCreateBoardWindow();
  } catch (e) {
    console.error('Create board failed', e);
    error.value = 'Failed to create board';
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<style scoped>
.error { color: #d00000; }
</style>
