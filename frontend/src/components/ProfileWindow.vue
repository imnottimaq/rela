<template>
  <WindowComponent
    title="Profile"
    v-model:visible="profileVisible"
    storage-key="rela-window-profile"
    :initial-size="{ width: 350, height: 370 }"
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
        <div class="field" v-if="profile.createdAt">
          <span class="label">Created:</span>
          <span class="value">{{ formatDate(profile.createdAt) }}</span>
        </div>
        <img src="https://cataas.com/cat?height=250&width=350" alt="Random Cat" class="cat-image" />
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

let requestToken = 0;

const formatDate = (value) => {
  if (!value) return "";
  try {
    return new Date(value).toLocaleString();
  } catch (dateError) {
    return String(value);
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
    profile.value = response?.data || null;
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

.cat-image {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 auto;
}

</style>
