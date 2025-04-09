import { toast as internalToast } from "vue-sonner";

// triggering too many toasts at once can cause the toaster to not render properly https://github.com/xiaoluoboding/vue-sonner/issues/98

export const toast = (...args: Parameters<typeof internalToast>): Promise<ReturnType<typeof internalToast>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast(...args)), 0);
  });
};

toast.success = (
  ...args: Parameters<typeof internalToast.success>
): Promise<ReturnType<typeof internalToast.success>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast.success(...args)), 0);
  });
};

toast.info = (...args: Parameters<typeof internalToast.info>): Promise<ReturnType<typeof internalToast.info>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast.info(...args)), 0);
  });
};

toast.warning = (
  ...args: Parameters<typeof internalToast.warning>
): Promise<ReturnType<typeof internalToast.warning>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast.warning(...args)), 0);
  });
};

toast.error = (...args: Parameters<typeof internalToast.error>): Promise<ReturnType<typeof internalToast.error>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast.error(...args)), 0);
  });
};

toast.message = (
  ...args: Parameters<typeof internalToast.message>
): Promise<ReturnType<typeof internalToast.message>> => {
  return new Promise(resolve => {
    setTimeout(() => resolve(internalToast.message(...args)), 0);
  });
};
