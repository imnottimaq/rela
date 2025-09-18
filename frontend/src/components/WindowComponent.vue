<script setup>
import "7.css/dist/7.css";
import {
  computed,
  defineComponent,
  h,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from "vue";

const layerManagerKey = "__relaWindowLayerManager__";

const getLayerManager = () => {
  const target =
    typeof window !== "undefined"
      ? window
      : typeof globalThis !== "undefined"
      ? globalThis
      : {};

  if (!target[layerManagerKey]) {
    target[layerManagerKey] = {
      maxLayer: 0,
      topWindowId: null,
      counter: 0,
    };
  }

  return target[layerManagerKey];
};

const layerManager = getLayerManager();

const createInstanceId = () => {
  layerManager.counter += 1;
  return layerManager.counter;
};

const normalizeLayer = (value) => {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return null;
  return Math.max(1, Math.round(numeric));
};

const syncGlobalLayer = (value) => {
  const normalized = normalizeLayer(value);
  if (normalized === null) return null;
  if (normalized > layerManager.maxLayer) {
    layerManager.maxLayer = normalized;
  }
  return normalized;
};

const allocateLayer = () => {
  layerManager.maxLayer += 1;
  return layerManager.maxLayer;
};

const resolveNumber = (value, fallback) =>
  typeof value === "number" && !Number.isNaN(value) ? value : fallback;

const props = defineProps({
  title: {
    type: String,
    default: "A glass window frame",
  },
  buttons: {
    type: Array,
    default: () => [],
  },
  menu: {
    type: Array,
    default: () => [],
  },
  initialPosition: {
    type: Object,
    default: () => ({ x: 80, y: 80 }),
  },
  initialSize: {
    type: Object,
    default: () => ({ width: 600, height: 420 }),
  },
  visible: {
    type: Boolean,
    default: false,
  },
  storageKey: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["update:visible"]);

const fallbackPosition = { x: 80, y: 80 };
const fallbackSize = { width: 600, height: 420 };
const minWidth = 240;
const minHeight = 160;

const position = reactive({
  x: resolveNumber(props.initialPosition?.x, fallbackPosition.x),
  y: resolveNumber(props.initialPosition?.y, fallbackPosition.y),
});

const size = reactive({
  width: resolveNumber(props.initialSize?.width, fallbackSize.width),
  height: resolveNumber(props.initialSize?.height, fallbackSize.height),
});

const lastInitialPosition = reactive({ ...position });
const lastInitialSize = reactive({ ...size });

const storageKey = computed(() =>
  typeof props.storageKey === "string" ? props.storageKey.trim() : ""
);

const internalVisible = ref(props.visible);

const instanceId = createInstanceId();
const layer = ref(allocateLayer());

const menuItems = computed(() =>
  Array.isArray(props.menu) ? props.menu : []
);

const windowStyle = computed(() => ({
  transform: `translate3d(${position.x}px, ${position.y}px, 0)`,
  width: `${size.width}px`,
  height: `${size.height}px`,
  maxWidth: "100vw",
  maxHeight: "100vh",
  zIndex: layer.value,
}));

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

const clampDimension = (value, min, max) => {
  if (typeof max === "number") {
    if (max < min) return max;
    return Math.min(Math.max(value, min), max);
  }
  return Math.max(value, min);
};

const clampToViewport = () => {
  if (typeof window === "undefined") return;

  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;

  if (viewportWidth < minWidth) {
    size.width = viewportWidth;
    position.x = 0;
  } else {
    size.width = clampDimension(size.width, minWidth, viewportWidth);
  }

  if (viewportHeight < minHeight) {
    size.height = viewportHeight;
    position.y = 0;
  } else {
    size.height = clampDimension(size.height, minHeight, viewportHeight);
  }

  const maxX = Math.max(0, viewportWidth - size.width);
  const maxY = Math.max(0, viewportHeight - size.height);

  position.x = Math.min(Math.max(position.x, 0), maxX);
  position.y = Math.min(Math.max(position.y, 0), maxY);
};

const parseStoredState = () => {
  if (typeof window === "undefined" || !storageKey.value) return null;

  try {
    const raw = window.localStorage.getItem(storageKey.value);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch (error) {
    return null;
  }
};

const persistState = () => {
  if (typeof window === "undefined" || !storageKey.value) return;

  const stateToPersist = {
    position: { x: position.x, y: position.y },
    size: { width: size.width, height: size.height },
    visible: internalVisible.value,
    layer: layer.value,
  };

  try {
    window.localStorage.setItem(
      storageKey.value,
      JSON.stringify(stateToPersist)
    );
  } catch (error) {
    // Ignore storage errors (quota, access issues)
  }
};

const applyStoredState = () => {
  const storedState = parseStoredState();
  if (!storedState) return;

  const storedPosition = storedState.position ?? {};
  const storedSize = storedState.size ?? {};

  if (typeof storedPosition.x === "number") {
    position.x = storedPosition.x;
  }

  if (typeof storedPosition.y === "number") {
    position.y = storedPosition.y;
  }

  if (typeof storedSize.width === "number") {
    size.width = clampDimension(storedSize.width, minWidth);
  }

  if (typeof storedSize.height === "number") {
    size.height = clampDimension(storedSize.height, minHeight);
  }

  if (typeof storedState.visible === "boolean") {
    internalVisible.value = storedState.visible;
  }

  const storedLayer = syncGlobalLayer(storedState.layer);
  if (storedLayer !== null) {
    layer.value = storedLayer;
  } else {
    syncGlobalLayer(layer.value);
  }

  clampToViewport();
};

const handleWindowResize = () => {
  clampToViewport();
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

const bringToFront = () => {
  if (
    layerManager.topWindowId === instanceId &&
    layer.value === layerManager.maxLayer
  ) {
    return;
  }

  layerManager.topWindowId = instanceId;
  layer.value = allocateLayer();
  persistState();
};

const handleWindowPointerDown = () => {
  bringToFront();
};

const isPrimaryPointer = (event) => {
  if (event.pointerType === "mouse" && event.button !== 0) return false;
  return event.isPrimary !== false;
};

const blurEventTarget = (event) => {
  const target = event?.currentTarget;
  if (target && typeof target.blur === "function") {
    target.blur();
  }
};

const startDrag = (event) => {
  if (isResizing.value || !isPrimaryPointer(event)) return;
  isDragging.value = true;
  dragOffset.x = event.clientX - position.x;
  dragOffset.y = event.clientY - position.y;
  bringToFront();
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
  bringToFront();
  attachListeners();
};

const handlePointerMove = (event) => {
  if (isDragging.value) {
    position.x = event.clientX - dragOffset.x;
    position.y = event.clientY - dragOffset.y;
    clampToViewport();
    return;
  }

  if (isResizing.value) {
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

    clampToViewport();
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

watch(
  () => props.initialPosition,
  (nextPosition) => {
    if (!nextPosition) return;

    let shouldClamp = false;

    if (Object.prototype.hasOwnProperty.call(nextPosition, "x")) {
      const nextX = resolveNumber(nextPosition.x, lastInitialPosition.x);
      if (!Object.is(nextX, lastInitialPosition.x)) {
        lastInitialPosition.x = nextX;
        position.x = nextX;
        shouldClamp = true;
      }
    }

    if (Object.prototype.hasOwnProperty.call(nextPosition, "y")) {
      const nextY = resolveNumber(nextPosition.y, lastInitialPosition.y);
      if (!Object.is(nextY, lastInitialPosition.y)) {
        lastInitialPosition.y = nextY;
        position.y = nextY;
        shouldClamp = true;
      }
    }

    if (shouldClamp) {
      clampToViewport();
    }
  },
  { deep: true }
);

watch(
  () => props.initialSize,
  (nextSize) => {
    if (!nextSize) return;

    let shouldClamp = false;

    if (Object.prototype.hasOwnProperty.call(nextSize, "width")) {
      const nextWidth = resolveNumber(nextSize.width, lastInitialSize.width);
      if (!Object.is(nextWidth, lastInitialSize.width)) {
        lastInitialSize.width = nextWidth;
        size.width = nextWidth;
        shouldClamp = true;
      }
    }

    if (Object.prototype.hasOwnProperty.call(nextSize, "height")) {
      const nextHeight = resolveNumber(nextSize.height, lastInitialSize.height);
      if (!Object.is(nextHeight, lastInitialSize.height)) {
        lastInitialSize.height = nextHeight;
        size.height = nextHeight;
        shouldClamp = true;
      }
    }

    if (shouldClamp) {
      clampToViewport();
    }
  },
  { deep: true }
);

watch(
  () => props.visible,
  (nextVisible) => {
    if (nextVisible !== internalVisible.value) {
      internalVisible.value = nextVisible;
    }
  }
);

watch(
  () => internalVisible.value,
  (nextVisible, previousVisible) => {
    if (nextVisible !== previousVisible) {
      if (nextVisible && !previousVisible) {
        bringToFront();
      }
      emit("update:visible", nextVisible);
      persistState();
    }
  }
);

watch(
  () => [position.x, position.y, size.width, size.height, layer.value],
  () => {
    persistState();
  }
);

onMounted(() => {
  applyStoredState();

  if (typeof window !== "undefined") {
    clampToViewport();
    window.addEventListener("resize", handleWindowResize);
  }
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", handleWindowResize);
  }
  detachListeners();
  persistState();
});

const menuItemClasses = (item) => {
  const classes = [];
  if (item?.class) classes.push(item.class);
  if (item?.divider) classes.push("has-divider");
  if (item?.disabled) classes.push("is-disabled");
  return classes;
};

const renderShortcut = (shortcut) => {
  if (!shortcut) return null;
  return h("span", shortcut);
};

const MenuList = defineComponent({
  name: "WindowMenuList",
  props: {
    items: {
      type: Array,
      default: () => [],
    },
  },
  setup(listProps) {
    const handleLinkClick = (event, item) => {
      if (item.disabled) {
        event.preventDefault();
        return;
      }

      if (item.onClick) {
        item.onClick(event);
        if (!item.href) {
          event.preventDefault();
        }
      }

      blurEventTarget(event);
    };

    const renderInteractiveContent = (item) => {
      const children = [item.label];
      const shortcut = renderShortcut(item.shortcut);
      if (shortcut) children.push(shortcut);

      const useButton = item.type === "button";

      if (useButton) {
        return h(
          "button",
          {
            type: "button",
            disabled: item.disabled,
            onClick:
              item.disabled || !item.onClick
                ? undefined
                : (event) => {
                    item.onClick(event);
                    blurEventTarget(event);
                  },
          },
          children
        );
      }

      return h(
        "a",
        {
          href: item.href ?? undefined,
          "aria-disabled": item.disabled || undefined,
          tabindex: item.disabled ? -1 : undefined,
          onClick: (event) => handleLinkClick(event, item),
        },
        children
      );
    };

    const renderMenuItem = (item, index) => {
      const key = item.id ?? `${item.label ?? "separator"}-${index}`;

      if (item.type === "separator") {
        return h("li", {
          key,
          role: "separator",
          class: ["menu-separator", item.class || ""].filter(Boolean),
        });
      }

      if (
        item.type === "submenu" &&
        Array.isArray(item.items) &&
        item.items.length
      ) {
        return h(
          "li",
          {
            key,
            role: "menuitem",
            tabindex: 0,
            "aria-haspopup": "true",
            class: menuItemClasses(item),
          },
          [item.label, h(MenuList, { items: item.items })]
        );
      }

      return h(
        "li",
        {
          key,
          role: "menuitem",
          tabindex: -1,
          class: menuItemClasses(item),
        },
        [renderInteractiveContent(item)]
      );
    };

    return () =>
      h(
        "ul",
        { role: "menu" },
        (listProps.items ?? []).map((item, index) =>
          renderMenuItem(item, index)
        )
      );
  },
});

</script>

<template>
  <div
    class="window glass active draggable-window"
    :style="windowStyle"
    v-if="internalVisible"
    @pointerdown.capture="handleWindowPointerDown"
  >
    <div class="title-bar" @pointerdown.prevent.stop="startDrag">
      <div class="title-bar-text">{{ title }}</div>

      <div class="title-bar-controls" v-if="buttons.length">
        <button
          v-for="(buttonElement, index) in buttons"
          :key="index"
          :aria-label="buttonElement.label"
          @click="buttonElement.onClick"
        ></button>
      </div>
    </div>

    <div class="window-body">
      <ul
        v-if="menuItems.length"
        role="menubar"
        class="menubar can-hover window-menubar"
      >
        <li
          v-for="(menuItem, menuIndex) in menuItems"
          :key="menuItem.id ?? `menu-${menuIndex}`"
          role="menuitem"
          tabindex="0"
          aria-haspopup="true"
          :class="menuItemClasses(menuItem)"
        >
          {{ menuItem.label }}
          <MenuList :items="menuItem.items ?? []" />
        </li>
      </ul>

      <div class="window-body-content">
        <slot></slot>
      </div>
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
.window {
  display: flex;
  flex-direction: column;
}

.draggable-window {
  position: absolute;
  top: 0;
  left: 0;
  display: flex;
  flex-direction: column;
  min-width: min(240px, 100vw);
  min-height: min(160px, 100vh);
}

.draggable-window .title-bar {
  cursor: move;
  user-select: none;
  touch-action: none;
}

.window-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.window-menubar {
  margin: 0;
  user-select: none;
}

.window-menubar > li {
  position: relative;
}

.window-body-content {
  flex: 1;
  overflow: auto;
  padding: var(--w7-w-space);
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
</style>
