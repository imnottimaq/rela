<template>
          <!-- :initial-size="{ width: 450, height: 520 }"
      :min-size="{ width: 300, height: 460 }" -->
    <WindowComponent
      class="main-window"
      title="Welcome to Rela"
      :menu="windowMenu"
      :initial-position="{ x: 160, y: 120 }"
      :initial-size="{ width: 450, height: 250 }"
      :min-size="{ width: 300, height: 200 }"
      v-model:visible="mainVisible"
      storage-key="rela-window-main"
    >
      <p>Rela is WIP task tracker with ability to self-host it.<br />You can move and resize all windows.<br />Use Account to login or register.</p>
      <!-- <img class="demo" src="../assets/confused_travolta.gif" alt="Confused Travolta" /> -->
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { showLoginWindow } from '../composables/useLoginWindow';
import { showRegisterWindow } from '../composables/useRegisterWindow';
import { showProfileWindow } from '../composables/useProfileWindow';
import { computed, ref, onMounted, watch } from 'vue';
import { useAuth } from '../composables/useAuth';
import { showConfirmLogoutWindow } from '../composables/useConfirmLogoutWindow';
import { useWorkspaces, openWorkspaceWindow, refreshWorkspaces } from '../composables/useWorkspaces';
import { showCreateWorkspaceWindow } from '../composables/useCreateWorkspaceWindow';
import { showAboutRelaWindow } from '../composables/useAboutRelaWindow.js';

const mainVisible = ref(true);
const { isAuthenticated } = useAuth();

const openProfile = () => {
  showProfileWindow();
};

const { workspaces } = useWorkspaces();

const openWorkspaceCreation = () => {
  showCreateWorkspaceWindow();
};

const unauthenticatedMenu = [
  {
    label: "Account",
    items: [
      { type: "button", label: "Login", onClick: showLoginWindow },
      { type: "button", label: "Register", onClick: showRegisterWindow },
    ],
  },
  {
    label: "Help",
    items: [
      {type: "button",label:"About Rela",onClick: showAboutRelaWindow },
    ]
  },
];

const authenticatedMenu = computed(() => {
  const ws = workspaces.value || [];
  const items = [
    {
      type: "button",
      label: "Create workspace",
      onClick: openWorkspaceCreation,
      divider: ws.length > 0,
    },
  ];
  if (ws.length > 0) {
    items.push({ type: "separator" });
  }
  items.push(
    ...ws.map((w) => ({
      type: "button",
      label: w.name || String(w._id || w.id),
      onClick: () => openWorkspaceWindow(w),
    }))
  );

  return [
    {
      label: "Account",
      items: [
        { type: "button", label: "Profile", onClick: openProfile },
        { type: "separator" },
        { type: "button", label: "Logout", onClick: () => showConfirmLogoutWindow() },
      ],
    },
    {
      label: "Workspace",
      items,
    },
    {
      label: "Help",
      items: [
        {type: "button",label:"About Rela",onClick: showAboutRelaWindow },
      ]
    },
  ];
});

const windowMenu = computed(() =>
  isAuthenticated.value ? authenticatedMenu.value : unauthenticatedMenu
);

onMounted(() => {
  if (isAuthenticated.value) {
    refreshWorkspaces();
  }
});

watch(() => isAuthenticated.value, (v) => {
  if (v) refreshWorkspaces();
});


</script>
<style scoped>
.demo {
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 auto;
}
</style>
