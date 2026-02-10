<template>
    <Teleport to="body">
        <Transition
            enter-active-class="transition-all duration-200 ease-out"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition-all duration-200 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
        >
            <div
                v-if="isOpen"
                class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
                @click="dialog?.closeModal"
            >
                <Transition
                    enter-active-class="transition-all duration-200 ease-out"
                    enter-from-class="opacity-0 scale-95 translate-x-[-50%] translate-y-[-48%]"
                    enter-to-class="opacity-100 scale-100 translate-x-[-50%] translate-y-[-50%]"
                    leave-active-class="transition-all duration-200 ease-in"
                    leave-from-class="opacity-100 scale-100 translate-x-[-50%] translate-y-[-50%]"
                    leave-to-class="opacity-0 scale-95 translate-x-[-50%] translate-y-[-48%]"
                >
                    <div
                        v-if="isOpen"
                        :class="[
                            'fixed left-1/2 top-1/2 z-50 grid w-full max-w-lg -translate-x-1/2 -translate-y-1/2 border bg-background shadow-lg rounded-lg',
                            customClass,
                        ]"
                        @click.stop
                    >
                        <slot />
                    </div>
                </Transition>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup>
import { inject, computed } from 'vue';

const props = defineProps({
    class: {
        type: String,
        default: '',
    },
});

const dialog = inject('dialog');

const isOpen = computed(() => dialog?.isOpen?.value ?? false);
const customClass = computed(() => props.class);


</script>
