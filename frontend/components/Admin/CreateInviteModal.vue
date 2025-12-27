<template>
  <BaseModal :dialog-id="DialogID.CreateInvite" title="Create Invite">
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="createInvite()">
      <div class="flex flex-col gap-4">
        <div class="flex flex-col gap-2">
          <Label for="invite-role">Role</Label>
          <Select :model-value="form.role" @update:model-value="v => (form.role = String(v))">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="owner">owner</SelectItem>
              <SelectItem value="admin">admin</SelectItem>
              <SelectItem value="editor">editor</SelectItem>
              <SelectItem value="viewer">viewer</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="flex flex-col gap-2">
          <Label for="invite-expires">Expires</Label>
          <div :class="form.no_expiry ? 'opacity-50 pointer-events-none' : ''">
            <Popover>
              <PopoverTrigger as-child>
                <Button
                  variant="outline"
                  class="w-full justify-start text-left font-normal"
                  :class="!form.expires_at && 'text-muted-foreground'"
                >
                  <CalendarIcon class="mr-2 size-4" />
                  {{ formattedExpires ? formattedExpires : "Pick a date" }}
                </Button>
              </PopoverTrigger>

              <PopoverContent class="w-auto p-0">
                <Calendar :model-value="localExpires as any" @update:model-value="val => (localExpires = val)" />
              </PopoverContent>
            </Popover>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <Label for="invite-max-uses">Max Uses</Label>
          <Input
            id="invite-max-uses"
            v-model.number="form.max_uses"
            type="number"
            min="1"
            :disabled="form.unlimited_uses"
          />
        </div>

        <div class="mt-2 flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <Checkbox id="no-expiry" v-model="form.no_expiry" />
            <Label for="no-expiry" class="select-none">No expiry</Label>
          </div>

          <div class="flex items-center gap-2">
            <Checkbox id="unlimited-uses" v-model="form.unlimited_uses" />
            <Label for="unlimited-uses" class="select-none">Unlimited uses</Label>
          </div>
        </div>
      </div>

      <div class="mt-4 flex flex-row-reverse gap-2">
        <Button type="submit">Generate Invite</Button>
        <Button variant="outline" type="button" @click="cancel">Cancel</Button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { reactive, computed, ref, watch } from "vue";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { useDialogHotkey, useDialog } from "~/components/ui/dialog-provider";
  import { Button } from "~/components/ui/button";
  import { Input } from "~/components/ui/input";
  import { Calendar } from "~/components/ui/calendar";
  import { Popover, PopoverTrigger, PopoverContent } from "~/components/ui/popover";
  import { Checkbox } from "~/components/ui/checkbox";
  import { Calendar as CalendarIcon } from "lucide-vue-next";
  import { format } from "date-fns";
  import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "~/components/ui/select";
  import { Label } from "~/components/ui/label";
  import { api, type Invite } from "~/mock/collections";
  import { toast } from "~/components/ui/sonner";
  import { DialogID } from "~/components/ui/dialog-provider/utils";

  useDialogHotkey(DialogID.CreateInvite, { code: "Digit9" });

  const { closeDialog } = useDialog();

  const form = reactive({
    role: "viewer",
    expires_at: undefined as unknown,
    max_uses: 1,
    no_expiry: false,
    unlimited_uses: false,
  });

  // local date ref to satisfy Calendar's expected Date type
  const localExpires = ref<Date | undefined>(form.expires_at as Date | undefined);

  watch(
    () => form.expires_at,
    v => {
      localExpires.value = (v as Date) || undefined;
    }
  );

  watch(localExpires, v => {
    form.expires_at = v as unknown;
  });

  const formattedExpires = computed(() => {
    const v = form.expires_at as Date | string | undefined | null;
    if (!v) return null;
    if (v instanceof Date) return format(v, "PPP");
    try {
      const d = new Date(String(v));
      if (!isNaN(d.getTime())) return format(d, "PPP");
    } catch (e) {
      // fallthrough
    }
    return String(v);
  });

  function reset() {
    form.role = "viewer";
    form.expires_at = undefined;
    form.max_uses = 1;
    form.no_expiry = false;
    form.unlimited_uses = false;
  }

  function cancel() {
    reset();
    closeDialog(DialogID.CreateInvite);
  }

  function generateCode(length = 6) {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let out = "";
    for (let i = 0; i < length; i++) out += chars.charAt(Math.floor(Math.random() * chars.length));
    return out;
  }

  function createInvite() {
    const collectionId = api.getCollections()[0]?.id ?? "";
    const invite: Partial<Invite> = {
      id: generateCode(6),
      collectionId,
      role: form.role as Invite["role"],
      created_at: new Date().toISOString(),
      expires_at: form.no_expiry
        ? undefined
        : form.expires_at
          ? form.expires_at instanceof Date
            ? form.expires_at.toISOString()
            : String(form.expires_at)
          : undefined,
      max_uses: form.unlimited_uses ? undefined : form.max_uses || undefined,
      uses: 0,
    };

    api.addInvite(invite);
    toast.success("Invite created");
    reset();
    closeDialog(DialogID.CreateInvite, true);
  }
</script>
