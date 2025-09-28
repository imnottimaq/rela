<template>
  <WindowComponent
    :title="`Create Board`"
    :buttons="[{ label: 'Close', onClick: onCancel }]"
    v-model:visible="modelVisible"
    :storage-key="`rela-window-create-board-${workspaceId}`"
    footer-buttons-align="right"
    :footer-buttons="[
      { label: 'Cancel', onClick: onCancel },
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
import { computed, ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { workspaceApi } from '../utils/http';

const props = defineProps({
  workspaceId: { type: [String, Number], required: true },
  visible: { type: Boolean, default: false },
});

const emit = defineEmits(['update:visible', 'created']);

const name = ref('');
const error = ref('');
const isSubmitting = ref(false);

const reset = () => {
  name.value = '';
  error.value = '';
};

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

watch(() => props.visible, (v) => {
  if (!v) {
    reset();
  }
});

const onCancel = () => {
  emit('update:visible', false);
};

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
    emit('update:visible', false);
    reset();
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
