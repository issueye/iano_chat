<template>
  <div class="markdown-content">
    <XMarkdown :markdown="content" :themes="{ light: 'andromeeda', dark: 'material-theme-darker' }" :codeXProps="{ enableCodeLineNumber: true }" />
  </div>
</template>

<script setup>
import { XMarkdown } from 'vue-element-plus-x';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});
</script>

<style>

</style>
