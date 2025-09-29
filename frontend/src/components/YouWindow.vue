<script setup>
import WindowComponent from "./WindowComponent.vue";
import { useAuth } from "../composables/useAuth.js";
import { authApi } from "../utils/http.js";
import { computed, ref, watch } from "vue";

const { isAuthenticated } = useAuth();
const profile = ref(null);
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "");

const getUserInfo = async () => {
  try {
    const response = await authApi.getUserInfo();
    const userProfile = response?.data || null;
    if (userProfile && userProfile.avatar) {
      userProfile.avatar = `${API_BASE_URL}/${userProfile.avatar}`;
    }
    profile.value = userProfile;
  } catch (error) {
    console.error("Failed to load profile", error);
    profile.value = null;
  }
};

watch(
  isAuthenticated,
  (isAuth) => {
    if (isAuth) {
      getUserInfo();
    } else {
      profile.value = null;
    }
  },
  { immediate: true }
);

const greeting = computed(() => {
  const hour = new Date().getHours();
  if (hour < 12) {
    return "Good morning";
  } else if (hour < 18) {
    return "Good afternoon";
  } else {
    return "Good evening";
  }
});

const initialPosition = computed(() => {
  if (typeof window === "undefined") {
    return { x: 0, y: 0 };
  }
  return { x: window.innerWidth - 270, y: 30 };
});
</script>

<template>
  <WindowComponent
    v-if="isAuthenticated"
    title="You"
    :initial-position="initialPosition"
    :initial-size="{ width: 250, height: 100 }"
    :closable="false"
    :movable="false"
    :resizable="false"
    :visible="isAuthenticated"
  >
    <div class="user-info">
      <img v-if="profile?.avatar" :src="profile.avatar" alt="User avatar" class="avatar" />
      <div v-else class="avatar-placeholder"></div>
      <div class="greeting">
        <p>{{ greeting }},</p>
        <p>{{ profile?.name }}</p>
      </div>
    </div>
  </WindowComponent>
</template>

<style scoped>
.user-info {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
}

.avatar,
.avatar-placeholder {
  width: 64px;
  height: 64px;
  border-radius: 4px;
  margin-right: 10px;
  border: 1px solid #000;
  object-fit: cover;
}

.avatar-placeholder {
  background-color: #ccc;
}

.greeting {
  display: flex;
  flex-direction: column;
}
</style>
