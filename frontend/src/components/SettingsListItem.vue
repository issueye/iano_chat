<template>
    <div
        class="group flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-all duration-200 cursor-pointer"
        :class="[
            isSelected
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/30'
        ]"
        role="listitem"
        tabindex="0"
        @click="$emit('click')"
        @keydown.enter="$emit('click')"
        @keydown.space.prevent="$emit('click')"
    >
        <div class="flex items-center gap-3 flex-1 min-w-0">
            <div
                class="w-9 h-9 rounded-lg bg-secondary flex items-center justify-center flex-shrink-0 transition-colors group-hover:bg-secondary/80"
            >
                <component
                    :is="icon"
                    class="h-4 w-4 text-muted-foreground"
                    :class="{ 'text-primary': isSelected }"
                />
            </div>
            <div class="min-w-0 flex-1">
                <p
                    class="font-medium text-sm truncate transition-colors"
                    :class="[
                        isSelected
                            ? 'text-primary'
                            : 'text-foreground group-hover:text-foreground'
                    ]"
                >
                    {{ title }}
                </p>
                <p
                    class="text-xs text-muted-foreground truncate"
                    :title="subtitle"
                >
                    {{ subtitle }}
                </p>
            </div>
        </div>
        <div class="flex items-center gap-2 flex-shrink-0">
            <Badge
                v-if="badge"
                variant="secondary"
                class="text-xs font-normal"
            >
                {{ badge }}
            </Badge>
            <Button
                variant="ghost"
                size="sm"
                class="text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity"
                @click.stop="$emit('configure')"
            >
                <Settings2 class="h-4 w-4 mr-1" />
                配置
            </Button>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Settings2 } from "lucide-vue-next";

const props = defineProps({
    title: { type: String, required: true },
    subtitle: { type: String, default: "" },
    icon: { type: Object, required: true },
    badge: { type: String, default: "" },
    isSelected: { type: Boolean, default: false },
});

defineEmits(["click", "configure"]);
</script>
