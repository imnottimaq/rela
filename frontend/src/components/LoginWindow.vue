<template>
    <WindowComponent
      title="Login"
      :buttons="[{ label: 'Close', onClick: hideLoginWindow }]"
      v-model:visible="loginVisible"
      storage-key="rela-window-login"
      footer-buttons-align="right"
      :footer-buttons="[
        { label: 'Cancel', onClick: hideLoginWindow },
        { label: 'Login', onClick: login, primary: true }
      ]"
      :initialSize="{ width: 200, height: 300 }"
      :minSize="{ width: 200, height: 300 }"
    >
    <div style="text-align: left; padding: 0 10px;">
        <h1>Login to your account</h1>
        <p>If you don't have an account, please register first.</p>
        <br>
      <div class="group" style="width: 100%">
        <label for="email">Email</label>
        <input id="email" type="email" v-model="email" />
      </div>
      <div class="group" style="width: 100%">
        <label for="password">Password</label>
        <input id="password" type="password" v-model="password" />
       </div>
    </div>
      
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { useLoginWindow } from '../composables/useLoginWindow';
import { ref } from 'vue';
import { authApi } from '../utils/http';

const { loginVisible, hideLoginWindow } = useLoginWindow();

const email = ref('');
const password = ref('');

const login = async () => {
  try {
    let response = await authApi.login(email.value, password.value);
    console.log('Login successful:', response.data);
    hideLoginWindow();
  } catch (error) {
    console.error('Login failed:', error);
  }
};



</script>
