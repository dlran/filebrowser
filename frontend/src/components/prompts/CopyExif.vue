<template>
  <div class="card floating">
    <div class="card-content">
      <p>
        Copy Exif
        {{ selectedCount }}
      </p>
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
import { files as api } from "@/api";
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
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {
      buttons.loading("copyExif");

      window.sessionStorage.setItem("modified", "true");
      try {
        // if (!this.isListing) {
        //   buttons.success("copyExif");

        //   this.currentPrompt?.confirm();
        //   this.closeHovers();
        //   return;
        // }

        this.closeHovers();

        if (this.selectedCount !== 2) {
          return;
        }

        const selectedItem = [];
        for (const index of this.selected) {
          selectedItem.push(this.req.items[index]);
        }

        await api.copyExif({
          from: selectedItem[0].url,
          to: selectedItem[1].url,
        });
        buttons.success("copyExif");
        this.reload = true;
      } catch (e) {
        buttons.done("copyExif");
        this.$showError(e);
        if (this.isListing) this.reload = true;
      }
    },
  },
};
</script>
