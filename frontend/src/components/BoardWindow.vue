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
import { computed, onMounted, ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { boardsApi } from '../utils/http';

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

// Local copy that we can enrich with fetched data
const localBoard = ref({ ...(props.board || {}) });

const boardId = computed(() => localBoard.value?._id || localBoard.value?.id || localBoard.value?.Id || localBoard.value?.name);
const boardName = computed(() => localBoard.value?.name || String(boardId.value));
const workspaceName = computed(() => props.workspaceName || String(props.workspaceId));

const close = () => {
  emit('update:visible', false);
  emit('close');
};

const fetchBoard = async () => {
  const id = boardId.value;
  // Fetch only if we have an id and no real name yet
  if (!id) return;
  if (localBoard.value?.name && localBoard.value.name !== String(id)) return;
  try {
    const { data } = await boardsApi.getBoard(id);
    if (data && (data._id || data.id)) {
      localBoard.value = { _id: data._id || data.id, name: data.name || String(id) };
    }
  } catch (e) {
    // keep placeholder on failure
    // console.error('Failed to fetch board info', e);
  }
};

onMounted(fetchBoard);
watch(() => boardId.value, () => fetchBoard());
</script>

<style scoped>
.content { text-align: left; padding: 0 12px 12px; }
.hint { color: #555; font-style: italic; }
</style>
