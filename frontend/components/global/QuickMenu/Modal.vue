<template>
    <BaseModal v-model="modal" :show-close-button="false">
        <div class="relative">
            <QuickMenuInput ref="inputBox" v-model="selectedAction" :actions="props.actions || []" @quickSelect="invokeAction"></QuickMenuInput>
            <ul v-if=false class="menu rounded-box w-full">
                <li v-for="(action, idx) in (actions || [])" :key="idx">
                    <button 
                        @click="invokeAction(action)" 
                        class="transition-colors w-full text-left rounded-btn p-3 hover:bg-neutral hover:text-white">
                        <b v-if="action.shortcut">{{action.shortcut}}.</b> 
                        
                        {{ action.text }}
                    </button>
                </li>
            </ul>
            <span class="text-base-300">Use number keys to quick select.</span>
        </div>
    </BaseModal>
</template>

<script setup lang="ts">
    import type { QuickMenuAction, QuickMenuInput } from "./Input.vue"
    
    const props = defineProps({
        modelValue: {
            type: Boolean,
            required: true,
        },
        actions: {
            type: Array as PropType<QuickMenuAction[]>,
            required: false,
        },
    });

    const modal = useVModel(props, "modelValue");
    const selectedAction = ref<QuickMenuAction>();
    
    const inputBox = ref<QuickMenuInput>({ focused: false, revealActions: () => {} });
    
    const onModalOpen = useTimeoutFn(() => {
        inputBox.value.focused = true;
    }, 50).start

    const onModalClose = () => {
        selectedAction.value = undefined
        inputBox.value.focused = false
    } 

    watch(modal, () => (modal.value ? onModalOpen : onModalClose)())


    onStartTyping(() => {
        inputBox.value.focused = true
    })
    
    function invokeAction(action: QuickMenuAction) {
        modal.value = false;
        useTimeoutFn(action.action, 100).start();
    }

    watch(selectedAction, (action) => {
        if (action)
            invokeAction(action)
    })
</script>