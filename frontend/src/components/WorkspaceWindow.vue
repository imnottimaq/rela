<template>
  <WindowComponent
    :title="`Workspace: ${workspace?.name || 'Unnamed'}`"
    :buttons="[{ label: 'Close', onClick: close }]"
    v-model:visible="modelVisible"
    :initial-size="{ width: 360, height: 280 }"
    :min-size="{ width: 320, height: 240 }"
    :storage-key="`rela-window-workspace-${workspace?.id || workspace?._id || workspace?.name}`"
  >
    <div class="content">
      <p>This is a placeholder for workspace window.</p>
      <p>ID: <code>{{ workspace?._id || workspace?.id }}</code></p>
      <p>Name: <strong>{{ workspace?.name }}</strong></p>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed } from 'vue';
import WindowComponent from './WindowComponent.vue';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: true },
});

const emit = defineEmits(['update:visible', 'close']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const close = () => {
  emit('update:visible', false);
  emit('close');
};
</script>

<style scoped>
.content { text-align: left; padding: 0 12px 12px; }
</style>

