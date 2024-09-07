export const messages: Object = () => {
  const messages = {};
  const modules = import.meta.glob("~//locales/**.json", { eager: true });
  for (const path in modules) {
    const key = path.slice(9, -5);
    messages[key] = modules[path];
  }
  return messages;
};
