import { computed, ref } from 'vue';
import { planEvent } from '../lib/api/resources/events';
import { toDisplayError } from '../lib/api/errors';
import { isPositiveQuantity } from '../lib/quantity';
import { es } from '../lib/i18n/es';
import type { ApiError, EventDemand, EventPlanningResponse } from '../types/api';

export interface DemandRow {
  key: string;
  productId: string;
  quantity: string;
}

export type EventMode = 'stored' | 'inline';

let demandSeq = 0;

export function newDemandRow(): DemandRow {
  demandSeq += 1;
  return { key: `dem-${demandSeq}`, productId: '', quantity: '' };
}

export function useEventPlanning() {
  const mode = ref<EventMode>('stored');
  const eventId = ref('EVT001');
  const rows = ref<DemandRow[]>([newDemandRow()]);
  const pending = ref(false);
  const error = ref<ApiError | Error | null>(null);
  const fieldError = ref('');
  const result = ref<EventPlanningResponse | null>(null);

  const canSubmit = computed(() => !pending.value);

  function addRow() {
    rows.value = [...rows.value, newDemandRow()];
  }

  function removeRow(key: string) {
    rows.value = rows.value.filter((row) => row.key !== key);
    if (rows.value.length === 0) {
      rows.value = [newDemandRow()];
    }
  }

  function updateRow(next: DemandRow) {
    rows.value = rows.value.map((row) => (row.key === next.key ? next : row));
  }

  function filledDemands(): EventDemand[] | null {
    const filled = rows.value.filter((row) => row.productId.trim() !== '' || row.quantity.trim() !== '');
    if (filled.length === 0) return null;
    for (const row of filled) {
      if (!row.productId.trim() || !isPositiveQuantity(row.quantity)) {
        return null;
      }
    }
    return filled.map((row) => ({
      product_id: row.productId.trim(),
      quantity: row.quantity.trim(),
    }));
  }

  async function submit() {
    fieldError.value = '';
    error.value = null;
    result.value = null;

    if (mode.value === 'stored') {
      if (!eventId.value.trim()) {
        fieldError.value = es.forms.required;
        return;
      }
      pending.value = true;
      try {
        result.value = await planEvent({ event_id: eventId.value.trim() });
      } catch (err) {
        error.value = toDisplayError(err);
      } finally {
        pending.value = false;
      }
      return;
    }

    const demands = filledDemands();
    if (!demands) {
      const anyPartial = rows.value.some((row) => row.productId.trim() !== '' || row.quantity.trim() !== '');
      fieldError.value = anyPartial ? es.events.invalidQuantity : es.events.emptyDemands;
      return;
    }

    pending.value = true;
    try {
      result.value = await planEvent({ demands });
    } catch (err) {
      error.value = toDisplayError(err);
    } finally {
      pending.value = false;
    }
  }

  function dismissError() {
    error.value = null;
  }

  return {
    mode,
    eventId,
    rows,
    pending,
    error,
    fieldError,
    result,
    canSubmit,
    addRow,
    removeRow,
    updateRow,
    submit,
    dismissError,
  };
}
