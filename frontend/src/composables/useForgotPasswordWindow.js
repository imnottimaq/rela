import { ref } from "vue";

export const forgotPasswordVisible = ref(false);

export const showForgotPasswordWindow = () => {
    forgotPasswordVisible.value = true;
    if (typeof window !== "undefined") {
        try {
            window.dispatchEvent(
                new CustomEvent("rela:focus-window", { detail: "rela-window-forgotpassword" })
            );
        } catch (_) {
            // no-op
        }
    }
};

export const hideForgotPasswordWindow = () => {
    forgotPasswordVisible.value = false;
};

export function useForgotPasswordWindow() {
    return {
        forgotPasswordVisible,
        showForgotPasswordWindow,
        hideForgotPasswordWindow,
    };
}