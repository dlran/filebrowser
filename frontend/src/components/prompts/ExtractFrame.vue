<template>
  <div class="card floating">
    <div class="card-title">
      <h2>Extract Video Frame</h2>
    </div>
    <div class="card-content">
      <p>FPS</p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.number="fps"
      />
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
        id="focus-prompt"
        @click="submit"
        class="button button--flat button--red"
        :aria-label="`CopyExif`"
        :title="`CopyExif`"
        tabindex="1"
      >
        COPY EXIF
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { tools as api } from "@/api";
import buttons from "@/utils/buttons";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "copyExif",
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
      "currentPrompt",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
  },
  data: function () {
    return {
      fps: 1,
    };
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {
      window.sessionStorage.setItem("modified", "true");
      try {
        if (this.selectedCount !== 1) {
          return;
        }

        await api.extractFrame({
          from: this.req.items[this.selected[0]].url,
          fps: this.fps,
        });
        this.reload = true;
      } catch (e) {
        this.$showError(e);
      }
      this.closeHovers();
    },
  },
};
</script>
