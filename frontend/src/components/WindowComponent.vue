<script setup>
import "7.css/dist/7.css";
import { computed, onBeforeUnmount, reactive, ref } from "vue";

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

const position = reactive({ x: 80, y: 80 });
const size = reactive({ width: 600, height: 420 });

const minWidth = 240;
const minHeight = 160;

const isDragging = ref(false);
const dragOffset = reactive({ x: 0, y: 0 });

const isResizing = ref(false);
const resizeDirection = ref("");
const resizeStart = reactive({
  x: 0,
  y: 0,
  width: 0,
  height: 0,
  left: 0,
  top: 0,
});

const resizeHandles = ["n", "ne", "e", "se", "s", "sw", "w", "nw"];

const windowStyle = computed(() => ({
  transform: `translate3d(${position.x}px, ${position.y}px, 0)`,
  width: `${size.width}px`,
  height: `${size.height}px`,
}));

let listenersAttached = false;
const previousUserSelect = ref(null);

const setBodySelection = (value) => {
  if (typeof document === "undefined") return;

  if (value === null) {
    if (previousUserSelect.value !== null) {
      document.body.style.userSelect = previousUserSelect.value;
      previousUserSelect.value = null;
    }
    return;
  }

  if (previousUserSelect.value === null) {
    previousUserSelect.value = document.body.style.userSelect || "";
  }

  document.body.style.userSelect = value;
};

const attachListeners = () => {
  if (listenersAttached || typeof document === "undefined") return;
  document.addEventListener("pointermove", handlePointerMove);
  document.addEventListener("pointerup", stopInteractions);
  document.addEventListener("pointercancel", stopInteractions);
  listenersAttached = true;
  setBodySelection("none");
};

const detachListeners = () => {
  if (!listenersAttached || typeof document === "undefined") return;
  document.removeEventListener("pointermove", handlePointerMove);
  document.removeEventListener("pointerup", stopInteractions);
  document.removeEventListener("pointercancel", stopInteractions);
  listenersAttached = false;
  setBodySelection(null);
};

const isPrimaryPointer = (event) => {
  if (event.pointerType === "mouse" && event.button !== 0) return false;
  return event.isPrimary !== false;
};

const startDrag = (event) => {
  if (isResizing.value || !isPrimaryPointer(event)) return;
  isDragging.value = true;
  dragOffset.x = event.clientX - position.x;
  dragOffset.y = event.clientY - position.y;
  attachListeners();
};

const startResize = (direction, event) => {
  if (isDragging.value || !isPrimaryPointer(event)) return;
  isResizing.value = true;
  resizeDirection.value = direction;
  resizeStart.x = event.clientX;
  resizeStart.y = event.clientY;
  resizeStart.width = size.width;
  resizeStart.height = size.height;
  resizeStart.left = position.x;
  resizeStart.top = position.y;
  attachListeners();
};

const handlePointerMove = (event) => {
  if (isDragging.value) {
    position.x = event.clientX - dragOffset.x;
    position.y = event.clientY - dragOffset.y;
  } else if (isResizing.value) {
    const dx = event.clientX - resizeStart.x;
    const dy = event.clientY - resizeStart.y;

    if (resizeDirection.value.includes("e")) {
      size.width = Math.max(minWidth, resizeStart.width + dx);
    }

    if (resizeDirection.value.includes("s")) {
      size.height = Math.max(minHeight, resizeStart.height + dy);
    }

    if (resizeDirection.value.includes("w")) {
      const newWidth = Math.max(minWidth, resizeStart.width - dx);
      position.x = resizeStart.left + (resizeStart.width - newWidth);
      size.width = newWidth;
    }

    if (resizeDirection.value.includes("n")) {
      const newHeight = Math.max(minHeight, resizeStart.height - dy);
      position.y = resizeStart.top + (resizeStart.height - newHeight);
      size.height = newHeight;
    }
  }
};

function stopInteractions(event) {
  if (!isDragging.value && !isResizing.value) return;
  if (event) event.preventDefault();
  isDragging.value = false;
  isResizing.value = false;
  resizeDirection.value = "";
  detachListeners();
}

onBeforeUnmount(() => {
  detachListeners();
});
</script>

<template>
    <div class="window glass active draggable-window" :style="windowStyle">
      <div class="title-bar" @pointerdown.prevent.stop="startDrag">
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

      <div
        v-for="handle in resizeHandles"
        :key="handle"
        :class="['resize-handle', `resize-${handle}`]"
        @pointerdown.prevent.stop="startResize(handle, $event)"
      ></div>
    </div>
</template>

<style scoped>
.draggable-window {
  position: absolute;
  top: 0;
  left: 0;
  display: flex;
  flex-direction: column;
}

.draggable-window .title-bar {
  cursor: move;
  user-select: none;
  touch-action: none;
}

.resize-handle {
  position: absolute;
  z-index: 10;
  background: transparent;
  touch-action: none;
}

.resize-n {
  top: -4px;
  left: 0;
  right: 0;
  height: 8px;
  cursor: ns-resize;
}

.resize-s {
  bottom: -4px;
  left: 0;
  right: 0;
  height: 8px;
  cursor: ns-resize;
}

.resize-e {
  top: 0;
  bottom: 0;
  right: -4px;
  width: 8px;
  cursor: ew-resize;
}

.resize-w {
  top: 0;
  bottom: 0;
  left: -4px;
  width: 8px;
  cursor: ew-resize;
}

.resize-ne {
  top: -4px;
  right: -4px;
  width: 12px;
  height: 12px;
  cursor: nesw-resize;
}

.resize-se {
  bottom: -4px;
  right: -4px;
  width: 12px;
  height: 12px;
  cursor: nwse-resize;
}

.resize-sw {
  bottom: -4px;
  left: -4px;
  width: 12px;
  height: 12px;
  cursor: nesw-resize;
}

.resize-nw {
  top: -4px;
  left: -4px;
  width: 12px;
  height: 12px;
  cursor: nwse-resize;
}

.window-body {
  flex: 1;
  overflow: auto;
}
</style>
