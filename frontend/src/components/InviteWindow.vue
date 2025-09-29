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
        <p class="debug-info">Token: {{ internalJoinToken }}</p>
        <p class="debug-info">Authenticated: {{ isAuthenticated }}</p>
      </template>
      <template v-else-if="!isAuthenticated">
        <p>You've been invited to join a workspace.</p>
        <p>Please log in or create an account to see the details and join.</p>
      </template>
      <template v-else-if="workspace">
        <div class="workspace-info">
          <img v-if="workspace.avatar" :src="workspace.avatar" alt="Workspace Avatar" class="avatar-image" />
          <div v-else class="avatar-placeholder"></div>
          <div class="details">
            <p>You're about to join this workspace:</p>
            <h3>{{ workspace.name }}</h3>
          </div>
        </div>
      </template>
      <template v-else>
        <p>This invite is invalid or has expired.</p>
        <p class="debug-info">Token: {{ internalJoinToken }}</p>
        <p class="debug-info">Authenticated: {{ isAuthenticated }}</p>
      </template>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import WindowComponent from './common/WindowComponent.vue';
import { useAuth } from '../composables/useAuth.js';
import { workspaceApi } from '../utils/http.js';
import { showLoginWindow } from "../composables/useLoginWindow";
import defaultAvatar from '/default-workspace.ico';

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
  if (!internalJoinToken.value) {
    console.log('[InviteWindow] No join token available');
    return;
  }

  console.log('[InviteWindow] Fetching workspace details for token:', internalJoinToken.value);
  loading.value = true;
  error.value = '';
  workspace.value = null;

  try {
    const response = await workspaceApi.getWorkspaceByInviteToken(internalJoinToken.value);
    console.log('[InviteWindow] Full response:', response);
    console.log('[InviteWindow] Response data:', response.data);

    const data = response.data;


    // Проверяем наличие id в ответе
    if (data && (data.id || data._id)) {
      const workspaceData = {
        id: data.id || data._id,
        name: data.name,
        avatar: data.avatar ? `${API_BASE_URL}/${data.avatar}` : defaultAvatar,
        ownedBy: data.owned_by || data.ownedBy,
        members: data.members
      };

      workspace.value = workspaceData;
      console.log('[InviteWindow] Workspace loaded:', workspace.value);
    } else {
      error.value = 'Invalid or expired invite link.';
      workspace.value = null;
      console.log('[InviteWindow] Invalid response structure:', data);
    }
  } catch (e) {
    console.error('[InviteWindow] Failed to fetch workspace details', e);
    console.error('[InviteWindow] Error response:', e.response);
    error.value = e.response?.data?.error || e.response?.data?.message || 'Invalid or expired invite link.';
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
  console.log('[InviteWindow] Mounted, token from props:', props.joinToken);
  console.log('[InviteWindow] Mounted, token from localStorage:', localStorage.getItem('rela_join_token'));
  console.log('[InviteWindow] Authenticated:', isAuthenticated.value);

  if (token) {
    internalJoinToken.value = token;
    localStorage.setItem('rela_join_token', token);
    console.log('[InviteWindow] Token set:', token);
  }
});

watch(() => props.joinToken, (newToken) => {
  console.log('[InviteWindow] joinToken prop changed:', newToken);
  if (newToken && newToken !== internalJoinToken.value) {
    internalJoinToken.value = newToken;
    localStorage.setItem('rela_join_token', newToken);
    resetComponentState(false);
  }
});

watch(internalVisible, (isVisible, wasVisible) => {
  console.log('[InviteWindow] Visibility changed:', isVisible);
  if (!isVisible && wasVisible) {
    resetComponentState(true);
  }
});

watch(() => [internalVisible.value, isAuthenticated.value, internalJoinToken.value],
    ([visible, isAuth, token]) => {
      console.log('[InviteWindow] Watch triggered - visible:', visible, 'auth:', isAuth, 'token:', token);

      if (!visible) {
        return;
      }

      if (token) {
        if (isAuth) {
          if (!workspace.value && !loading.value && !error.value) {
            console.log('[InviteWindow] Conditions met, fetching workspace details');
            fetchWorkspaceDetails();
          }
        } else {
          console.log('[InviteWindow] User not authenticated, resetting state');
          resetComponentState(false);
        }
      } else {
        console.log('[InviteWindow] No token available');
      }
    },
    { immediate: true }
);

const handleLogin = () => {
  showLoginWindow();
};

const handleJoin = async () => {
  console.log('[InviteWindow] Attempting to join workspace');
  loading.value = true;
  error.value = '';

  try {
    const response = await workspaceApi.acceptInvite(internalJoinToken.value);
    console.log('[InviteWindow] Join successful:', response.data);
    emit('joined');
    internalVisible.value = false;
  } catch (e) {
    console.error('[InviteWindow] Failed to join workspace', e);
    console.error('[InviteWindow] Error response:', e.response?.data);
    error.value = e.response?.data?.error || e.response?.data?.message || 'Could not join the workspace.';
  } finally {
    loading.value = false;
  }
};

const handleCancel = () => {
  internalVisible.value = false;
  localStorage.removeItem("rela_join_token");
};

const windowTitle = computed(() => {
  if (workspace.value) return `Join ${workspace.value.name}`;
  return 'Workspace Invitation';
});

const footerButtons = computed(() => {
  if (!isAuthenticated.value) {
    return [
      { label: 'Cancel', onClick: handleCancel },
      { label: 'Login to Join', onClick: handleLogin, primary: true },
    ];
  }
  if (workspace.value) {
    return [
      { label: 'Cancel', onClick: handleCancel },
      { label: 'Join', onClick: handleJoin, primary: true, loading: loading.value, disabled: loading.value },
    ];
  }
  if (loading.value) return [];
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
.debug-info {
  font-size: 0.85em;
  color: #666;
  margin-top: 8px;
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