<template>
  <WindowComponent
    :title="windowTitle"
    v-model:visible="internalVisible"
    :closable="true"
    :initial-size="{ width: 400, height: 250 }"
    :footer-buttons="footerButtons"
    storage-key="rela-window-invite"
  >
    <div class="content">
      <template v-if="loading">
        <p>Loading invite...</p>
      </template>
      <template v-else-if="error">
        <p class="error">{{ error }}</p>
      </template>
      <template v-else-if="!isAuthenticated">
        <p>You need to be logged in to accept this invite.</p>
        <p>Please log in or create an account.</p>
      </template>
      <template v-else-if="workspace">
        <div class="workspace-info">
          <img v-if="workspace.avatar" :src="workspace.avatar" alt="Workspace Avatar" class="avatar-image" />
          <div v-else class="avatar-placeholder"></div>
          <div class="details">
            <p>You have been invited to join:</p>
            <h3>{{ workspace.name }}</h3>
          </div>
        </div>
      </template>
      <template v-else>
        <p>This invite is invalid or has expired.</p>
      </template>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { useAuth } from '../composables/useAuth.js';
import { workspaceApi } from '../utils/http.js';
import { showLoginWindow } from "../composables/useLoginWindow";

const props = defineProps({
  modelValue: Boolean,
  joinToken: String,
});

const emit = defineEmits(['update:modelValue', 'joined']);

const internalVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
});

const { isAuthenticated } = useAuth();

const workspace = ref(null);
const loading = ref(false);
const error = ref('');
const internalJoinToken = ref(null);
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "");

const fetchWorkspaceDetails = async () => {
  if (!internalJoinToken.value) return;

  loading.value = true;
  error.value = '';
  try {
    const { data } = await workspaceApi.getWorkspaceByInviteToken(internalJoinToken.value);
    if (data && data.id) {
      if (data.avatar) {
        data.avatar = `${API_BASE_URL}/${data.avatar}`;
      }
      workspace.value = data;
    } else {
      error.value = 'Invalid or expired invite link.';
      workspace.value = null;
    }
  } catch (e) {
    console.error('Failed to fetch workspace details', e);
    error.value = e.response?.data?.error || 'Invalid or expired invite link.';
    workspace.value = null;
  } finally {
    loading.value = false;
  }
};

const resetComponentState = (clearToken = false) => {
  workspace.value = null;
  error.value = '';
  loading.value = false;
  if (clearToken) {
    localStorage.removeItem('rela_join_token');
    internalJoinToken.value = null;
  }
};

onMounted(() => {
  const token = props.joinToken || localStorage.getItem('rela_join_token');
  if (token) {
    internalJoinToken.value = token;
    localStorage.setItem('rela_join_token', token);
  }
});

watch(() => props.joinToken, (newToken) => {
  if (newToken && newToken !== internalJoinToken.value) {
    internalJoinToken.value = newToken;
    localStorage.setItem('rela_join_token', newToken);
    resetComponentState(false);
  }
});

watch(() => [internalVisible.value, isAuthenticated.value, internalJoinToken.value], ([visible, isAuth, token]) => {
  if (!visible) {
    resetComponentState(true);
    return;
  }

  if (token) {
    if (isAuth) {
      if (!workspace.value && !loading.value && !error.value) {
        fetchWorkspaceDetails();
      }
    } else {
      resetComponentState(false);
    }
  }
}, { immediate: true });


const handleLogin = () => {
  showLoginWindow();
};

const handleJoin = async () => {
  loading.value = true; error.value = '';
  try {
    await workspaceApi.acceptInvite(internalJoinToken.value);
    emit('joined');
    internalVisible.value = false;
  } catch (e) {
    console.error('Failed to join workspace', e);
    error.value = e.response?.data?.error || 'Could not join the workspace.';
  } finally {
    loading.value = false;
  }
};

const handleCancel = () => {
  internalVisible.value = false;
};

const windowTitle = computed(() => {
  if (workspace.value) return `Join ${workspace.value.name}`;
  return 'Workspace Invitation';
});

const footerButtons = computed(() => {
  if (!isAuthenticated.value) {
    return [
      { label: 'Cancel', onClick: handleCancel },
      { label: 'Login', onClick: handleLogin, primary: true },
    ];
  }
  if (workspace.value) {
    return [
      { label: 'Decline', onClick: handleCancel },
      { label: 'Join Workspace', onClick: handleJoin, primary: true, loading: loading.value, disabled: loading.value },
    ];
  }
  return [{ label: 'Close', onClick: handleCancel }];
});

</script>

<style scoped>
.content {
  padding: 12px;
  text-align: left;
}
.error {
  color: #c00;
}
.workspace-info {
  display: flex;
  align-items: center;
  gap: 16px;
}
.avatar-image,
.avatar-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid #ccc;
}
.avatar-placeholder {
  background-color: #eee;
}
.details h3 {
  margin: 0;
  font-size: 1.2em;
}
.details p {
  margin: 0;
}
</style>
