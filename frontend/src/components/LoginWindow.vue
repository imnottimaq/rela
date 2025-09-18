<template>
    <WindowComponent
      title="Login"
      :buttons="[{ label: 'Close', onClick: hideLoginWindow }]"
      v-model:visible="loginVisible"
      storage-key="rela-window-login"
    >
      <p>Login placeholder</p>
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { useLoginWindow } from '../composables/useLoginWindow';

const { loginVisible, hideLoginWindow } = useLoginWindow();
</script>
