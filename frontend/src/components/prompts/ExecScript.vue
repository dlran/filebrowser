<template>
  <div class="card floating">
    <div class="card-content">
      <p>Execute ({{ selectedCount }} Selected)</p>
    </div>

    <div class="card-action">
      <button
        @click="closeHovers"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        tabindex="2"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat button--red"
        :aria-label="`ExecScript`"
        :title="`ExecScript`"
        tabindex="1"
      >
        Execute
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { tasks as api } from "@/api";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "executeScript",
  inject: ["$showError"],
  data: function () {
    return {};
  },
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {
      try {
        const selectedItem = [];
        for (const index of this.selected) {
          const item = this.req.items[index];
          if (!item.isDir) {
            selectedItem.push(item.path);
          }
        }

        if (!selectedItem.length) {
          this.$showError("no file selected", false);
          return;
        }

        await api.taskCall({
          src: selectedItem,
          taskName: "file:exec_scripts",
        });
        // this.reload = true;
        this.closeHovers();
      } catch (e) {
        this.$showError(e);
        // if (this.isListing) this.reload = true;
      }
    },
  },
};
</script>
