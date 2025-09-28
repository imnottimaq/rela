<template>
    <WindowComponent
      title="Register"
      :buttons="[{ label: 'Close', onClick: hideRegisterWindow }]"
      v-model:visible="registerVisible"
      storage-key="rela-window-register"
      footer-buttons-align="right"
      :footer-buttons="[
        { label: 'Cancel', onClick: hideRegisterWindow },
        { label: 'Register', onClick: registerUser, primary: true, loading: isSubmitting, disabled: isSubmitting }
      ]"
      :initialSize="{ width: 320, height: 460 }"
      :minSize="{ width: 320, height: 460 }"
    >
    <div style="text-align: left; padding: 0 10px;">
        <h1>Create an account</h1>
        <p>Already have an account? <button class="inline-link" type="button" @click="showLoginWindow">Login here</button></p>
        <br>
      <div class="group" style="width: 100%">
        <label for="name">Name</label>
        <input id="name" type="text" v-model="name" aria-describedby="name-error" />
        <div v-if="nameError" class="error" id="name-error" role="tooltip">{{ nameError }}</div>
      </div>
      <div class="group" style="width: 100%">
        <label for="email">Email</label>
        <input id="email" type="email" v-model="email" aria-describedby="email-error" />
        <div v-if="emailError" class="error" id="email-error" role="tooltip">{{ emailError }}</div>
      </div>
      <div class="group" style="width: 100%">
        <label for="password">Password</label>
        <input id="password" type="password" v-model="password" aria-describedby="password-error" />
        <div v-if="passwordError" class="error" id="password-error" role="tooltip">{{ passwordError }}</div>
       </div>
      <div class="group" style="width: 100%">
        <label for="confirm-password">Confirm password</label>
        <input id="confirm-password" type="password" v-model="confirmPassword" aria-describedby="confirm-password-error" />
        <div v-if="confirmPasswordError" class="error" id="confirm-password-error" role="tooltip">{{ confirmPasswordError }}</div>
       </div>
      <p v-if="registerError" class="error">{{ registerError }}</p>
    </div>
      
    </WindowComponent>
</template>
<script setup>
import { ref, watch } from 'vue';
import WindowComponent from './WindowComponent.vue';
import { useRegisterWindow } from '../composables/useRegisterWindow';
import { hideLoginWindow, showLoginWindow } from '../composables/useLoginWindow';
import { authApi } from '../utils/http';
import { useAuth } from '../composables/useAuth';

const { registerVisible, hideRegisterWindow } = useRegisterWindow();
const { handleAuthSuccess } = useAuth();

const name = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');

const nameError = ref('');
const emailError = ref('');
const passwordError = ref('');
const confirmPasswordError = ref('');
const registerError = ref('');
const isSubmitting = ref(false);

const resetState = () => {
  name.value = '';
  email.value = '';
  password.value = '';
  confirmPassword.value = '';
  nameError.value = '';
  emailError.value = '';
  passwordError.value = '';
  confirmPasswordError.value = '';
  registerError.value = '';
};

watch(registerVisible, (newValue) => {
  if (newValue === false) {
    resetState();
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
  const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(value);
  return hasLower && hasUpper && hasDigit && hasSpecial;
};

const validateName = (value) => {
  if (!value) {
    return false;
  }
  return value.trim().length >= 2;
};

const registerUser = async () => {
  try {
    nameError.value = '';
    emailError.value = '';
    passwordError.value = '';
    confirmPasswordError.value = '';
    registerError.value = '';

    if (!validateName(name.value)) {
      nameError.value = 'Please enter your name (at least 2 characters).';
    } else {
      name.value = name.value.trim();
    }

    if (!validateEmail(email.value)) {
      emailError.value = 'Please enter a valid email address (e.g. user@example.com).';
    } else {
      email.value = email.value.trim().toLowerCase();
    }

    if (!validatePassword(password.value)) {
      passwordError.value = 'Password must be at least 8 characters and include upper, lower, digit, and special character.';
    }

    if (password.value !== confirmPassword.value) {
      confirmPasswordError.value = 'Passwords do not matter.';
    }

    if (nameError.value || emailError.value || passwordError.value || confirmPasswordError.value) {
      return;
    }

    if (isSubmitting.value) return;
    isSubmitting.value = true;
    const response = await authApi.createUser(name.value, email.value, password.value);
    const { token, refreshToken } = response?.data || {};
    const stored = handleAuthSuccess({ token, refreshToken });
    if (!stored) {
      registerError.value = 'Registration succeeded without an access token. Please try logging in.';
      return;
    }
    hideRegisterWindow();
    hideLoginWindow();
  } catch (error) {
    console.error('Registration failed:', error);
    registerError.value = 'Registration failed. Please check your details and try again.';
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
