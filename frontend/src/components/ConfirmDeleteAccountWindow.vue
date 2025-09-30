<template>
  <WindowComponent
    title="Confirm Account Deletion"
    :buttons="[{ label: 'Close', onClick: handleCancel }]"
    v-model:visible="deleteAccountConfirmVisible"
    storage-key="rela-window-confirm-delete-account"
    :initial-size="{ width: 360, height: 240 }"
    :min-size="{ width: 340, height: 220 }"
    footer-buttons-align="right"
    :footer-buttons="footerButtons"
  >
    <div class="content">
      <p class="warning-text">Are you sure you want to delete your account? This action cannot be undone.</p>
      <div class="field">
        <label for="password-input" class="label">Enter your password to confirm:</label>
        <input
          id="password-input"
          type="password"
          v-model="password"
          class="password-input"
          @keyup.enter="handleConfirm"
          placeholder="Password"
        />
      </div>
      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </WindowComponent>

</template>

<script setup>
import { ref, computed, watch } from 'vue';
import WindowComponent from './common/WindowComponent.vue';
import { useAuth } from '../composables/useAuth';
import { useConfirmDeleteAccountWindow } from '../composables/useConfirmDeleteAccountWindow';
import { authApi, clearAuthTokens } from '../utils/http';

const { logout } = useAuth();
const { deleteAccountConfirmVisible, hideConfirmDeleteAccountWindow, __onConfirmedRef, __userEmailRef } = useConfirmDeleteAccountWindow();

const password = ref('');
const error = ref('');
const loading = ref(false);

const handleCancel = () => {
  password.value = '';
  error.value = '';
  hideConfirmDeleteAccountWindow();
};

const handleConfirm = async () => {
  error.value = '';

  if (!password.value) {
    error.value = 'Password is required to delete your account.';
    return;
  }

  const email = __userEmailRef?.value;
  if (!email) {
    error.value = 'Unable to retrieve user email.';
    return;
  }

  loading.value = true;

  try {
    await authApi.deleteUser({ email: email, password: password.value });
    clearAuthTokens();
    logout();

    const cb = __onConfirmedRef?.value;
    hideConfirmDeleteAccountWindow();

    if (typeof cb === 'function') {
      try { cb(); } catch (e) { /* swallow callback errors */ }
    }
  } catch (err) {
    console.error("Failed to delete account", err);
    error.value = err.response?.data?.message || "Failed to delete account.";
  } finally {
    loading.value = false;
  }
};

// Reset state when window is closed
watch(deleteAccountConfirmVisible, (visible) => {
  if (!visible) {
    password.value = '';
    error.value = '';
    loading.value = false;
  }
});

const footerButtons = computed(() => [
  { label: 'Cancel', onClick: handleCancel },
  { label: loading.value ? 'Deleting...' : 'Delete Account', onClick: handleConfirm, primary: true, disabled: loading.value },
]);
</script>

<style scoped>
.content {
  text-align: left;
  padding: 0 12px 12px;
}

.warning-text {
  color: #d00000;
  font-weight: 600;
  margin-bottom: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.label {
  font-weight: 600;
}

.password-input {
  padding: 8px 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1em;
}

.error {
  color: #d00000;
  margin-top: 8px;
  font-size: 0.9em;
}
</style>
