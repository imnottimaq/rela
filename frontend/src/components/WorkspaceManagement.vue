<template>
  <div class="workspace-management">
    <ManageButton @click="editWorkspaceVisible = true" />

    <EditWorkspaceWindow
      :workspace="workspace"
      v-model:visible="editWorkspaceVisible"
      @save="handleEditWorkspace"
      @promote="handlePromote"
      @kick="handleKick"
    />

    <DeleteWorkspaceWindow
      :workspace="workspace"
      v-model:visible="deleteWorkspaceVisible"
      @confirm="handleDeleteWorkspace"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import EditWorkspaceWindow from './EditWorkspaceWindow.vue';
import DeleteWorkspaceWindow from './DeleteWorkspaceWindow.vue';
import { workspaceApi } from '../utils/http';
import { createWorkspaceInviteLink } from '../composables/useWorkspaces';

const props = defineProps({
  workspace: { type: Object, required: true },
});

const emit = defineEmits(['workspace-updated', 'workspace-deleted', 'update:menu-items']);

const editWorkspaceVisible = ref(false);
const deleteWorkspaceVisible = ref(false);
const loading = ref(false);
const error = ref('');

const wsId = computed(() => props.workspace?._id || props.workspace?.id);

const handleEditWorkspace = async ({ name, avatar }) => {
  if (!name && !avatar) {
    editWorkspaceVisible.value = false;
    return;
  }

  loading.value = true;
  error.value = '';

  try {
    const updatedFields = { id: wsId.value };

    if (name) {
      await workspaceApi.updateWorkspace(wsId.value, { name });
      updatedFields.name = name;
    }

    if (avatar) {
      const formData = new FormData();
      formData.append('img', avatar);
      const { data } = await workspaceApi.uploadAvatar(wsId.value, formData);
      updatedFields.avatar = data.avatar;
    }

    emit('workspace-updated', updatedFields);
    editWorkspaceVisible.value = false;
  } catch (e) {
    console.error('Failed to update workspace', e);
    error.value = 'Failed to update workspace.';
    window.alert(error.value);
  } finally {
    loading.value = false;
  }
};

const handleDeleteWorkspace = async () => {
  loading.value = true;
  error.value = '';
  try {
    await workspaceApi.deleteWorkspace(wsId.value);
    deleteWorkspaceVisible.value = false;
    emit('workspace-deleted', wsId.value);
  } catch (e) {
    console.error('Failed to delete workspace', e);
    error.value = 'Failed to delete workspace.';
    window.alert(error.value);
  } finally {
    loading.value = false;
  }
};

const handlePromote = (memberId) => {
  console.log('Promoting member:', memberId);
  // workspaceApi.promoteMember(wsId.value, memberId)... 
};

const handleKick = (memberId) => {
  console.log('Kicking member:', memberId);
  // workspaceApi.kickMember(wsId.value, memberId)...
};

const menuItems = computed(() => [
  {
    label: 'Delete Workspace',
    onClick: () => (deleteWorkspaceVisible.value = true),
    type: 'button',
  },
  {
    label: 'Create and Copy Invite Link',
    onClick: () => createWorkspaceInviteLink(wsId.value),
    type: 'button',
  },
]);

watch(menuItems, (newItems) => {
  emit('update:menu-items', newItems);
}, { immediate: true, deep: true });
</script>

<style scoped>
.workspace-management {
  margin-left: auto;
}
</style>
