import { ref } from "vue";

export const createBoardVisible = ref(false);
export const createBoardWorkspaceId = ref(null);

export const showCreateBoardWindow = (workspaceId) => {
    if (!workspaceId) {
        console.error("showCreateBoardWindow requires a workspaceId");
        return;
    }
    createBoardWorkspaceId.value = workspaceId;
    createBoardVisible.value = true;
    if (typeof window !== "undefined") {
        try {
            window.dispatchEvent(
                new CustomEvent("rela:focus-window", { detail: "rela-window-create-board" })
            );
        } catch (_) {
            // no-op
        }
    }
};

export const hideCreateBoardWindow = () => {
    createBoardVisible.value = false;
    createBoardWorkspaceId.value = null;
};

export function useCreateBoardWindow() {
    return {
        createBoardVisible,
        createBoardWorkspaceId,
        showCreateBoardWindow,
        hideCreateBoardWindow,
    };
}
