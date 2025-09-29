<template>
  <WindowComponent
    title="Delete Workspace"
    v-model:visible="modelVisible"
    :initial-size="{ width: 380, height: 180 }"
    :footer-buttons="footerButtons"
  >
    <div class="content">
      <p>Are you sure you want to delete the workspace "<strong>{{ workspace.name }}</strong>"?</p>
      <p>This action cannot be undone.</p>
    </div>
  </WindowComponent>
</template>

<script setup>
import { computed } from 'vue';
import WindowComponent from './common/WindowComponent.vue';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: false },
});

const emit = defineEmits(['update:visible', 'confirm']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const handleConfirm = () => {
  emit('confirm');
};

const handleCancel = () => {
  modelVisible.value = false;
};

const footerButtons = computed(() => [
  { label: 'Cancel', onClick: handleCancel },
  { label: 'Delete', onClick: handleConfirm, primary: true },
]);
</script>

<style scoped>
.content {
  padding: 12px;
  text-align: left;
  line-height: 1.5;
}
strong {
  font-weight: 600;
}
</style>
