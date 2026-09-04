<script setup lang="ts">
const props = defineProps<{
  modelValue: string;
  id: string;
  label: string;
  error?: string;
  disabled?: boolean;
  required?: boolean;
  unit?: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();
</script>

<template>
  <div class="field">
    <label class="field__label" :for="id">{{ label }}</label>
    <div class="field__row">
      <input
        :id="id"
        class="input field__control tabular-nums"
        type="text"
        inputmode="decimal"
        autocomplete="off"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        :aria-invalid="error ? true : undefined"
        :aria-describedby="error ? `${id}-error` : undefined"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <span v-if="unit" class="field__unit">{{ unit }}</span>
    </div>
    <p v-if="error" :id="`${id}-error`" class="field__error">{{ error }}</p>
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.field__label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text);
}

.field__row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.field__control {
  min-height: 2.75rem;
}

.field__unit {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  white-space: nowrap;
}

</style>
