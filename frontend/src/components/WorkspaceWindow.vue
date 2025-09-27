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
      <div v-if="loading" class="hint">Working...</div>
      <div v-if="error" class="error">{{ error }}</div>

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

    <EditWorkspaceWindow
      :workspace="workspace"
      v-model:visible="editWorkspaceVisible"
      @save="handleEditWorkspace"
    />

    <DeleteWorkspaceWindow
      :workspace="workspace"
      v-model:visible="deleteWorkspaceVisible"
      @confirm="handleDeleteWorkspace"
    />
  </WindowComponent>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import CreateBoardWindow from './CreateBoardWindow.vue';
import EditWorkspaceWindow from './EditWorkspaceWindow.vue';
import DeleteWorkspaceWindow from './DeleteWorkspaceWindow.vue';
import { workspaceApi } from '../utils/http';
import { openBoardWindow } from '../composables/useBoards';
import { createWorkspaceInviteLink } from '../composables/useWorkspaces';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: true },
});

const emit = defineEmits(['update:visible', 'close', 'workspace-updated', 'workspace-deleted']);

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
const editWorkspaceVisible = ref(false);
const deleteWorkspaceVisible = ref(false);
const selectedBoardId = ref(null);

const loading = ref(false);
const error = ref('');

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
  await loadBoards();
  if (!board) return;

  const boardId = board._id || board.id;
  let boardToOpen = null;

  if (boardId) {
    boardToOpen = boards.value.find(b => (b._id || b.id) === boardId);
  } else if (board.name) {
    boardToOpen = boards.value.find(b => b.name === board.name);
  }

  if (boardToOpen) {
    openBoardWindow(props.workspace, boardToOpen);
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

const handleEditWorkspace = async (newName) => {
  loading.value = true;
  error.value = '';
  try {
    await workspaceApi.updateWorkspace(wsId.value, { name: newName });
    emit('workspace-updated', { id: wsId.value, name: newName });
    editWorkspaceVisible.value = false;
  } catch (e) {
    console.error('Failed to update workspace', e);
    error.value = 'Failed to update workspace name.';
    window.alert(error.value); // Or use a more sophisticated notification
  } finally {
    loading.value = false;
  }
};

const handleDeleteWorkspace = async () => {
  loading.value = true;
  error.value = '';
  try {
    await workspaceApi.deleteWorkspace(wsId.value);
    deleteWorkspaceVisible.value = false;
    emit('workspace-deleted', wsId.value);
    close();
  } catch (e) {
    console.error('Failed to delete workspace', e);
    error.value = 'Failed to delete workspace.';
    window.alert(error.value); // Or use a more sophisticated notification
  } finally {
    loading.value = false;
  }
};

const windowMenu = computed(() => {
  const list = Array.isArray(boards.value) ? boards.value : [];
  const boardItems = [
    {
      type: 'button',
      label: 'Create board',
      onClick: () => (createBoardVisible.value = true),
    },
  ];

  if (list.length > 0) {
    boardItems.push({ type: 'separator' });
    boardItems.push(
      ...list.map((b) => ({
        type: 'button',
        label: b.name || String(b._id || b.id),
        onClick: () => selectBoard(b),
      }))
    );
  }

  const workspaceItems = [
    {
      label: 'Edit Name',
      onClick: () => (editWorkspaceVisible.value = true),
      type: 'button',
    },
    { type: 'separator' },
    {
      label: 'Delete Workspace',
      onClick: () => (deleteWorkspaceVisible.value = true),
      type: 'button',
    },
    {
      label: 'Create and Copy Invite Link',
      onClick: () => createWorkspaceInviteLink(wsId.value),
      type: 'button',
    },
  ];

  return [
    { label: 'Workspace', items: workspaceItems },
    { label: 'Boards', items: boardItems },
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
