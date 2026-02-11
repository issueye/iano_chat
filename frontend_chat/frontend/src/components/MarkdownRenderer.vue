<template>
  <XMarkdown :markdown="content" :themes="{ light: 'andromeeda', dark: 'material-theme-darker' }" :codeXProps="{ enableCodeLineNumber: true }" />
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
