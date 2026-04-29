import { defineStore } from "pinia";
import { ref } from "vue";

export const useTaskStore = defineStore("task", () => {
  const visible = ref<boolean>(false);
  const toggle = () => (visible.value = !visible.value);
  return {
    // STATE
    visible,

    // ACTIONS
    toggle,
  };
});
