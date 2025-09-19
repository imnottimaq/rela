<template>
  <main :style="mainStyle">
    <MainWindow />

    <LoginWindow />
    <RegisterWindow />
    <ProfileWindow />
    <CreateWorkspaceWindow />

    <WorkspaceWindow
      v-for="win in openWorkspaceWindows"
      :key="win.id"
      :workspace="{ _id: win.id, name: win.name }"
      v-model:visible="win.visible"
      @close="closeWorkspaceWindow(win.id)"
    />
  </main>
</template>

<script setup>
import MainWindow from './components/MainWindow.vue';
import LoginWindow from './components/LoginWindow.vue';
import RegisterWindow from './components/RegisterWindow.vue';
import ProfileWindow from './components/ProfileWindow.vue';
import CreateWorkspaceWindow from './components/CreateWorkspaceWindow.vue';
import WorkspaceWindow from './components/WorkspaceWindow.vue';
import { useWorkspaces } from './composables/useWorkspaces';
import background from './assets/windows7.jpg';

const mainStyle = `background-image: url(${background})`;

const { openWorkspaceWindows, closeWorkspaceWindow } = useWorkspaces();
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
