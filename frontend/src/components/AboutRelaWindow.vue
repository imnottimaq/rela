<template>
  <WindowComponent
    title="About Rela"
    :buttons="[{ label: 'Close', onClick: handleCancel }]"
    v-model:visible="aboutRelaVisible"
    storage-key="rela-window-about"
    :initial-size="{ width: 480, height: 320 }"
    :min-size="{ width: 480, height: 320 }"
    :max-size="{ width: 480, height: 320 }"
    footer-buttons-align="right"
    :footer-buttons="[{ label: 'OK', onClick: handleCancel, primary: true }]"
    no-padding
  >
    <div class="about-vista">
      <div class="header">
        <img class="logo" src="../assets/rela.png" alt="Rela Logo" />
      </div>
      <div class="content">
        <p><strong>Rela Task Manager</strong></p>
        <p>Version 1.0.0 (WIP)</p>
        <p>&copy; 2025 Rela contributors. Zero rights reserved.</p>
        <hr />
        <p class="disclaimer">
          This software is a work-in-progress and is provided "as is" without warranty of any kind.
        </p>
      </div>
    </div>
  </WindowComponent>
</template>

<script setup>
import WindowComponent from './common/WindowComponent.vue';
import { useAboutRelaWindow } from '../composables/useAboutRelaWindow';

const { aboutRelaVisible, hideAboutRelaWindow } = useAboutRelaWindow();

const handleCancel = () => {
  hideAboutRelaWindow();
};
</script>

<style scoped>
.about-vista {
  display: flex;
  flex-direction: column;
  height: 100%;
  font-family: "Segoe UI", Frutiger, "Frutiger Linotype", "Dejavu Sans", "Helvetica Neue", Arial, sans-serif;
  font-size: 12px;
  color: #333;
  background-color: #f0f0f0;
  overflow: hidden;
}

.header {
  border-bottom: 1px solid #a9cde8;
}

.logo {
  width: 100%;
  height: auto;
  display: block;
}

.content {
  flex: 1;
  padding: 12px 24px;
  text-align: left;
}

.content p {
  margin: 4px 0;
}

.content hr {
  border: 0;
  border-top: 1px solid #ccc;
  margin: 12px 0;
}

.disclaimer {
  font-size: 11px;
  color: #666;
}
</style>
