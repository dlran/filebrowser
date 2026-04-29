<template>
  <div class="card floating">
    <div class="card-title">
      <h2>Encryption ({{ selectedCount }} Selected)</h2>
    </div>

    <div class="card-content">
      <p>Password</p>
      <input class="input input--block" type="text" v-model.trim="password" />
      <p>Confirm Password</p>
      <input
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="confirmPwd"
      />
      <p>
        <label>
          <input type="checkbox" name="mode" v-model="mode" />
          {{ mode ? "encrypt" : "decrypt" }}
        </label>
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
        @click="submit"
        class="button button--flat button--red"
        :aria-label="`Encrypt`"
        :title="`Encrypt`"
        tabindex="1"
      >
        Encrypt
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
  name: "Encrypt",
  inject: ["$showError"],
  data: function () {
    return {
      password: "",
      mode: false,
      confirmPwd: "",
    };
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
        if (this.password !== this.confirmPwd) {
          this.$showError("Two inputs are inequality", false);
          return;
        }

        await api.taskCall({
          src: selectedItem,
          password: this.password,
          mode: this.mode,
          taskName: "encrypt_task",
          // taskName: "test_task",
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
