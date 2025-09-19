import { ref } from "vue";
import {
  getAccessToken,
  setAuthTokens,
  clearAuthTokens,
  onAuthStateChange,
} from "../utils/http";

const isAuthenticated = ref(Boolean(getAccessToken()));

onAuthStateChange((hasToken) => {
  isAuthenticated.value = hasToken;
});

const handleAuthSuccess = ({ token, refreshToken }) => {
  if (!token) {
    return false;
  }
  setAuthTokens({ accessToken: token, refreshToken });
  return true;
};

const logout = () => {
  console.log("Logout placeholder: send HTTP request later");
  clearAuthTokens();
};

export function useAuth() {
  return {
    isAuthenticated,
    handleAuthSuccess,
    logout,
  };
}

