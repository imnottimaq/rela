<template>
    <WindowComponent
      title="Login"
      :buttons="[{ label: 'Close', onClick: hideLoginWindow }]"
      v-model:visible="loginVisible"
      storage-key="rela-window-login"
      footer-buttons-align="right"
      :footer-buttons="[
        { label: 'Cancel', onClick: hideLoginWindow },
        { label: 'Login', onClick: hideLoginWindow }
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
        <input id="email" type="text" />
      </div>
      <div class="group" style="width: 100%">
        <label for="password">Password</label>
        <input id="password" type="password" />
       </div>
    </div>
      
    </WindowComponent>
</template>
<script setup>
import WindowComponent from './WindowComponent.vue';
import { useLoginWindow } from '../composables/useLoginWindow';

const { loginVisible, hideLoginWindow } = useLoginWindow();
</script>
