import { ref } from "vue";

const profileVisible = ref(false);

export const showProfileWindow = () => {
  profileVisible.value = true;
};

export const hideProfileWindow = () => {
  profileVisible.value = false;
};

export function useProfileWindow() {
  return {
    profileVisible,
    showProfileWindow,
    hideProfileWindow,
  };
}

