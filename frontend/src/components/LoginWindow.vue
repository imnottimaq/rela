<template>
    <WindowComponent
      title="Login"
      :buttons="[{ label: 'Close', onClick: hideLoginWindow }]"
      v-model:visible="loginVisible"
      storage-key="rela-window-login"
      footer-buttons-align="right"
      :footer-buttons="[
        { label: 'Cancel', onClick: hideLoginWindow },
        { label: 'Login', onClick: login, primary: true, loading: isSubmitting, disabled: isSubmitting }
      ]"
      :initialSize="{ width: 300, height: 370 }"
      :minSize="{ width: 300, height: 370 }"
    >
    <div style="text-align: left; padding: 0 10px;">
        <br>
        <h1 style="text-align: center;">Sign in to Rela</h1>
        <br>
      <div class="group" style="width: 100%">
        <input id="email" type="email" v-model="email" placeholder="E-mail" />
        <p v-if="emailError" class="error">{{ emailError }}</p>
      </div>
      <div class="group" style="width: 100%">
        <br>
        <input id="password" type="password" v-model="password" placeholder="Password"/>
        <p v-if="passwordError" class="error">{{ passwordError }}</p>
       </div>
      <button class="inline-link" type="button" @click="showForgotPasswordWindow">Forgot your password?</button>
      <br>
      <p v-if="loginError" class="error">{{ loginError }}</p>
      <br>
      <p style="">Don't have an account? <button class="inline-link" type="button" @click="showRegisterWindow">Register</button></p>
    </div>
      
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { useLoginWindow } from '../composables/useLoginWindow';
import { showForgotPasswordWindow } from '../composables/useForgotPasswordWindow';
import { hideRegisterWindow, showRegisterWindow } from '../composables/useRegisterWindow';
import { ref, watch } from 'vue';
import { authApi } from '../utils/http';
import { useAuth } from '../composables/useAuth';

const { loginVisible, hideLoginWindow } = useLoginWindow();
const { handleAuthSuccess } = useAuth();

const email = ref('');
const password = ref('');
const emailError = ref('');
const passwordError = ref('');
const loginError = ref('');
const isSubmitting = ref(false);

const clearForm = () => {
  email.value = '';
  password.value = '';
  emailError.value = '';
  passwordError.value = '';
  loginError.value = '';
  isSubmitting.value = false;
};

watch(loginVisible, (newValue) => {
  if (newValue === false) {
    clearForm();
  }
});

const validateEmail = (value) => {
  if (!value) {
    return false;
  }
  const normalized = value.trim().toLowerCase();
  const emailPattern = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/;
  return emailPattern.test(normalized);
};

const validatePassword = (value) => {
  if (!value || value.length < 8) {
    return false;
  }
  const hasLower = /[a-z]/.test(value);
  const hasUpper = /[A-Z]/.test(value);
  const hasDigit = /\d/.test(value);
  const hasSpecial = /[!"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]/.test(value);
  return hasLower && hasUpper && hasDigit && hasSpecial;
};

const login = async () => {
  try {
    emailError.value = '';
    passwordError.value = '';
    loginError.value = '';

    if (!validateEmail(email.value)) {
      emailError.value = 'Please enter a valid email address (e.g. user@example.com).';
    } else {
      email.value = email.value.trim().toLowerCase();
    }

    if (!validatePassword(password.value)) {
      passwordError.value = 'Password must be at least 8 characters and include upper, lower, digit, and special character.';
    }

    if (emailError.value || passwordError.value) {
      return;
    }
    if (isSubmitting.value) return;
    isSubmitting.value = true;
    const response = await authApi.login(email.value, password.value);
    const { token, refreshToken } = response?.data || {};
    const stored = handleAuthSuccess({ token, refreshToken });
    if (!stored) {
      loginError.value = 'Login failed. Missing access token in response.';
      return;
    }
    console.log('Login successful:', response.data);
    hideLoginWindow();
    hideRegisterWindow();
  } catch (error) {
    console.error('Login failed:', error);
    loginError.value = 'Login failed. Please check your credentials and try again.';
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<style scoped>
.inline-link {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  color: #0066cc;
  cursor: pointer;
  text-decoration: underline;
}

.inline-link:hover {
  color: #004b99;
}
</style>
