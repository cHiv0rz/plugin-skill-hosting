<script setup lang="ts">
defineProps<{
  name: string
  type: string
  required?: boolean
}>()
</script>

<template>
  <div class="param-row">
    <div class="param-head">
      <code class="param-name">{{ name }}</code>
      <span class="param-type">{{ type }}</span>
      <span v-if="required" class="param-required">required</span>
    </div>
    <div class="param-desc">
      <slot />
    </div>
  </div>
</template>
