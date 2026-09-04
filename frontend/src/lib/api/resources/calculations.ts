import { post } from '../client';
import type {
  DirectCalculationRequest,
  DirectCalculationResponse,
  InverseCalculationRequest,
  InverseCalculationResponse,
} from '../../../types/api';

export async function calculateDirect(
  body: DirectCalculationRequest,
  signal?: AbortSignal
): Promise<DirectCalculationResponse> {
  const result = await post<DirectCalculationResponse>('/api/v1/calculate/direct', body, { signal });
  if (!result) {
    throw new Error('Empty direct calculation response');
  }
  return result;
}

export async function calculateInverse(
  body: InverseCalculationRequest,
  signal?: AbortSignal
): Promise<InverseCalculationResponse> {
  const result = await post<InverseCalculationResponse>('/api/v1/calculate/inverse', body, { signal });
  if (!result) {
    throw new Error('Empty inverse calculation response');
  }
  return result;
}
