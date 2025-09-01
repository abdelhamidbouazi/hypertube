import { createRequire } from 'module';const require = createRequire(import.meta.url);
import {
  ElementRef
} from "./chunk-BWLR3LOX.js";

// node_modules/@angular/cdk/fesm2022/element.mjs
function coerceNumberProperty(value, fallbackValue = 0) {
  if (_isNumberValue(value)) {
    return Number(value);
  }
  return arguments.length === 2 ? fallbackValue : 0;
}
function _isNumberValue(value) {
  return !isNaN(parseFloat(value)) && !isNaN(Number(value));
}
function coerceElement(elementOrRef) {
  return elementOrRef instanceof ElementRef ? elementOrRef.nativeElement : elementOrRef;
}

// node_modules/@angular/cdk/fesm2022/array.mjs
function coerceArray(value) {
  return Array.isArray(value) ? value : [value];
}

export {
  coerceNumberProperty,
  coerceElement,
  coerceArray
};
//# sourceMappingURL=chunk-7K4ZL6HA.js.map
