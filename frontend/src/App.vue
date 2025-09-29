<template>
  <main :style="mainStyle">
    <MainWindow />
    <YouWindow />

    <LoginWindow />
    <RegisterWindow />
    <ForgotPasswordWindow/>
    <ProfileWindow />
    <ConfirmLogoutWindow />
    <CreateWorkspaceWindow />
    <AboutRelaWindow />
    <EditWorkspaceWindow />
    <EditTaskWindow />


    <CreateBoardWindow
        v-if="createBoardWorkspaceId"
        :workspace-id="createBoardWorkspaceId"
        @created="handleBoardCreated"
    />
    <InviteWindow
        v-if="joinToken"
        :model-value="true"
        :join-token="joinToken"
        @joined="handleJoinedWorkspace"
    />

    <WorkspaceWindow
      v-for="win in openWorkspaceWindows"
      :key="win.id"
      :workspace="{ _id: win.id, name: win.name }"
      v-model:visible="win.visible"
      @close="closeWorkspaceWindow(win.id)"
    />

    <BoardWindow
      v-for="bwin in openBoardWindows"
      :key="bwin.id"
      :workspace-id="bwin.workspaceId"
      :workspace-name="bwin.workspaceName"
      :board="{ _id: bwin.id.split(':')[1], name: bwin.name }"
      v-model:visible="bwin.visible"
      @close="closeBoardWindow(bwin.id)"
    />
  </main>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue';
import MainWindow from './components/MainWindow.vue';
import LoginWindow from './components/LoginWindow.vue';
import RegisterWindow from './components/RegisterWindow.vue';
import ProfileWindow from './components/ProfileWindow.vue';
import ConfirmLogoutWindow from './components/ConfirmLogoutWindow.vue';
import CreateWorkspaceWindow from './components/CreateWorkspaceWindow.vue';
import BoardWindow from './components/BoardWindow.vue';
import YouWindow from "./components/YouWindow.vue";
import InviteWindow from "./components/InviteWindow.vue";
import { useWorkspaces } from './composables/useWorkspaces';
import { useBoards, restoreBoardWindowsFromStorage } from './composables/useBoards';
import { useAuth } from './composables/useAuth';
import background from './assets/windows7.jpg';
import ForgotPasswordWindow from "./components/ForgotPasswordWindow.vue";
import AboutRelaWindow from "./components/AboutRelaWindow.vue";
import { useLoginWindow } from "./composables/useLoginWindow.js";
import WorkspaceWindow from "./components/WorkspaceWindow.vue";
import CreateBoardWindow from "./components/CreateBoardWindow.vue";
import EditWorkspaceWindow from "./components/EditWorkspaceWindow.vue";
import EditTaskWindow from "./components/EditTaskWindow.vue";
import { useCreateBoardWindow } from "./composables/useCreateBoardWindow.js";

const mainStyle = `background-image: url(${background})`;

const joinToken = ref(null);

const { openWorkspaceWindows, closeWorkspaceWindow, fetchWorkspaces, updateWorkspaceInList } = useWorkspaces();
const { openBoardWindows, closeBoardWindow } = useBoards();
const { isAuthenticated } = useAuth();
const { showLoginWindow } = useLoginWindow();
const { createBoardWorkspaceId } = useCreateBoardWindow();

const isInviteWindowVisible = ref(false);
const inviteToken = ref(null);

const handleWorkspaceUpdatedEvent = (event) => {
  const workspace = event.detail.workspace;
  const id = workspace.id || workspace._id;
  updateWorkspaceInList(id, workspace);
};

onMounted(() => {
  if (typeof window === 'undefined') return;

  window.addEventListener('rela:workspace-updated', handleWorkspaceUpdatedEvent);

  const path = window.location.pathname;
  const match = path.match(/^\/invite\/([^/]+)/);
  let isHandlingInvite = false;

  if (match) {
    isHandlingInvite = true;
    inviteToken.value = match[1];
    isInviteWindowVisible.value = true;
    window.history.replaceState({}, document.title, '/');
  }

  if (!isAuthenticated?.value && !isHandlingInvite) {
    showLoginWindow();
  }

  if (isAuthenticated?.value) {
    try { restoreBoardWindowsFromStorage(); } catch (_) {}
  }
  const storedToken = localStorage.getItem('rela_join_token');
  if (storedToken) {
    joinToken.value = storedToken;
    isInviteWindowVisible.value = true;
  }
});

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('rela:workspace-updated', handleWorkspaceUpdatedEvent);
  }
});

const handleJoinedWorkspace = () => {
  fetchWorkspaces();
};

const handleBoardCreated = (board) => {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('rela:board-created', { detail: { board } }));
  }
};

</script>

<style scoped>
main {
  min-height: 100vh;
  display: grid;
  place-items: center;
  text-align: center;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}
</style>
