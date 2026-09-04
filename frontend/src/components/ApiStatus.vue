<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { get } from '../lib/api/client';
import type { HealthResponse } from '../types/api';
import { es } from '../lib/i18n/es';

const status = ref<'loading' | 'ok' | 'error'>('loading');
const version = ref<string>('');
const errorMessage = ref<string>('');

// Get API URL from environment (must be accessed in script section)
const apiUrl = computed(() => {
  if (typeof import.meta !== 'undefined' && import.meta.env) {
    return import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
  }
  return 'http://localhost:8080';
});

async function checkHealth() {
  status.value = 'loading';
  errorMessage.value = '';
  
  try {
    const response = await get<HealthResponse>('/health/ready', { timeoutMs: 5000 });
    
    if (response?.status === 'ok') {
      status.value = 'ok';
      version.value = response.version || '';
    } else {
      status.value = 'error';
      errorMessage.value = es.apiStatus.disconnected;
    }
  } catch (error) {
    status.value = 'error';
    if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = es.apiStatus.error;
    }
  }
}

onMounted(() => {
  checkHealth();
});
</script>

<template>
  <div class="api-status" :class="`status-${status}`">
    <div class="status-indicator">
      <span 
        class="status-dot" 
        :class="`dot-${status}`"
        :aria-label="status === 'ok' ? es.apiStatus.connected : es.apiStatus.disconnected"
      />
      <span class="status-text">
        <template v-if="status === 'loading'">
          {{ es.apiStatus.checking }}
        </template>
        <template v-else-if="status === 'ok'">
          {{ es.apiStatus.connected }}
          <span v-if="version" class="status-version">({{ version }})</span>
        </template>
        <template v-else>
          {{ es.apiStatus.error }}
        </template>
      </span>
    </div>
    
    <div v-if="status === 'error'" class="status-error">
      <p class="error-message">{{ errorMessage }}</p>
      <p class="error-hint">
        {{ es.apiStatus.checkConnection }}
        <code>{{ apiUrl }}</code>
      </p>
      <button 
        type="button" 
        class="btn-retry"
        @click="checkHealth"
      >
        {{ es.apiStatus.retry }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.api-status {
  padding: 1rem;
  border-radius: 0.375rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-dot {
  display: inline-block;
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  transition: background-color 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.dot-loading {
  background-color: var(--color-text-subtle);
  animation: pulse 1.5s ease-in-out infinite;
}

.dot-ok {
  background-color: var(--color-success);
}

.dot-error {
  background-color: var(--color-error);
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@media (prefers-reduced-motion: reduce) {
  .dot-loading {
    animation: none;
    opacity: 0.7;
  }
}

.status-text {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text);
}

.status-version {
  font-weight: 400;
  color: var(--color-text-muted);
  margin-left: 0.5rem;
}

.status-error {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

.error-message {
  font-size: 0.875rem;
  color: var(--color-error);
  margin-bottom: 0.5rem;
}

.error-hint {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin-bottom: 1rem;
}

.error-hint code {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  background: var(--color-bg-warm);
  border-radius: 0.25rem;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.75rem;
}

.btn-retry {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary);
  background: transparent;
  border: 1px solid var(--color-primary);
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-retry:hover {
  background: var(--color-primary);
  color: white;
}

.btn-retry:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
