<script setup lang="ts">

import WindowComponent from "./common/WindowComponent.vue";
import { useForgotPasswordWindow } from '../composables/useForgotPasswordWindow';

const { forgotPasswordVisible, hideForgotPasswordWindow } = useForgotPasswordWindow();
</script>

<template>
  <WindowComponent
      title="Forgot password"
      :buttons="[{ label: 'Close', onClick: hideForgotPasswordWindow }]"
      v-model:visible="forgotPasswordVisible"
      storage-key="rela-window-forgotpassword"
      footer-buttons-align="right"
      :footer-buttons="[]"
      :initialSize="{ width: 300, height: 370 }"
      :minSize="{ width: 300, height: 370 }"
  >
    <div style="text-align: left; padding: 0 10px;">
      <h1 style="text-align: center;">In Development</h1>
    </div>

  </WindowComponent>
</template>

<style scoped>

</style>