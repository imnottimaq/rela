<script setup>
import "7.css/dist/7.css";

defineProps({
  title: {
    type: String,
    default: "A glass window frame",
  },
  buttons: {
    type: Array,
    default: () => [],
  },
});
</script>

<template>
    <div class="window glass active" style="max-width: 100%">
      <div class="title-bar">
        <div class="title-bar-text">{{ title }}</div>

        <div class="title-bar-controls">
          <button
            v-for="(buttonElement, index) in buttons"
            :key="index"
            :aria-label="buttonElement.label"
            @click="buttonElement.onClick"
          ></button>
        </div>
      </div>

      <div class="window-body has-space">
        <slot></slot>
      </div>
    </div>
</template>
