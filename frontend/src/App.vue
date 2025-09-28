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

    <InviteWindow
      v-model:visible="isInviteWindowVisible"
      :join-token="inviteToken"
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
import { onMounted, ref } from 'vue';
import MainWindow from './components/MainWindow.vue';
import LoginWindow from './components/LoginWindow.vue';
import RegisterWindow from './components/RegisterWindow.vue';
import ProfileWindow from './components/ProfileWindow.vue';
import ConfirmLogoutWindow from './components/ConfirmLogoutWindow.vue';
import CreateWorkspaceWindow from './components/CreateWorkspaceWindow.vue';
import WorkspaceWindow from './components/WorkspaceWindow.vue';
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

const mainStyle = `background-image: url(${background})`;

const { openWorkspaceWindows, closeWorkspaceWindow, fetchWorkspaces } = useWorkspaces();
const { openBoardWindows, closeBoardWindow } = useBoards();
const { isAuthenticated } = useAuth();
const { showLoginWindow } = useLoginWindow();

const isInviteWindowVisible = ref(false);
const inviteToken = ref(null);

onMounted(() => {
  if (typeof window === 'undefined') return;

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
});

const handleJoinedWorkspace = () => {
  fetchWorkspaces();
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
