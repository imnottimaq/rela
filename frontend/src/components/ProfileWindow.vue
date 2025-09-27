<template>
  <WindowComponent
    title="Profile"
    v-model:visible="profileVisible"
    storage-key="rela-window-profile"
    :initial-size="{ width: 350, height: 420 }"
    :min-size="{ width: 300, height: 320 }"
    footer-buttons-align="right"
    :footer-buttons="footerButtons"
    :buttons="[{ label: 'Close', onClick: hideProfileWindow }]"
  >
    <div class="content">
      <template v-if="loading">
        <p>Loading profile...</p>
      </template>
      <template v-else-if="error">
        <p class="error">{{ error }}</p>
      </template>
      <template v-else-if="profile">
        <div class="avatar-container">
          <img :src="profile.avatar" alt="User Avatar" class="avatar-image" />
          <input type="file" @change="uploadAvatar" accept="image/png, image/jpeg" ref="fileInput" class="file-input"/>
          <button @click="triggerFileInput" class="upload-button">Upload Avatar</button>
        </div>
        <div class="field">
          <span class="label">Name:</span>
          <span class="value">{{ profile.name }}</span>
        </div>
        <div class="field">
          <span class="label">Email:</span>
          <span class="value">{{ profile.email }}</span>
        </div>
        <div class="field" v-if="profile.role">
          <span class="label">Role:</span>
          <span class="value">{{ profile.role }}</span>
        </div>
      </template>
      <template v-else>
        <p>No profile data available.</p>
      </template>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import WindowComponent from "./WindowComponent.vue";
import { useProfileWindow } from "../composables/useProfileWindow";
import { useAuth } from "../composables/useAuth";
import { showConfirmLogoutWindow } from "../composables/useConfirmLogoutWindow";
import { authApi } from "../utils/http";

const { profileVisible, hideProfileWindow } = useProfileWindow();
const { isAuthenticated } = useAuth();

const profile = ref(null);
const loading = ref(false);
const error = ref("");
const fileInput = ref(null);
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "");

let requestToken = 0;

const triggerFileInput = () => {
  fileInput.value.click();
};

const uploadAvatar = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("img", file);

  loading.value = true;
  error.value = "";

  try {
    await authApi.uploadAvatar(formData);
    await loadProfile(); 
  } catch (uploadError) {
    console.error("Failed to upload avatar", uploadError);
    error.value = "Failed to upload avatar.";
  } finally {
    loading.value = false;
  }
};

const loadProfile = async () => {
  requestToken += 1;
  const currentToken = requestToken;
  loading.value = true;
  error.value = "";

  try {
    const response = await authApi.getUserInfo();
    if (currentToken !== requestToken) {
      return;
    }
    const userProfile = response?.data || null;
    if (userProfile && userProfile.avatar) {
      userProfile.avatar = `${API_BASE_URL}/${userProfile.avatar}`;
    }
    profile.value = userProfile;
  } catch (loadError) {
    if (currentToken !== requestToken) {
      return;
    }
    console.error("Failed to load profile", loadError);
    error.value = "Failed to load profile information.";
    profile.value = null;
  } finally {
    if (currentToken === requestToken) {
      loading.value = false;
    }
  }
};

watch(
  () => profileVisible.value,
  (visible) => {
    if (visible) {
      loadProfile();
    } else {
      profile.value = null;
      error.value = "";
      requestToken += 1;
    }
  }
);

watch(
  () => isAuthenticated.value,
  (authenticated) => {
    if (!authenticated && profileVisible.value) {
      hideProfileWindow();
    }
  }
);

const handleLogout = () => {
  showConfirmLogoutWindow(() => hideProfileWindow());
};

const footerButtons = computed(() => [
  { label: "Cancel", onClick: hideProfileWindow },
  { label: "Logout", onClick: handleLogout, primary: true },
]);

</script>

<style scoped>
.content {
  text-align: left;
  padding: 0 12px 12px;
}

.avatar-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 16px;
}

.avatar-image {
  width: 128px;
  height: 128px;
  border-radius: 4px;
  object-fit: cover;
  border: 2px solid #ccc;
  margin-bottom: 12px;
}

.file-input {
  display: none;
}

.upload-button {
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
  background-color: #f0f0f0;
}

.field {
  display: flex;
  margin-bottom: 8px;
  gap: 8px;
}

.label {
  font-weight: 600;
  min-width: 80px;
}

.value {
  flex: 1;
  word-break: break-word;
}

.error {
  color: #d00000;
}
</style>
