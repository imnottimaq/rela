<template>
  <WindowComponent
    title="Edit Workspace"
    v-model:visible="modelVisible"
    :initial-size="{ width: 400, height: 450 }"
    :footer-buttons="footerButtons"
  >
    <div class="tabs">
      <button @click="activeTab = 'general'" :class="{ active: activeTab === 'general' }">General</button>
      <button @click="activeTab = 'members'" :class="{ active: activeTab === 'members' }">Members</button>
      <button @click="activeTab = 'danger'" :class="{ active: activeTab === 'danger' }">Danger Zone</button>
    </div>

    <div class="content">
      <!-- General Tab -->
      <div v-if="activeTab === 'general'">
        <div class="avatar-section">
          <img :src="avatarSrc" alt="Workspace Avatar" class="avatar-preview" />
          <input type="file" ref="fileInput" @change="handleFileChange" accept="image/png, image/jpeg" style="display: none" />
          <button @click="triggerFileInput" class="change-avatar-btn">Change Avatar</button>
        </div>
        <div class="name-section">
          <p>Enter a new name for the workspace:</p>
          <input v-model="newName" type="text" class="input-field" ref="nameInput" />
          <p v-if="error" class="error">{{ error }}</p>
        </div>
      </div>

      <!-- Members Tab -->
      <div v-if="activeTab === 'members'">
        <div class="members-section">
          <div v-for="member in workspaceMembers" :key="member._id" class="member-item">
            <img :src="getMemberAvatar(member)" alt="Member Avatar" class="member-avatar" />
            <span class="member-name">{{ member.name }}</span>
            <div class="member-actions">
              <button @click="handlePromote(member._id)" class="action-btn">Promote</button>
              <button @click="handleKick(member._id)" class="action-btn kick-btn">Kick</button>
            </div>
          </div>
          <div class="invite-link">
            <button @click="handleCreateInviteLink" class="action-btn">Create Invite Link</button>
          </div>
        </div>
      </div>

      <!-- Danger Zone Tab -->
      <div v-if="activeTab === 'danger'">
          <div class="danger-item">
            <h4>Delete Workspace</h4>
            <p>Once you delete a workspace, there is no going back. Please be certain.</p>
            <button @click="confirmDelete" class="action-btn kick-btn">Delete this workspace</button>
          </div>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { workspaceApi } from '../utils/http';
import { createWorkspaceInviteLink } from '../composables/useWorkspaces';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: false },
});

const emit = defineEmits(['update:visible', 'workspace-updated', 'workspace-deleted']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, "");
const wsId = computed(() => props.workspace?._id || props.workspace?.id);

const detailedWorkspace = ref({});
const activeTab = ref('general');
const newName = ref('');
const newAvatarFile = ref(null);
const avatarPreviewUrl = ref(null);
const error = ref('');
const nameInput = ref(null);
const fileInput = ref(null);
const loading = ref(false);

const workspaceName = computed(() => detailedWorkspace.value?.name);
const workspaceAvatar = computed(() => detailedWorkspace.value?.avatar);
const workspaceMembers = computed(() => detailedWorkspace.value?.memberDetails || []);

const avatarSrc = computed(() => {
  if (avatarPreviewUrl.value) {
    return avatarPreviewUrl.value;
  }
  if (workspaceAvatar.value) {
    return `${API_BASE_URL}/${workspaceAvatar.value}`;
  }
  return 'https://via.placeholder.com/128';
});

watch(() => props.visible, async (isVisible) => {
  if (isVisible && wsId.value) {
    activeTab.value = 'general';
    newAvatarFile.value = null;
    avatarPreviewUrl.value = null;
    error.value = '';

    loading.value = true;
    try {
      const { data } = await workspaceApi.getWorkspaceInfo(wsId.value);
      detailedWorkspace.value = data;
      newName.value = data.name;
    } catch (e) {
      console.error('Failed to fetch workspace info', e);
      error.value = 'Failed to load workspace details.';
      detailedWorkspace.value = {};
    } finally {
      loading.value = false;
    }

    nextTick(() => {
      if (activeTab.value === 'general' && nameInput.value) {
        nameInput.value.focus();
        nameInput.value.select();
      }
    });
  } else {
    detailedWorkspace.value = {};
    newName.value = '';
  }
});

const triggerFileInput = () => {
  fileInput.value.click();
};

const handleFileChange = (event) => {
  const file = event.target.files[0];
  if (file) {
    newAvatarFile.value = file;
    const reader = new FileReader();
    reader.onload = (e) => {
      avatarPreviewUrl.value = e.target.result;
    };
    reader.readAsDataURL(file);
  }
};

