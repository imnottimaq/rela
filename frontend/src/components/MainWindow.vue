<template>
              <!-- :buttons="[
        { label: 'Minimize', onClick: test },
        { label: 'Maximize', onClick: test },
        { label: 'Close', onClick: test }
      ]" -->
    <WindowComponent
      class="main-window"
      title="Welcome to Rela"
      :menu="windowMenu"
      :initial-position="{ x: 160, y: 120 }"
      :initial-size="{ width: 300, height: 460 }"
      :min-size="{ width: 300, height: 460 }"
      v-model:visible="mainVisible"
      storage-key="rela-window-main"
    >
      <p>Rela is WIP task tracker with ability to self-host it.<br />Use Account to login or register.</p>
      <img class="demo" src="../assets/confused_travolta.gif" alt="Confused Travolta" />
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { showLoginWindow } from '../composables/useLoginWindow';
import { showRegisterWindow } from '../composables/useRegisterWindow';
import { computed, ref } from 'vue';
import { useAuth } from '../composables/useAuth';

const mainVisible = ref(true);
const { isAuthenticated, logout } = useAuth();

const openProfile = () => {
  console.log("Profile placeholder window");
};

const openWorkspaceCreation = () => {
  console.log("Workspace creation placeholder window");
};

const unauthenticatedMenu = [
  {
    label: "Account",
    items: [
      { type: "button", label: "Login", onClick: showLoginWindow },
      { type: "button", label: "Register", onClick: showRegisterWindow },
    ],
  },
];

const authenticatedMenu = [
  {
    label: "Account",
    items: [
      { type: "button", label: "Profile", onClick: openProfile },
      { type: "separator" },
      { type: "button", label: "Logout", onClick: logout },
    ],
  },
  {
    label: "Workspace",
    items: [,
      { type: "button", label: "Create workspace", onClick: openWorkspaceCreation },
    ],
  },
];

const windowMenu = computed(() =>
  isAuthenticated.value ? authenticatedMenu : unauthenticatedMenu
);


</script>
<style scoped>
.demo {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 auto;
}
</style>
