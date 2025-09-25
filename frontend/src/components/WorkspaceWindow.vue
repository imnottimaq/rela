<template>
  <WindowComponent
    :title="`Workspace: ${workspace?.name || 'Unnamed'}`"
    :buttons="[{ label: 'Close', onClick: close }]"
    v-model:visible="modelVisible"
    :initial-size="{ width: 520, height: 420 }"
    :min-size="{ width: 420, height: 320 }"
    :storage-key="`rela-window-workspace-${workspace?.id || workspace?._id || workspace?.name}`"
    :menu="windowMenu"
  >
    <div class="content">
      <section class="ws-info">
        <p>ID: <code>{{ wsId }}</code></p>
        <p>Name: <strong>{{ workspace?.name }}</strong></p>
      </section>

      <section class="boards">
        <div class="boards-header">
          <h3>Boards</h3>
        </div>
        <div v-if="boardsError" class="error">{{ boardsError }}</div>
        <div v-if="loadingBoards" class="hint">Loading boards…</div>
        <ul v-else class="board-list">
          <li v-if="boards.length === 0" class="hint">No boards yet</li>
          <li v-for="b in boards" :key="b._id || b.id">
            <button class="inline-link" type="button" @click="selectBoard(b)">
              <span class="board-name">{{ b.name || (b._id || b.id) }}</span>
            </button>
          </li>
        </ul>
      </section>
    </div>

    <CreateBoardWindow
      :workspace-id="wsId"
      v-model:visible="createBoardVisible"
      @created="handleBoardCreated"
    />
  </WindowComponent>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import CreateBoardWindow from './CreateBoardWindow.vue';
import { workspaceApi } from '../utils/http';
import { openBoardWindow } from '../composables/useBoards';
import { createWorkspaceInviteLink } from '../composables/useWorkspaces';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: true },
});

const emit = defineEmits(['update:visible', 'close']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const close = () => {
  emit('update:visible', false);
  emit('close');
};

const wsId = computed(() => props.workspace?._id || props.workspace?.id || props.workspace?.Id || props.workspace?.name);

const boards = ref([]);
const loadingBoards = ref(false);
const boardsError = ref('');
const createBoardVisible = ref(false);
const selectedBoardId = ref(null);

const loadBoards = async () => {
  if (!wsId.value) return;
  loadingBoards.value = true;
  boardsError.value = '';
  try {
    const { data } = await workspaceApi.getBoards(wsId.value);
    boards.value = Array.isArray(data) ? data : (Array.isArray(data?.boards) ? data.boards : []);
  } catch (e) {
    console.error('Failed to load boards', e);
    boards.value = [];
    boardsError.value = 'Failed to load boards';
  } finally {
    loadingBoards.value = false;
  }
};

const handleBoardCreated = async (board) => {
  // Optimistically add to list; fallback to reloading
  if (board && (board._id || board.id || board.name)) {
    // If no id returned, refresh to resolve its id
    if (!board._id && !board.id) {
      await loadBoards();
      const byName = (boards.value || []).find((b) => (b.name || '').trim() === (board.name || '').trim());
      const toOpen = byName || board;
      if (byName && !boards.value.some((b) => (b._id || b.id) === (byName._id || byName.id))) {
        boards.value = [byName, ...boards.value];
      }
      openBoardWindow(props.workspace, toOpen);
    } else {
      boards.value = [board, ...boards.value];
      openBoardWindow(props.workspace, board);
    }
  } else {
    loadBoards();
  }
};

onMounted(() => {
  loadBoards();
});

watch(() => wsId.value, () => loadBoards());

const selectBoard = (board) => {
  const id = board?._id || board?.id || board?.name;
  selectedBoardId.value = id ?? null;
  openBoardWindow(props.workspace, board);
};

const windowMenu = computed(() => {
  const list = Array.isArray(boards.value) ? boards.value : [];
  const items = [
    {
      type: 'button',
      label: 'Create board',
      onClick: () => (createBoardVisible.value = true),
      divider: list.length > 0,
    },
  ];
  if (list.length > 0) {
    items.push({ type: 'separator' });
    items.push(
      ...list.map((b) => ({
        type: 'button',
        label: b.name || String(b._id || b.id),
        onClick: () => selectBoard(b),
      }))
    );
  }
  return [
    { label: 'Boards', items },
    { label: 'Get Invite Link', createWorkspaceInviteLink},
  ];
});
</script>

<style scoped>
.content { text-align: left; padding: 0 12px 12px; }
.ws-info { margin-bottom: 10px; }
.boards { margin-top: 10px; }
.boards-header { display: flex; align-items: center; justify-content: space-between; }
.board-list { list-style: none; padding: 0; margin: 8px 0 0; }
.board-name { font-weight: 500; }
.error { color: #c00; }
.hint { color: #555; font-style: italic; }
.btn { font-size: 12px; padding: 4px 8px; }
.inline-link {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  color: #0066cc;
  cursor: pointer;
  text-decoration: underline;
}
.inline-link:hover { color: #004b99; }
</style>
