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
      :initial-size="{ width: 640, height: 460 }"
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
import { ref } from 'vue';

const mainVisible = ref(true);

let windowMenu = [
  {
    label: "Account",
    items: [
      { type: "button", label: "Login", onClick: showLoginWindow },
      { type: "button", label: "Register", onClick: showRegisterWindow }
    ],
  },
];


</script>
<style scoped>
.demo {
  max-width: 100%;
  max-height: 70%;
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 auto;
}
</style>