const getMemberAvatar = (member) => {
  if (member.avatar) {
    console.log(`${API_BASE_URL}/${member.avatar}`);
    return `${API_BASE_URL}/${member.avatar}`;
  }
  return 'https://via.placeholder.com/40';
};

const handlePromote = (memberId) => {
  console.log('Promoting member:', memberId, 'in workspace', wsId.value);
  try{
    workspaceApi.promoteMember(wsId.value, memberId);
  } catch (error){
    console.error('Failed to promote member', error);
  }
};

const handleKick = (memberId) => {
  console.log('Kicking member:', memberId, 'in workspace', wsId.value);
  try{
    workspaceApi.kickMember(wsId.value, memberId)
  }catch (error){
    console.error('Failed to kick member', error);
  }
};

const handleSave = async () => {
  const trimmedName = newName.value.trim();
  if (!trimmedName) {
    error.value = 'Workspace name cannot be empty.';
    return;
  }

  const hasNameChanged = trimmedName !== workspaceName.value;
  const hasAvatarChanged = !!newAvatarFile.value;

  if (!hasNameChanged && !hasAvatarChanged) {
    modelVisible.value = false;
    return;
  }

  loading.value = true;
  error.value = '';

  try {
    const payload = {};
    if (hasNameChanged) {
      payload.name = trimmedName;
    }

    if (hasAvatarChanged) {
      const formData = new FormData();
      formData.append('img', newAvatarFile.value);
      const { data } = await workspaceApi.uploadAvatar(wsId.value, formData);
      payload.avatar = data.avatar;
    }

    const { data: updatedWorkspace } = await workspaceApi.updateWorkspace(wsId.value, payload);

    emit('workspace-updated', {
      id: wsId.value,
      name: updatedWorkspace.name,
      avatar: updatedWorkspace.avatar,
    });
    modelVisible.value = false;
  } catch (e) {
    console.error('Failed to update workspace', e);
    error.value = 'Failed to update workspace.';
  } finally {
    loading.value = false;
  }
};

const handleCreateInviteLink = () => {
  createWorkspaceInviteLink(wsId.value);
};

const confirmDelete = () => {
  if (window.confirm('Are you sure you want to delete this workspace? This action cannot be undone.')) {
    handleDeleteWorkspace();
  }
};

const handleDeleteWorkspace = async () => {
  loading.value = true;
  error.value = '';
  try {
    await workspaceApi.deleteWorkspace(wsId.value);
    emit('workspace-deleted', wsId.value);
    modelVisible.value = false;
  } catch (e) {
    console.error('Failed to delete workspace', e);
    error.value = 'Failed to delete workspace.';
    window.alert(error.value); // Keep alert for critical errors
  } finally {
    loading.value = false;
  }
};

const handleCancel = () => {
  modelVisible.value = false;
};

const footerButtons = computed(() => {
  if (activeTab.value === 'general') {
    return [
      { label: 'Cancel', onClick: handleCancel },
      { label: 'Save', onClick: handleSave, primary: true },
    ];
  }
  return [{ label: 'Close', onClick: handleCancel, primary: true }];
});
</script>

<style scoped>
.tabs {
  display: flex;
  border-bottom: 1px solid #ccc;
}
.tabs button {
  padding: 10px 20px;
  border: none;
  background-color: transparent;
  cursor: pointer;
}
.tabs button.active {
  border-bottom: 2px solid #007bff;
  font-weight: bold;
}
.content {
  padding: 12px;
  text-align: center;
}
.avatar-section {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.avatar-preview {
  width: 128px;
  height: 128px;
  border-radius: 4px;
  object-fit: cover;
  border: 2px solid #ccc;
  margin-bottom: 12px;
}
.change-avatar-btn {
  padding: 8px 16px;
}
.name-section {
  text-align: left;
}
.input-field {
  width: 100%;
  padding: 8px;
  margin-top: 8px;
  box-sizing: border-box;
}
.error {
  color: #c00;
  margin-top: 8px;
}
.members-section {
  text-align: left;
}
.member-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}
.member-avatar {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  object-fit: cover;
  margin-right: 12px;
}
.member-name {
  flex-grow: 1;
}
.member-actions .action-btn {
  padding: 4px 8px;
  margin-left: 8px;
}
.kick-btn {
  background-color: #dc3545;
  color: red;
  border-color: #dc3545;
}
.danger-zone {
  text-align: left;
  border: 1px solid #dc3545;
  border-radius: 4px;
  padding: 12px;
}
.danger-item {
  margin-bottom: 16px;
}
.danger-item:last-child {
  margin-bottom: 0;
}
.danger-item h4 {
  margin: 0 0 4px;
  color: #dc3545;
}
.danger-item p {
  font-size: 12px;
  margin: 0 0 8px;
}
</style>
