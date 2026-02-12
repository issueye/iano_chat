<script setup>
import { ref, onMounted, onUnmounted } from "vue"
import { Check, X } from "lucide-vue-next"

const props = defineProps({
  message: {
    type: String,
    required: true,
  },
  type: {
    type: String,
    default: "success",
    validator: (value) => ["success", "error", "info"].includes(value),
  },
  duration: {
    type: Number,
    default: 2000,
  },
})

const emit = defineEmits(["close"])

const visible = ref(false)

onMounted(() => {
  visible.value = true
  if (props.duration > 0) {
    setTimeout(() => {
      close()
    }, props.duration)
  }
})

function close() {
  visible.value = false
  setTimeout(() => {
    emit("close")
  }, 200)
}

const iconMap = {
  success: Check,
  error: X,
  info: "div",
}

const colorMap = {
  success: "bg-green-500",
  error: "bg-red-500",
  info: "bg-blue-500",
}
</script>

<template>
  <Transition name="toast">
    <div
      v-if="visible"
      :class="[
        'fixed top-4 right-4 z-[100] flex items-center gap-2 px-4 py-3 rounded-lg shadow-lg text-white text-sm',
        colorMap[type],
      ]"
    >
      <component
        :is="iconMap[type]"
        :class="[
          'w-4 h-4',
          type === 'success' && 'text-white',
          type === 'error' && 'text-white',
        ]"
      />
      <span>{{ message }}</span>
    </div>
  </Transition>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
