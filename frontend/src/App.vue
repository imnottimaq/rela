<template>
  <main :style="mainStyle">
    <MainWindow />

    <LoginWindow />
    <RegisterWindow />
    <ForgotPasswordWindow/>
    <ProfileWindow />
    <ConfirmLogoutWindow />
    <CreateWorkspaceWindow />

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
import MainWindow from './components/MainWindow.vue';
import LoginWindow from './components/LoginWindow.vue';
import RegisterWindow from './components/RegisterWindow.vue';
import ProfileWindow from './components/ProfileWindow.vue';
import ConfirmLogoutWindow from './components/ConfirmLogoutWindow.vue';
import CreateWorkspaceWindow from './components/CreateWorkspaceWindow.vue';
import WorkspaceWindow from './components/WorkspaceWindow.vue';
import BoardWindow from './components/BoardWindow.vue';
import { useWorkspaces } from './composables/useWorkspaces';
import { useBoards, restoreBoardWindowsFromStorage } from './composables/useBoards';
import { useAuth } from './composables/useAuth';
import background from './assets/windows7.jpg';
import ForgotPasswordWindow from "./components/ForgotPasswordWindow.vue";

const mainStyle = `background-image: url(${background})`;

const { openWorkspaceWindows, closeWorkspaceWindow } = useWorkspaces();
const { openBoardWindows, closeBoardWindow } = useBoards();
const { isAuthenticated } = useAuth();

// Attempt to restore board windows from storage only when authenticated
if (typeof window !== 'undefined' && isAuthenticated?.value) {
  try { restoreBoardWindowsFromStorage(); } catch (_) {}
}
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
