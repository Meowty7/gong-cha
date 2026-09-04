import { post } from '../client';
import type { EventPlanningRequest, EventPlanningResponse } from '../../../types/api';

export async function planEvent(
  body: EventPlanningRequest,
  signal?: AbortSignal
): Promise<EventPlanningResponse> {
  const result = await post<EventPlanningResponse>('/api/v1/calculate/event', body, { signal });
  if (!result) {
    throw new Error('Empty event planning response');
  }
  return result;
}
