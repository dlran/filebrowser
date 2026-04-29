<template>
  <div v-if="visible" class="task-list" v-bind:class="{ closed: !open }">
    <div class="card floating">
      <div class="card-title">
        <h2>Task list</h2>
        <action icon="lock" label="" show="encrypt"> </action>
        <button
          class="action"
          @click="getTaskList"
          aria-label="Refresh"
          title="Refresh"
        >
          <i class="material-icons">{{ "refresh" }}</i>
        </button>
        <button
          class="action"
          @click="toggle"
          aria-label="Toggle file upload list"
          title="Toggle file upload list"
        >
          <i class="material-icons">{{
            open ? "keyboard_arrow_down" : "keyboard_arrow_up"
          }}</i>
        </button>
      </div>

      <div class="card-content">
        <template v-if="data.length">
          <div class="file" v-for="item in data" :key="item">
            <div class="task-item">
              <div class="task-item-bd" :title="item.id">
                <span>{{ item.type }} {{ item.result }}</span>
                &nbsp;
                <span>{{ item.retention }}</span>
                <div class="task-item-subt">
                  <template v-if="item.state === 'archived'">{{
                    item.last_failed_at
                  }}</template>
                  <template v-else>{{ item.completed_at }}</template>
                </div>
              </div>
              <i class="material-icons" v-if="item.state === 'active'">sync</i>
              <i class="material-icons" v-if="item.state === 'archived'"
                >block</i
              >
              <i class="material-icons" v-if="item.state === 'completed'"
                >done</i
              >
              <button
                class="action"
                @click="deleteTask(item.id)"
                aria-label="Delete"
                title="Delete"
              >
                <i class="material-icons">delete_outline</i>
              </button>
            </div>
            <div class="file-progress">
              <div
                v-bind:style="{
                  width: '100%',
                }"
              ></div>
            </div>
          </div>
        </template>
        <template v-else>
          <h2 class="message">
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ t("files.lonely") }}</span>
          </h2>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch, ref, inject } from "vue";
import { storeToRefs } from "pinia";
import { tasks as api } from "@/api";
import Action from "@/components/header/Action.vue";
import { useTaskStore } from "@/stores/task";
import { useI18n } from "vue-i18n";
const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;
const taskStore = useTaskStore();

const { visible } = storeToRefs(taskStore);
const open = ref<boolean>(true);
const data = ref<string[]>([]);

watch(visible, (newValue) => {
  newValue && getTaskList();
});

const toggle = () => {
  open.value = !open.value;
};

const getTaskList = async () => {
  try {
    const res = await api.taskList();
    data.value = res.tasks;
  } catch (e) {
    $showError(e);
  }
};
const deleteTask = async (id) => {
  try {
    await api.taskDelete(id);
    getTaskList();
  } catch (e) {
    $showError(e);
  }
};
</script>

<style scoped>
.task-list .card.floating {
  left: auto;
  top: auto;
  margin: 0;
  right: 0;
  bottom: 0;
  transform: none;
  z-index: 3;
}
@media (max-width: 450px) {
  .task-list .card.floating {
    max-width: 100%;
    width: 100%;
  }
}
.task-list .card-content {
  overflow-y: auto;
  max-height: 60vh;
}
.task-list.closed .card-content {
  display: none;
  padding: 0em 1em 1em 1em;
}
.task-list .card .card-title {
  align-items: center;
  justify-content: center;
  font-size: 0.8em;
  padding: 1em 1em 0em;
}
.task-list.closed .card-title {
  font-size: 0.7em;
  padding: 0.5em 1em;
}
.task-list .file {
  margin-bottom: 8px;
}

.sub-text {
  min-width: 19ch;
  width: auto;
  text-align: left;
}
.task-item {
  display: flex;
  align-items: center;
}
.task-item > .task-item-bd {
  flex: 1 1 auto;
}
.task-item .task-item-subt {
  font-size: 14px;
  color: #999;
}
.task-item .action > i {
  padding: 0.2em;
}
</style>
