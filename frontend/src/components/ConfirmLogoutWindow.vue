<template>
  <WindowComponent
    title="Confirm Logout"
    :buttons="[{ label: 'Close', onClick: handleCancel }]"
    v-model:visible="logoutConfirmVisible"
    storage-key="rela-window-confirm-logout"
    :initial-size="{ width: 340, height: 180 }"
    :min-size="{ width: 320, height: 160 }"
    footer-buttons-align="right"
    :footer-buttons="footerButtons"
  >
    <div class="content">
      <p>Are you sure you want to logout?</p>
    </div>
  </WindowComponent>
  
</template>

<script setup>
import { computed } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { useAuth } from '../composables/useAuth';
import { useConfirmLogoutWindow } from '../composables/useConfirmLogoutWindow';

const { logout } = useAuth();
const { logoutConfirmVisible, hideConfirmLogoutWindow, __onConfirmedRef } = useConfirmLogoutWindow();

const handleCancel = () => {
  hideConfirmLogoutWindow();
};

const handleConfirm = () => {
  try {
    logout();
  } finally {
    const cb = __onConfirmedRef?.value;
    hideConfirmLogoutWindow();
    if (typeof cb === 'function') {
      try { cb(); } catch (e) { /* swallow callback errors */ }
    }
  }
};

const footerButtons = computed(() => [
  { label: 'Cancel', onClick: handleCancel },
  { label: 'Logout', onClick: handleConfirm, primary: true },
]);
</script>

<style scoped>
.content {
  text-align: left;
  padding: 0 12px 12px;
}
</style>

