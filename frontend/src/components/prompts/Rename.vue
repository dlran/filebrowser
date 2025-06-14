<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.rename") }}</h2>
    </div>

    <div class="card-content">
      <p>
        {{ $t("prompts.renameMessage") }} <code>{{ oldName() }}</code
        >:
      </p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="generateExifName()"
      >
        exif
      </button>
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat"
        type="submit"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')"
      >
        {{ $t("buttons.rename") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import url from "@/utils/url";
import { files as api, tools as tools } from "@/api";

export default {
  name: "rename",
  data: function () {
    return {
      name: "",
    };
  },
  created() {
    this.name = this.oldName();
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, [
      "req",
      "selected",
      "selectedCount",
      "isListing",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    cancel: function () {
      this.closeHovers();
    },
    oldName: function () {
      if (!this.isListing) {
        return this.req.name;
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return;
      }

      return this.req.items[this.selected[0]].name;
    },
    async generateExifName() {
      try {
        const res = await tools.extractExif({
          from: this.req.items[this.selected[0]].url,
        });
        const uc = res.UserComment.split('_');
        const extension = this.name.split('.').pop();
        const randStr = this.generateRandomString(4)
        const name = `${uc[3]}_${uc[0]}_${uc[1]}_${randStr}.${extension}`;
        this.name = name;
      } catch (e) {
        this.$showError(e);
      }
    },
    generateRandomString(length) {
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
      let result = '';
      const charactersLength = characters.length;

      for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
      }

      return result;
    },
    submit: async function () {
      let oldLink = "";
      let newLink = "";

      if (!this.isListing) {
        oldLink = this.req.url;
      } else {
        oldLink = this.req.items[this.selected[0]].url;
      }

      newLink =
        url.removeLastDir(oldLink) + "/" + encodeURIComponent(this.name);

      window.sessionStorage.setItem("modified", "true");
      try {
        await api.move([{ from: oldLink, to: newLink }]);
        if (!this.isListing) {
          this.$router.push({ path: newLink });
          return;
        }

        this.reload = true;
      } catch (e) {
        this.$showError(e);
      }

      this.closeHovers();
    },
  },
};
</script>
