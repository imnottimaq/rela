<template>
  <WindowComponent
    title="Edit Workspace"
    v-model:visible="modelVisible"
    :initial-size="{ width: 350, height: 200 }"
    :footer-buttons="footerButtons"
  >
    <div class="content">
      <p>Enter a new name for the workspace:</p>
      <input v-model="newName" type="text" class="input-field" ref="nameInput" />
      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </WindowComponent>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue';
import WindowComponent from './WindowComponent.vue';

const props = defineProps({
  workspace: { type: Object, required: true },
  visible: { type: Boolean, default: false },
});

const emit = defineEmits(['update:visible', 'save']);

const modelVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const newName = ref('');
const error = ref('');
const nameInput = ref(null);

watch(() => props.visible, (isVisible) => {
  if (isVisible) {
    newName.value = props.workspace.name;
    error.value = '';
    nextTick(() => {
      nameInput.value?.focus();
      nameInput.value?.select();
    });
  }
});

const handleSave = () => {
  const trimmedName = newName.value.trim();
  if (!trimmedName) {
    error.value = 'Workspace name cannot be empty.';
    return;
  }
  if (trimmedName === props.workspace.name) {
    modelVisible.value = false;
    return;
  }
  emit('save', trimmedName);
};

const handleCancel = () => {
  modelVisible.value = false;
};

const footerButtons = computed(() => [
  { label: 'Cancel', onClick: handleCancel },
  { label: 'Save', onClick: handleSave, primary: true },
]);
</script>

<style scoped>
.content {
  padding: 12px;
  text-align: left;
}
.input-field {
  width: 100%;
  padding: 8px;
  margin-top: 8px;
  box-sizing: border-box;
}
.error {
  color: #c00;
  margin-top: 8px;
}
</style>
