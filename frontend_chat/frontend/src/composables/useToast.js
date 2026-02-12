import { ref, createApp, h } from "vue"
import Toast from "@/components/ui/toast/Toast.vue"

const toasts = ref([])
let toastId = 0

function show(options) {
  const id = ++toastId
  const toast = {
    id,
    ...options,
  }
  toasts.value.push(toast)
  return id
}

function remove(id) {
  const index = toasts.value.findIndex((t) => t.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

export function useToast() {
  return {
    success: (message, duration = 2000) => {
      const container = document.createElement("div")
      document.body.appendChild(container)

      const app = createApp({
        render() {
          return h(Toast, {
            message,
            type: "success",
            duration,
            onClose: () => {
              app.unmount()
              container.remove()
            },
          })
        },
      })

      app.mount(container)
    },
    error: (message, duration = 3000) => {
      const container = document.createElement("div")
      document.body.appendChild(container)

      const app = createApp({
        render() {
          return h(Toast, {
            message,
            type: "error",
            duration,
            onClose: () => {
              app.unmount()
              container.remove()
            },
          })
        },
      })

      app.mount(container)
    },
    info: (message, duration = 2000) => {
      const container = document.createElement("div")
      document.body.appendChild(container)

      const app = createApp({
        render() {
          return h(Toast, {
            message,
            type: "info",
            duration,
            onClose: () => {
              app.unmount()
              container.remove()
            },
          })
        },
      })

      app.mount(container)
    },
  }
}
