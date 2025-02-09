<script setup lang="ts">
import { toast } from 'vue-sonner';
import { Button, Group } from '~/components/ui/button';
import { Shortcut } from '~/components/ui/shortcut';
</script>

<template>
  <h1 class="text-3xl">Testing</h1>
  <p>This is a page for testing during the transition to shadcn.</p>
  <Button
    @click="
      toast.success('This is a test', {
        duration: 100000,
      })
    "
    >Create Toast</Button
  >
  <Group>
    <Button variant="outline">Button 1</Button>
    <Button variant="destructive">Button 2</Button>
    <Button>Button 3</Button>
  </Group>
  <LabelSelector />
  <span class="flex items-center gap-1 text-sm">
    Use
    <Shortcut size="sm" :keys="['Shift']" /> +
    <Shortcut size="sm" :keys="['Enter']" /> to create and add another.
  </span>
</template>
