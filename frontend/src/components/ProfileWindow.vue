<template>
  <WindowComponent
    title="Profile"
    v-model:visible="profileVisible"
    storage-key="rela-window-profile"
    :initial-size="{ width: 350, height: 550 }"
    :min-size="{ width: 300, height: 450 }"
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
          <input type="text" v-model="editableProfile.name" class="value-input" />
        </div>
        <div class="field">
          <span class="label">Email:</span>
          <input type="email" v-model="editableProfile.email" class="value-input" />
        </div>
        <div class="field" v-if="profile.role">
          <span class="label">Role:</span>
          <span class="value">{{ profile.role }}</span>
        </div>
        <div class="field">
          <span class="label">New Password:</span>
          <input type="password" v-model="newPassword" class="value-input" placeholder="Leave blank to keep current" />
        </div>
        <div class="field">
          <span class="label">Confirm Password:</span>
          <input type="password" v-model="confirmPassword" class="value-input" />
        </div>

        <p v-if="saveError" class="error">{{ saveError }}</p>
        <p v-if="saveSuccess" class="success">Profile updated successfully!</p>

        <div class="button-group">
          <button @click="saveProfile" :disabled="saveLoading" class="action-button primary">
            {{ saveLoading ? 'Saving...' : 'Save Changes' }}
          </button>
        </div>

        <div class="danger-zone">
          <h3 class="danger-zone-title">Danger Zone</h3>
          <div class="danger-buttons">
            <button @click="handleLogout" class="action-button">Logout</button>
            <button @click="confirmDeleteAccount" :disabled="deleteLoading" class="action-button delete">
              {{ deleteLoading ? 'Deleting...' : 'Delete Account' }}
            </button>
          </div>
          <p v-if="deleteError" class="error">{{ deleteError }}</p>
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
import WindowComponent from "./common/WindowComponent.vue";
import { useProfileWindow } from "../composables/useProfileWindow";
import { useAuth } from "../composables/useAuth";
import { showConfirmLogoutWindow } from "../composables/useConfirmLogoutWindow";
import { authApi, clearAuthTokens } from "../utils/http";
import defaultAvatar from '/default-avatar.jpg';

const { profileVisible, hideProfileWindow } = useProfileWindow();
const { isAuthenticated, logout } = useAuth();

const profile = ref(null);
const editableProfile = ref({ name: '', email: '' });
const newPassword = ref('');
const confirmPassword = ref('');
const loading = ref(false);
const saveLoading = ref(false);
const deleteLoading = ref(false);
const error = ref("");
const saveError = ref("");
const deleteError = ref("");
const saveSuccess = ref(false);
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
  saveError.value = "";
  deleteError.value = "";
  saveSuccess.value = false;

  try {
    const response = await authApi.getUserInfo();
    if (currentToken !== requestToken) {
      return;
    }
    const userProfile = response?.data || null;
    if (userProfile && userProfile.avatar) {
      userProfile.avatar = `${API_BASE_URL}/${userProfile.avatar}`;
    } else if (!userProfile.avatar) {
      userProfile.avatar = defaultAvatar;
    }
    profile.value = userProfile;
    if (userProfile) {
      editableProfile.value.name = userProfile.name;
      editableProfile.value.email = userProfile.email;
    }
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

const saveProfile = async () => {
  saveLoading.value = true;
  saveError.value = "";
  saveSuccess.value = false;

  if (newPassword.value && newPassword.value !== confirmPassword.value) {
    saveError.value = "New password and confirm password do not match.";
    saveLoading.value = false;
    return;
  }

  const payload = {
    name: editableProfile.value.name,
    email: editableProfile.value.email,
  };

  if (newPassword.value) {
    payload.password = newPassword.value;
  }

  try {
    await authApi.updateUserInfo(payload);
    saveSuccess.value = true;
    newPassword.value = '';
    confirmPassword.value = '';
    await loadProfile(); // Reload profile to reflect changes
  } catch (err) {
    console.error("Failed to update profile", err);
    saveError.value = err.response?.data?.message || "Failed to update profile.";
  } finally {
    saveLoading.value = false;
  }
};

const confirmDeleteAccount = () => {
  showConfirmDialog({
    title: "Delete Account",
    message: "Are you sure you want to delete your account? This action cannot be undone. Please enter your password to confirm.",
    confirmText: "Delete",
    confirmButtonClass: "delete",
    showInput: true,
    inputType: "password",
    onConfirm: (password) => deleteAccount(password),
  });
};

const deleteAccount = async (password) => {
  deleteLoading.value = true;
  deleteError.value = "";

  if (!password) {
    deleteError.value = "Password is required to delete your account.";
    deleteLoading.value = false;
    return;
  }

  try {
    await authApi.deleteUser({ password });
    clearAuthTokens();
    logout();
    hideProfileWindow();
  } catch (err) {
    console.error("Failed to delete account", err);
    deleteError.value = err.response?.data?.message || "Failed to delete account.";
  } finally {
    deleteLoading.value = false;
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
      saveError.value = "";
      deleteError.value = "";
      saveSuccess.value = false;
      newPassword.value = '';
      confirmPassword.value = '';
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
  { label: "Close", onClick: hideProfileWindow },
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
  align-items: center;
  margin-bottom: 8px;
  gap: 8px;
}

.label {
  font-weight: 600;
  min-width: 120px; /* Adjusted for better alignment */
}

.value-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1em;
}

.value {
  flex: 1;
  word-break: break-word;
}

.error {
  color: #d00000;
  margin-top: 8px;
  text-align: center;
}

.success {
  color: #28a745;
  margin-top: 8px;
  text-align: center;
}

.danger-zone {
  border: 1px solid #d00000;
  border-radius: 4px;
  padding: 12px;
  margin-top: 24px;
}

.danger-zone-title {
  color: #d00000;
  margin: 0 0 12px;
  text-align: center;
  font-size: 1.2em;
  font-weight: 600;
}

.danger-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.danger-buttons .action-button {
  width: 100%;
}

.button-group {
    margin-top: 16px;
}
</style>
