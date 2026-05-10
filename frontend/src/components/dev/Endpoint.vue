<script setup lang="ts">
import { computed, useSlots } from 'vue'

const props = withDefaults(defineProps<{
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  path: string
  summary: string
  auth?: boolean
}>(), { auth: true })

const slots = useSlots()

const slug = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')

const elementId = computed(
  () => `ep-${props.method.toLowerCase()}-${slug(props.path)}`,
)
</script>

<template>
  <div class="endpoint" :id="elementId">
    <div class="endpoint-head">
      <span class="endpoint-method" :class="`method-${method.toLowerCase()}`">{{ method }}</span>
      <code class="endpoint-path">{{ path }}</code>
      <span
        class="endpoint-auth"
        :class="{ 'endpoint-auth--public': !auth }"
      >{{ auth ? 'auth required' : 'public' }}</span>
    </div>
    <p class="endpoint-summary">{{ summary }}</p>
    <div v-if="slots.params" class="endpoint-block">
      <h4>Path parameters</h4>
      <slot name="params" />
    </div>
    <div v-if="slots.request" class="endpoint-block">
      <h4>Request body</h4>
      <slot name="request" />
    </div>
    <div v-if="slots.response" class="endpoint-block">
      <h4>Response</h4>
      <slot name="response" />
    </div>
    <div v-if="slots.errors" class="endpoint-block">
      <h4>Errors</h4>
      <slot name="errors" />
    </div>
    <div v-if="slots.example" class="endpoint-block">
      <h4>Example</h4>
      <slot name="example" />
    </div>
    <div v-if="slots.notes" class="endpoint-block endpoint-notes">
      <slot name="notes" />
    </div>
  </div>
</template>
