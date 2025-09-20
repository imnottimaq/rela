<template>
  <WindowComponent
    title="Create Workspace"
    :buttons="[{ label: 'Close', onClick: onCancel }]"
    v-model:visible="createWorkspaceVisible"
    storage-key="rela-window-create-workspace"
    footer-buttons-align="right"
    :footer-buttons="[
      { label: 'Cancel', onClick: onCancel },
      { label: 'Create', onClick: onCreate, primary: true }
    ]"
    :initial-size="{ width: 320, height: 220 }"
    :min-size="{ width: 320, height: 220 }"
  >
    <div style="text-align: left; padding: 0 10px;">
      <h1>New workspace</h1>
      <div class="group" style="width: 100%">
        <label for="ws-name">Name</label>
        <input id="ws-name" type="text" v-model="name" @keydown.enter.prevent="onCreate" />
        <p v-if="error" class="error">{{ error }}</p>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { workspaceApi } from '../utils/http';
import { useWorkspaces } from '../composables/useWorkspaces';
import { openWorkspaceWindow } from '../composables/useWorkspaces';
import { useCreateWorkspaceWindow } from '../composables/useCreateWorkspaceWindow';

const { refreshWorkspaces } = useWorkspaces();
const { createWorkspaceVisible, hideCreateWorkspaceWindow } = useCreateWorkspaceWindow();

const name = ref('');
const error = ref('');

const reset = () => {
  name.value = '';
  error.value = '';
};

const onCancel = () => {
  hideCreateWorkspaceWindow();
  reset();
};

const onCreate = async () => {
  error.value = '';
  const trimmed = name.value.trim();
  if (!trimmed) {
    error.value = 'Please enter workspace name';
    return;
  }
  try {
    const { data } = await workspaceApi.createWorkspace({ name: trimmed });
    // Open newly created workspace window immediately using returned data
    if (data) {
      openWorkspaceWindow(data);
    }
    // Refresh list to sync menus and cache, but don't auto-open others
    await refreshWorkspaces({ skipRestore: true });
    onCancel();
  } catch (e) {
    console.error('Create workspace failed', e);
    error.value = 'Failed to create workspace';
  }
};
</script>

<style scoped>
.error { color: #d00000; }
</style>
