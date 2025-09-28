<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import { useVModel } from '@vueuse/core'

const {textarea} = useTextareaAutosize()

const props = defineProps<{
  class?: HTMLAttributes['class']
  defaultValue?: string | number
  modelValue?: string | number
  autosize?: boolean
}>()

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void
}>()

const modelValue = useVModel(props, 'modelValue', emits, {
  passive: true,
  defaultValue: props.defaultValue,
})
</script>

<template>
  <textarea v-if="!props.autosize" v-model="modelValue" :class="cn('flex min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-md ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50', props.class)" />
   <div v-else :class="cn('flex w-full rounded-md border border-input', props.class)">
    <textarea ref="textarea" v-model="modelValue" class="w-full min-h-20 resize-none w-full rounded-md border-input bg-background px-3 py-2 text-md ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
   </div>
</template>
