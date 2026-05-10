<script setup lang="ts">
defineProps<{ message?: string | null }>()
</script>

<template>
  <div v-if="message" class="error">{{ message }}</div>
</template>
