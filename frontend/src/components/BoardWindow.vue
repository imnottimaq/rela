<template>
  <WindowComponent
    :title="`Board: ${boardName}`"
    :buttons="[{ label: 'Close', onClick: close }]"
    v-model:visible="modelVisible"
    :initial-size="{ width: 420, height: 320 }"
    :min-size="{ width: 360, height: 260 }"
    :storage-key="`rela-window-board-${workspaceId}-${boardId}`"
  >
    <div class="content">
      <p>Workspace: <strong>{{ workspaceName }}</strong></p>
      <p>Board ID: <code>{{ boardId }}</code></p>
      <p>Board Name: <strong>{{ boardName }}</strong></p>
      <p class="hint">This is a placeholder window for the board.</p>
    </div>
  </WindowComponent>
  
</template>

<script setup>
import { computed } from 'vue';
import WindowComponent from './WindowComponent.vue';

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

const boardId = computed(() => props.board?._id || props.board?.id || props.board?.Id || props.board?.name);
const boardName = computed(() => props.board?.name || String(boardId.value));
const workspaceName = computed(() => props.workspaceName || String(props.workspaceId));

const close = () => {
  emit('update:visible', false);
  emit('close');
};
</script>

<style scoped>
.content { text-align: left; padding: 0 12px 12px; }
.hint { color: #555; font-style: italic; }
</style>

