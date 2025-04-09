import { toast as internalToast } from "vue-sonner";

// triggering too many toasts at once can cause the toaster to not render properly https://github.com/xiaoluoboding/vue-sonner/issues/98

const wrapToast = <T extends (...args: any[]) => any>(fn: T): ((...args: Parameters<T>) => Promise<ReturnType<T>>) => {
  return (...args: Parameters<T>) =>
    new Promise(resolve => {
      setTimeout(() => resolve(fn(...args)), 0);
    });
};

const toast = (...args: Parameters<typeof internalToast>) => internalToast(...args);

toast.success = wrapToast(internalToast.success);
toast.info = wrapToast(internalToast.info);
toast.warning = wrapToast(internalToast.warning);
toast.error = wrapToast(internalToast.error);
toast.message = wrapToast(internalToast.message);

export { toast };
