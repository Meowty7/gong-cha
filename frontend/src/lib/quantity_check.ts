import { isPositiveQuantity, isNonNegativeQuantity } from './quantity';

const cases: Array<[string, boolean, boolean]> = [
  ['1', true, true],
  ['0.5', true, true],
  ['10.25', true, true],
  ['0', false, true],
  ['0.0', false, true],
  ['00', false, true],
  ['', false, false],
  ['-1', false, false],
  ['abc', false, false],
  ['1e2', false, false],
];

for (const [value, wantPositive, wantNonNeg] of cases) {
  const gotPositive = isPositiveQuantity(value);
  const gotNonNeg = isNonNegativeQuantity(value);
  if (gotPositive !== wantPositive || gotNonNeg !== wantNonNeg) {
    throw new Error(
      `quantity check failed for ${JSON.stringify(value)}: positive ${gotPositive}≠${wantPositive}, nonnegative ${gotNonNeg}≠${wantNonNeg}`
    );
  }
}

console.log('quantity_check ok');
